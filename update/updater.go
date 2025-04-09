package update

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	semver "github.com/Masterminds/semver/v3"
	"github.com/mholt/archiver"
	"github.com/pkg/errors"
)

const (
	LatestReleaseGitHubEndpoint = "https://api.github.com/repos/jbc2212321/wails-react-3-demo/releases"
	GithubToken                 = ""
)

// Updater updates an app.
type Updater struct {
	// CurrentVersion is the current install version.
	CurrentVersion string
	// LatestReleaseGitHubEndpoint is the URL of the API to get latest release data.
	// For example, https://api.github.com/repos/jbc2212321/wails-react-3-demo/releases
	LatestReleaseGitHubEndpoint string
	// Client is the HTTP client to use to access the
	// API and download the assets.
	Client *http.Client
	// github token
	GitHubToken string
	// SelectAsset selects the Asset to install.
	SelectAsset SelectAssetFunc
	// DownloadBytesLimit is the maximum number of bytes to download.
	DownloadBytesLimit int64
	// GetExecutable is the function that gets the current
	// executable. If nil, os.Executable will be used.
	GetExecutable func() (string, error)
}

// SelectAssetFunc selects the Asset to install.
type SelectAssetFunc func(release Release, asset Asset) bool

// Update checks and installs the update.
// Returns nil, nil if no update is required.
func (u *Updater) Update() (*Release, error) {
	if u.DownloadBytesLimit == 0 {
		return nil, errors.New("must set DownloadBytesLimit")
	}
	if u.SelectAsset == nil {
		return nil, errors.New("missing SelectAsset func")
	}
	if u.GetExecutable == nil {
		u.GetExecutable = os.Executable
	}
	latest, err := u.getLatestRelease()
	if err != nil {
		return nil, err
	}
	hasUpdate := hasUpdate(u.CurrentVersion, latest.TagName)
	if !hasUpdate {
		return nil, nil
	}
	var selectedAsset *Asset
	for _, asset := range latest.Assets {
		if u.SelectAsset(*latest, asset) {
			selectedAsset = &asset
			break
		}
	}
	if selectedAsset == nil {
		return nil, errors.New("no asset selected, use SelectAssetFunc to select an asset")
	}
	fmt.Println(selectedAsset)
	err = u.downloadAndReplaceApp(*selectedAsset)
	if err != nil {
		return nil, errors.Wrap(err, "download update")
	}
	return latest, nil
}

// Restart spawns the current executable again, and terminates
// the running one.
func (u *Updater) Restart() error {
	time.Sleep(1 * time.Second)
	thisExecuable, err := os.Executable()
	if err != nil {
		return errors.Wrap(err, "get executable")
	}
	log.Println("restarting", thisExecuable)
	cmd := exec.Command(thisExecuable)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: false,
		Credential: &syscall.Credential{
			NoSetGroups: true,
		},
	}
	cmd.Dir = filepath.Dir(thisExecuable)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "XBAR_UPDATE_RESTART_COUNTER=1")
	cmd.Args = os.Args
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Start()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return errors.Wrapf(err, "starting new app failed: exit code %d", exitErr.ExitCode())
		}
		return errors.Wrap(err, "starting new app failed")
	}
	log.Println("waiting before terminating after update...")
	time.Sleep(1 * time.Second)
	log.Println("terminating after update.")
	os.Exit(0)
	return nil
}

