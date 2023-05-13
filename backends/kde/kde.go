package kde

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"encoding/base64"
	"encoding/xml"
	"encoding/json"

	"github.com/Fishwaldo/go-yocto/utils"
	"github.com/Fishwaldo/go-yocto/repo"

	"github.com/spf13/viper"
	"github.com/pterm/pterm"
	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v3"
)


type Deps struct {
	Dependencies []struct {
		On []string
		Require map[string]string
	} `yaml:"Dependencies"`
	Options interface{} `yaml:"Options"`
	Environment interface{} `yaml:"Environment"`
}

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

type Project struct {
	Name string
	ProjectPath string
	Repoactive bool
	Repopath string
	Identifier string
	Hasrepo bool
	Description string
	Source string
	Bugzilla struct {
		Product string
		Component string
		DNUlegacyproduct string `yaml:"__do_not_use-legacy-product"`
	}
	Topics []string
	MetaData struct {
		Branch string
		Dependencies Deps
		AppStream AppStream
	}
}

type Layer struct {
	MetaDataRepo repo.Repo
	br map[string]map[string]string
	pr map[string]Project
}

func init() {
	viper.SetDefault("kdeconfig.release", "@stable");
	viper.SetDefault("kdeconfig.defaultbranch", "master");
	viper.SetDefault("kdeconfig.kdegitlaburl", "https://invent.kde.org/api/v4")
}

func NewBackend() (l *Layer) {
	l = &Layer{}
	return l
}

func (l *Layer) GetName() string {
	return "kde-invent"
}

func (l *Layer) Init() {
	l.MetaDataRepo = repo.Repo{
		Url: "https://invent.kde.org/sysadmin/repo-metadata",
		Name: "kde-metadata",
	}
	utils.Logger.Trace("Checking metadata repo", utils.Logger.Args("repo", l.MetaDataRepo, "layer", l.GetName()))
	err := l.MetaDataRepo.CheckRepo()
	if (err != nil) {
		utils.Logger.Info("Cloning repo", utils.Logger.Args("repo", l.MetaDataRepo, "layer", l.GetName()))
		err := l.MetaDataRepo.CloneRepo()
		if (err != nil) {
			utils.Logger.Error("Failed to clone repo", utils.Logger.Args("error", err))
			os.Exit(-1)
		}
	}
	err = l.ParseMetadata()
	if (err != nil) {
		utils.Logger.Error("Failed to parse metadata", utils.Logger.Args("error", err))
		os.Exit(-1)
	}
}

func (l *Layer) GetDir() (dir string) {
	dir = utils.Config.BaseDir + "/" + l.MetaDataRepo.Name
	return dir
}


