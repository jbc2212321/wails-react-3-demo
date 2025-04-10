package update

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"
)

const Version = "0.1.5"

func DoSelfUpdate() bool {
	v := semver.MustParse(Version)

	up, err := selfupdate.NewUpdater(selfupdate.Config{
		APIToken: "***",
	})
	if err != nil {
		fmt.Println(err)
	}

	latest, err := up.UpdateSelf(v, "jbc2212321/wails-react-3-demo")
	if err != nil {
		log.Println("Binary update failed:", err)
		return false
	}
	if latest.Version.Equals(v) {
		// latest version is the same as current version. It means current binary is up to date.
		log.Println("Current binary is the latest version", Version)
		return true
	} else {
		log.Println("Successfully updated to version", latest.Version)
		log.Println("Release note:\n", latest.ReleaseNotes)
		return true
	}
}

func doSelfUpdate() {
	v := semver.MustParse(version)
	latest, err := selfupdate.UpdateSelf(v, "myname/myrepo")
	if err != nil {
		log.Println("Binary update failed:", err)
		return
	}
	if latest.Version.Equals(v) {
		// latest version is the same as current version. It means current binary is up to date.
		log.Println("Current binary is the latest version", version)
	} else {
		log.Println("Successfully updated to version", latest.Version)
		log.Println("Release note:\n", latest.ReleaseNotes)
	}
}

func DoSelfUpdateMac() bool {
	latest, found, _ := selfupdate.DetectLatest("jbc2212321/wails-react-3-demo")
	if found {
		homeDir, _ := os.UserHomeDir()
		downloadPath := filepath.Join(homeDir, "Downloads", "RiftShare.zip")
		err := exec.Command("curl", "-L", latest.AssetURL, "-o", downloadPath).Run()
		if err != nil {
			log.Println("curl error:", err)
			return false
		}
		var appPath string
		cmdPath, err := os.Executable()
		appPath = strings.TrimSuffix(cmdPath, "RiftShare.app/Contents/MacOS/RiftShare")
		if err != nil {
			appPath = "/Applications/"
		}
		err = exec.Command("ditto", "-xk", downloadPath, appPath).Run()
		if err != nil {
			log.Println("ditto error:", err)
			return false
		}
		err = exec.Command("rm", downloadPath).Run()
		if err != nil {
			log.Println("removing error:", err)
			return false
		}
		return true
	} else {
		return false
	}
}

func CheckForUpdate() (bool, string) {
	latest, found, err := selfupdate.DetectLatest("achhabra2/riftshare")
	if err != nil {
		log.Println("Error occurred while detecting version:", err)
		return false, ""
	}

	v := semver.MustParse(Version)
	if !found || latest.Version.LTE(v) {
		log.Println("Current version is the latest")
		return false, ""
	}

	return true, latest.Version.String()
}