// getLatestRelease gets the latest release.
func (u *Updater) getLatestRelease() (*Release, error) {
	method := "GET"
	req, err := http.NewRequest(method, u.LatestReleaseGitHubEndpoint, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Authorization", u.GitHubToken)

	resp, err := u.Client.Do(req)

	if err != nil {
		return nil, errors.Wrap(err, "get latest release")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to check for updates: got %s", resp.Status)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, u.DownloadBytesLimit))
	if err != nil {
		return nil, errors.Wrap(err, "read body")
	}
	var latestReleases []Release
	err = json.Unmarshal(b, &latestReleases)
	if err != nil {
		return nil, errors.Wrap(err, "marshal")
	}
	latestReleases[0].CreatedAt, err = time.Parse(time.RFC3339Nano, latestReleases[0].CreatedAtString)
	if err != nil {
		return nil, errors.Wrap(err, "time.Parse: created_at")
	}
	return &latestReleases[0], nil
}

// HasUpdate checks whether there's an update or not.
func (u *Updater) HasUpdate() (*Release, bool, error) {
	latest, err := u.getLatestRelease()
	fmt.Println(latest)
	if err != nil {
		return nil, false, err
	}
	hasUpdate := hasUpdate(u.CurrentVersion, latest.TagName)
	return latest, hasUpdate, nil
}

// hasUpdate compares the current and latest version strings to
// see if there is an update.
// Returns false if the versions match.
// Returns false is current is in front of latest.
// If semver checking fails, direct string comparison is used.
func hasUpdate(current, latest string) bool {
	semverValid := true
	currentV, err := semver.NewVersion(current)
	if err != nil {
		fmt.Println("CURRENT SEMVER NON VALID")
		semverValid = false
	}
	latestV, err := semver.NewVersion(latest)
	if err != nil {
		fmt.Println("LATEST SEMVER NON VALID")
		semverValid = false
	}
	if semverValid {
		if currentV.Equal(latestV) {
			return false // up-to-date
		}
		if currentV.GreaterThan(latestV) {
			return false // local version is higher
		}
	} else {
		// semver failed - just check tags
		if latest == current {
			return false
		}
	}
	return true
}

func (u *Updater) downloadAndReplaceApp(asset Asset) error {
	filename := path.Base(asset.BrowserDownloadURL)
	switch {
	case strings.HasSuffix(filename, ".zip"):
		// fine
	default:
		return errors.Errorf("file not supported: %s", filename)
	}

	method := "GET"
	req, err := http.NewRequest(method, asset.BrowserDownloadURL, nil)
	if err != nil {
		fmt.Println(err)
	}
	req.Header.Set("Authorization", u.GitHubToken)

	resp, err := u.Client.Do(req)
	if err != nil {
		return errors.Wrap(err, "download asset")
	}
	defer resp.Body.Close()
	fmt.Println(resp.Status)
	const defaultTempDir = ""
	f, err := os.CreateTemp(defaultTempDir, "*-"+filename)
	if err != nil {
		return errors.Wrap(err, "create temp file")
	}
	_, err = io.Copy(f, io.LimitReader(resp.Body, u.DownloadBytesLimit))
	if err != nil {
		f.Close()
		return errors.Wrap(err, "download asset")
	}
	f.Close()

	if u.GetExecutable == nil {
		u.GetExecutable = os.Executable
	}
	executable, err := u.GetExecutable()
	if err != nil {
		return errors.Wrap(err, "get executable")
	}
	fmt.Println("executable:", executable)
	appPath, err := appPathFromExecutable(executable)
	fmt.Println("appPath:", appPath)
	if err != nil {
		return errors.Wrap(err, "find app path")
	}
	appPathDir := filepath.Dir(appPath)
	appPreviousPath := appPath + ".previous"
	err = os.Rename(appPath, appPreviousPath)
	if err != nil {
		_, statErr := os.Stat(appPath)
		// not exist is ok, just ignore it
		if !os.IsNotExist(statErr) {
			return errors.Wrap(err, "rename existing app")
		}
	}

	fmt.Println("UPDATERRRRR", filename, f.Name(), appPath, appPathDir, appPreviousPath)
	err = archiver.Unarchive(f.Name(), appPathDir)
	if err != nil {
		return errors.Wrap(err, "unarchive")
	}
	err = os.RemoveAll(appPreviousPath)
	if err != nil {
		return errors.Wrap(err, "remove previous")
	}
	return nil
}

// Release is a GitHub release.
type Release struct {
	TagName         string    `json:"tag_name"`
	Assets          []Asset   `json:"assets"`
	Body            string    `json:"body"`
	CreatedAtString string    `json:"created_at"`
	CreatedAt       time.Time `json:"created_at_time"`
}

// Asset is a file within a Release on GitHub.
type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// appPathFromExecutable gets the .app path from the currently
// running executable.
func appPathFromExecutable(p string) (string, error) {
	return p, nil
	//if !strings.HasSuffix(p, "/Contents/MacOS/wails-react-3-demo") {
	//	return "", errors.New("executable not where it should be")
	//}
	//if !strings.HasSuffix(p, "wails-react-3-demo.app/Contents/MacOS/wails-react-3-demo") {
	//	return "", errors.New("executable not where it should be")
	//}
	//
	//return strings.TrimSuffix(p, "wails-react-3-demo"), nil
}

func AppUpdate() error {
	updater := &Updater{
		CurrentVersion:              version,
		LatestReleaseGitHubEndpoint: LatestReleaseGitHubEndpoint,
		Client:                      &http.Client{Timeout: 10 * time.Minute},
		GitHubToken:                 GithubToken,
		SelectAsset: func(release Release, asset Asset) bool {
			// look for the zip file
			return strings.Contains(asset.Name, fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)) && filepath.Ext(asset.Name) == ".zip"
		},
		DownloadBytesLimit: 52_428_800, // 50MB
	}

	latest, hasUpdate, err := updater.HasUpdate()

	if err != nil {
		fmt.Println(err)
	}
	if hasUpdate {
		fmt.Println("updating")
		fmt.Println(latest)
	}
	var selectedAsset *Asset
	for _, asset := range latest.Assets {
		if updater.SelectAsset(*latest, asset) {
			selectedAsset = &asset
			break
		}
	}
	if selectedAsset == nil {
		fmt.Println("no selected asset")
		return err
	}
	fmt.Println(selectedAsset)
	err = updater.downloadAndReplaceApp(*selectedAsset)
	if err != nil {
		return err
	}
	return err
}