func (l *Layer) ParseMetadata() (err error) {
	l.pr = make(map[string]Project)
	utils.Logger.Info("Parsing metadata", utils.Logger.Args("layer", l.Name))
	brfile, err := ioutil.ReadFile(l.GetDir() + "/branch-rules.yml")
	if err != nil {
		utils.Logger.Error("Failed to read branch-rules.yaml", utils.Logger.Args("error", err))
		os.Exit(-1)
	}

	l.br = make(map[string]map[string]string)

	err = yaml.Unmarshal(brfile, &l.br)
	if err != nil {
		utils.Logger.Error("Failed to unmarshal branch-rules.yaml", utils.Logger.Args("error", err))
		os.Exit(-1)
	}
	/* make sure we have a valid release */
	if _, ok := l.br[utils.Config.KDEConfig.Release]; !ok {
		utils.Logger.Error("Invalid release", utils.Logger.Args("release", Config.KDEConfig.Release))
		os.Exit(-1)
	}

	gl, err := gitlab.NewClient(utils.Config.KDEConfig.AccessToken, gitlab.WithBaseURL(utils.Config.KDEConfig.KDEGitLabURL))
	if err != nil {
		utils.Logger.Error("Failed to create GitLab client", utils.Logger.Args("error", err))
		os.Exit(-1)
	}

	/* now parse the directory */
	files := findmetdata(l.GetDir())
	p, _ := pterm.DefaultProgressbar.WithTotal(len(files)).WithTitle("Parsing Metadata...").Start()
	for i := 0; i < p.Total; i++ {
		p.Increment()
		md, err := os.Open(files[i])
		if err != nil {
			utils.Logger.Error("Failed to open metadata file", utils.Logger.Args("error", err))
			continue
		}
		ymldec := yaml.NewDecoder(md)
		ymldec.KnownFields(true);
		var data Project
		err = ymldec.Decode(&data)
		if err != nil {
			utils.Logger.Error("Failed to decode metadata file", utils.Logger.Args("file", files[i], "error", err))
			continue
		}
		data.MetaData.Branch = utils.Config.KDEConfig.DefaultBranch
		/* find out which branch this is in... */
		for project, branch := range l.br[utils.Config.KDEConfig.Release] {
			ok, _ := filepath.Match(project, data.Repopath)
			if ok {
				data.MetaData.Branch = branch
				break
			}
		}

		/* get the .kde-ci.yml for dependencies */
		gf := &gitlab.GetFileOptions{
			Ref: gitlab.String(data.MetaData.Branch),
		}
		f, res, err := gl.RepositoryFiles.GetFile(data.Repopath, ".kde-ci.yml", gf)
		if err != nil {
			if res.StatusCode != 404 {
				utils.Logger.Error("Failed to get .kde-ci.yml", utils.Logger.Args("error", err))
			}
		} else { 
			/* now parse the .kde-ci.yml */
			var deps Deps
			content, err := base64.StdEncoding.DecodeString(f.Content)
			if err != nil {
				utils.Logger.Error("Failed to decode .kde-ci.yml", utils.Logger.Args("error", err))
			} else {
				ymldeps := yaml.NewDecoder(strings.NewReader(string(content)))
				ymldeps.KnownFields(false);
				err = ymldeps.Decode(&deps)
				if err != nil {
					utils.Logger.Error("Failed to unmarshal .kde-ci.yml", utils.Logger.Args("error", err, "project", data.Repopath))
				} else {
					data.MetaData.Dependencies = deps
				}
			}
		}
		/* now get appstream if it exists */ 
		f, res, err = gl.RepositoryFiles.GetFile(data.Repopath, "org.kde." + data.Identifier + ".appdata.xml", gf)
		if err != nil {
			if res.StatusCode != 404 {
				utils.Logger.Error("Failed to get appstream", utils.Logger.Args("error", err))
			}
		} else {
			/* now parse the appstream */
			content, err := base64.StdEncoding.DecodeString(f.Content)
			if err != nil {
				utils.Logger.Error("Failed to decode appstream", utils.Logger.Args("error", err))
			} else {
				var as AppStream
				err = xml.Unmarshal(content, &as)
				if err != nil {
					utils.Logger.Error("Failed to unmarshal appstream", utils.Logger.Args("error", err, "project", data.Repopath))
				} else {
					data.MetaData.AppStream = as
				}
			}
		}

		if _, ok := l.pr[data.Identifier]; ok {
			utils.Logger.Error("Duplicate identifier", utils.Logger.Args("identifier", data.Identifier))
			continue
		}
		data.Source = l.Name
		l.pr[data.Identifier] = data
	}
	utils.Logger.Trace("Parsed metadata", utils.Logger.Args("layers", len(files)));

	cache, err := json.Marshal(l.pr)
	if err != nil {
		utils.Logger.Error("Failed to marshal metadata", utils.Logger.Args("error", err))
		os.Exit(-1)
	}
	err = ioutil.WriteFile(utils.Config.BaseDir + "/metadata.json", cache, 0644)
	if err != nil {
		utils.Logger.Error("Failed to write metadata", utils.Logger.Args("error", err))
		os.Exit(-1)
	}

	return nil
}

func findmetdata(path string) (files []string) {
	utils.Logger.Trace("Searching...", utils.Logger.Args("path", path))
    err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
        if strings.Contains(path, "metadata.yaml") {
            files = append(files, path)
        }
        return err
    })

    if err != nil {
        panic(err)
    }
    return files
}