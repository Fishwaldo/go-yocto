package appstream

import (
	"encoding/xml"
	"io"
	"sort"
	"strings"

	//	"github.com/Fishwaldo/go-yocto/source"
	"github.com/Fishwaldo/go-yocto/utils"
	"github.com/Masterminds/semver/v3"
)

type AsSummary struct {
	Lang string `xml:"lang,attr"`
	Summary string `xml:",chardata"`
}

type AsDescription struct {
	Lang string `xml:"lang,attr"`
	Description string `xml:",chardata"`
}

type AsReleases struct {
	Version string `xml:"version,attr"`
	Date string `xml:"date,attr"`
}

type AppStream struct {
	Component xml.Name `xml:"component"`
	Name string `xml:"name"`
	Summary []AsSummary `xml:"summary"`
	Description []AsDescription `xml:"description>p"`
	Releases []AsReleases `xml:"releases>release"`
}

type ASProject struct {
	Backend string
	BackendID string
	Project AppStream
}

type AppStreamParser struct {
	ready bool
}

func NewParser() *AppStreamParser {
	return &AppStreamParser{}
}

func (k *AppStreamParser) GetName() string {
	return "AppStream"
}

func (k *AppStreamParser) Init() error {
	utils.Logger.Trace("Initializing AppStream Parser")
	k.ready = true
	return nil
}

func (k *AppStreamParser) Ready() bool {
	return k.ready
}

func (k *AppStreamParser) Parse(data io.Reader) (metadata map[string]interface{}, err error) {
	/* now parse the appstream */
	var as AppStream
	var raw []byte
	raw, err = io.ReadAll(data)
	metadata = make(map[string]interface{})
	if err != nil {
		utils.Logger.Error("Failed to read appstream", utils.Logger.Args("error", err))
	} else { 
		err = xml.Unmarshal(raw, &as)
		if err != nil {
			utils.Logger.Error("Failed to unmarshal appstream", utils.Logger.Args("error", err))
		} else {
			for _, v := range as.Summary {
				if v.Lang == "" {
					metadata["summary"] = strings.TrimSpace(v.Summary)
					break;
				}
			}
			for _, v := range as.Description {
				if v.Lang == "" {
					metadata["description"] = strings.TrimSpace(v.Description)
					break;
				}
			}
			var releases []*semver.Version
			for _, v := range as.Releases {
				ver, err := semver.NewVersion(v.Version)
				if err != nil {
					utils.Logger.Error("Failed to parse version", utils.Logger.Args("version", v.Version, "error", err))
					continue
				}
				if ver.Prerelease() != "" {
					utils.Logger.Trace("Skipping Prerelease", utils.Logger.Args("version", v.Version))
					continue
				}
				releases = append(releases, ver)
			}
			if len(releases) > 0 {
				sort.Sort(semver.Collection(releases))
				metadata["version"] = releases[len(releases)-1].Original()
			}
		}
	}
	return metadata, nil
}