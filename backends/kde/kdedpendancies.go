package kde

import (
	"path"

	"github.com/Fishwaldo/go-yocto/utils"
)

var inheritmap map[string]string = make(map[string]string)
var dependmap map[string]string = make(map[string]string)

func init() {
	inheritmap = map[string]string {
		"frameworks/extra-cmake-modules": "cmake_plasma",
		"frameworks/kauth": "kauth",
		"frameworks/kcmutils": "kcmutils",
		"frameworks/kconfig": "kconfig",
		"frameworks/kcoreaddons": "kcoreaddons",
		"frameworks/kdoctools": "kdoctools",
		"frameworks/ki18n" : "ki18n",
	}
	dependmap = map[string]string {
	}

}



func GetInherits(pr Project, depmap map[string][]string) (inherits []string, err error) {
	utils.Logger.Trace("Getting Inherits", utils.Logger.Args("project", pr.Name))
	if _, ok := depmap[pr.ProjectPath]; ok {
		for _, v := range depmap[pr.ProjectPath] {
			if i, ok := inheritmap[v]; ok {
				inherits = append(inherits, i)
			}
		}
	}
	/* add our default inherit */
	inherits = append(inherits, "reuse_license_checksums")

	return inherits, nil
}

func GetDepends(pr Project, depmap map[string][]string) (depends []string, err error) {
	utils.Logger.Trace("Getting Depends", utils.Logger.Args("project", pr.Name))
	if _, ok := depmap[pr.ProjectPath]; ok {
		for _, v := range depmap[pr.ProjectPath] {
			if _, ok := inheritmap[v]; !ok {
				depends = append(depends, path.Base(v))
			}
		}
	}
	return depends, nil
}