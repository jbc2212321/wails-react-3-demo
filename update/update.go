package update

import (
	"bufio"
	"fmt"
	"github.com/blang/semver"
	"github.com/rhysd/go-github-selfupdate/selfupdate"

	"log"
	"os"
)

const version = "0.1.9"

func confirmAndSelfUpdate() {
	up, err := selfupdate.NewUpdater(selfupdate.Config{
		APIToken: "***",
	})
	if err != nil {
		fmt.Println(err)
	}
	latest, found, err := up.DetectLatest("jbc2212321/wails-react-3-demo")
	if err != nil {
		log.Println("Error occurred while detecting version:", err)
		return
	}

	v := semver.MustParse(version)
	if !found || latest.Version.LTE(v) {
		log.Println("Current version is the latest")
		return
	}

	fmt.Print("Do you want to update to", latest.Version, "? (y/n): ")
	input, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil || (input != "y\n" && input != "n\n") {
		log.Println("Invalid input")
		return
	}
	if input == "n\n" {
		return
	}

	exe, err := os.Executable()
	if err != nil {
		log.Println("Could not locate executable path")
		return
	}
	if err := selfupdate.UpdateTo(latest.AssetURL, exe); err != nil {
		log.Println("Error occurred while updating binary:", err)
		return
	}
	log.Println("Successfully updated to version", latest.Version)
}
