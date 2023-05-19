package kde

import (
	"bufio"
	"regexp"
	"net/http"
	"encoding/json"
	"io/ioutil"
	"errors"
	"strings"
	"crypto/sha256"
	"io"
	"fmt"

	"github.com/Fishwaldo/go-yocto/utils"
	"github.com/pterm/pterm"
)

type dirListing struct {
	Directory string
}

var files map[string]map[string]dirListing = make(map[string]map[string]dirListing)

func LoadDownloadLocationsCache() (error) {
	utils.Logger.Trace("Loading KDE Download Cache")
	cache, err := ioutil.ReadFile(utils.Config.BaseDir + "/kdedownload-cache.json")
	if err != nil {
		utils.Logger.Error("Failed to read download cache", utils.Logger.Args("error", err))
		return RefreshDownloadLocations()
	} else {
		err = json.Unmarshal(cache, &files)
		if err != nil {
			utils.Logger.Error("Failed to unmarshal cache", utils.Logger.Args("error", err))
			return err
		}
	}
	return nil
}

func RefreshDownloadLocations() (error) {
	file, err := http.Get("https://download.kde.org/ls-lR")
	if err != nil {
		return err
	}
	defer file.Body.Close()
	scanner := bufio.NewScanner(file.Body)
	var directory = regexp.MustCompile(`^(\./)?(.+):$`)
	var fn = regexp.MustCompile(`^[^dl].* ((.*)-(.*)\.tar\.(bz2|xz))$`)
	var curdir string
	for scanner.Scan() {
		if directory.MatchString(scanner.Text()) {
			curdir = strings.TrimPrefix( directory.FindStringSubmatch(scanner.Text())[2], "/srv/archives/ftp/")

		}
		if fn.MatchString(scanner.Text()) {
			f := fn.FindStringSubmatch(scanner.Text())
			if _, ok := files[f[2]]; !ok {
				files[f[2]] = make(map[string]dirListing)
			}
			files[f[2]][f[3]] = dirListing{Directory: curdir + "/" + f[1]}
		}
	}
	cache, err := json.Marshal(files)
	if err != nil {
		utils.Logger.Error("Failed to marshal metadata", utils.Logger.Args("error", err))
		return err
	}
	err = ioutil.WriteFile(utils.Config.BaseDir + "/kdedownload-cache.json", cache, 0644)
	if err != nil {
		utils.Logger.Error("Failed to write metadata", utils.Logger.Args("error", err))
		return err
	}
	return nil
}

func GetDownloadPath(source string, version string) (string, error) {
	if _, ok := files[source]; !ok {
		return "", errors.New("Source not found")
	}
	if _, ok := files[source][version]; !ok {
		return "", errors.New("Version not found")
	}
	return "https://download.kde.org/" + files[source][version].Directory, nil
}

func GetDownloadSHA(source string, version string) (string, error) {
	if _, ok := files[source]; !ok {
		return "", errors.New("Source not found")
	}
	if _, ok := files[source][version]; !ok {
		return "", errors.New("Version not found")
	}
	path, err := GetDownloadPath(source, version)
	if err != nil {
		return "", err
	}
	utils.Logger.Trace("Getting SHA", utils.Logger.Args("path", path))
	spinnerInfo, _ := pterm.DefaultSpinner.Start("Downloading Source for SHA Calculation")

	file, err := http.Get(path)
	if err != nil {
		utils.Logger.Warn("Failed to get SHA", utils.Logger.Args("path", path, "error", err))
		spinnerInfo.Fail("Failed to Download Source for SHA Calculation")
		return "", err
	}
	defer file.Body.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file.Body); err != nil {
		utils.Logger.Warn("Failed to calculate SHA", utils.Logger.Args("path", path, "error", err))
		spinnerInfo.Fail("Failed to Download Source for SHA Calculation")
		return "", err
	}
	spinnerInfo.Success("Downloaded Source for SHA Calculation")
	return fmt.Sprintf("%x", hash.Sum(nil)), nil

}