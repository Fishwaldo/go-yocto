package kde

import (
	"encoding/base64"
	"encoding/json"

	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/Fishwaldo/go-yocto/parsers"
	"github.com/Fishwaldo/go-yocto/repo"
	"github.com/Fishwaldo/go-yocto/source"
	"github.com/Fishwaldo/go-yocto/utils"

	"golang.org/x/exp/maps"

	"github.com/pterm/pterm"
	"github.com/spf13/viper"
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

type Project struct {
	source.RecipeSource	`yaml:",inline"`
	ProjectPath string
	Repoactive bool
	Repopath string
	Identifier string
	Hasrepo bool
	Source string
	Bugzilla struct {
		Product string
		Component string
		DNUlegacyproduct string `yaml:"__do_not_use-legacy-product"`
	}
	Topics []string
}

type KDEBe struct {
	MetaDataRepo repo.Repo
	br map[string]map[string]string
	pr map[string]Project
	ready bool
}

func init() {
	viper.SetDefault("kdeconfig.release", "@stable");
	viper.SetDefault("kdeconfig.defaultbranch", "master");
	viper.SetDefault("kdeconfig.kdegitlaburl", "https://invent.kde.org/")
}

func NewBackend() (l *KDEBe) {
	l = &KDEBe{}
	return l
}

func (l *KDEBe) GetName() string {
	return "kde-invent"
}

func (l *KDEBe) Init() (err error) {
	utils.Logger.Trace("Initializing KDE Backend")
	l.MetaDataRepo = repo.Repo{
		Url: "https://invent.kde.org/sysadmin/repo-metadata",
		Name: "kde-metadata",
	}
	l.ready = true
	return nil
}

func (l *KDEBe) Ready() bool {
	return l.ready
}

func (l *KDEBe) getDir() (dir string) {
	dir = utils.Config.BaseDir + "/" + l.MetaDataRepo.Name
	return dir
}

func (l *KDEBe) LoadSource() (err error) {
	utils.Logger.Trace("Checking metadata repo", utils.Logger.Args("repo", l.MetaDataRepo, "layer", l.GetName()))
	err = l.MetaDataRepo.CheckRepo()
	if (err != nil) {
		utils.Logger.Info("Cloning repo", utils.Logger.Args("repo", l.MetaDataRepo, "layer", l.GetName()))
		err := l.MetaDataRepo.CloneRepo()
		if (err != nil) {
			utils.Logger.Error("Failed to clone repo", utils.Logger.Args("error", err))
			os.Exit(-1)
		}
	}
	maps.Clear(l.br) 
	maps.Clear(l.pr)
	err = l.parseMetadata()
	if (err != nil) {
		utils.Logger.Error("Failed to parse metadata", utils.Logger.Args("error", err))
		os.Exit(-1)
	}
	return nil
}


func (l *KDEBe) parseMetadata() (err error) {
	l.pr = make(map[string]Project)
	utils.Logger.Trace("Parsing metadata", utils.Logger.Args("layer", l.GetName()))
	brfile, err := ioutil.ReadFile(l.getDir() + "/branch-rules.yml")
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
		utils.Logger.Error("Invalid release", utils.Logger.Args("release", utils.Config.KDEConfig.Release))
		os.Exit(-1)
	}

	gl, err := gitlab.NewClient(utils.Config.KDEConfig.AccessToken, gitlab.WithBaseURL(utils.Config.KDEConfig.KDEGitLabURL+"/api/v4"))
	if err != nil {
		utils.Logger.Error("Failed to create GitLab client", utils.Logger.Args("error", err))
		os.Exit(-1)
	}

	/* now parse the directory */
	files := findmetdata(l.getDir())
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
		data.MetaData = make(map[string]map[string]interface{})
		data.MetaData["branch-rules"] = make(map[string]interface{})
		data.MetaData["branch-rules"]["branch"] = utils.Config.KDEConfig.DefaultBranch
		data.RecipeSource.Backend = l.GetName()
		data.RecipeSource.Url, _ = url.JoinPath(utils.Config.KDEConfig.KDEGitLabURL,  data.Repopath)
		/* find out which branch this is in... */
		for project, branch := range l.br[utils.Config.KDEConfig.Release] {
			ok, _ := filepath.Match(project, data.Repopath)
			if ok {
				data.MetaData["branch-rules"]["branch"] = branch
				break
			}
		}

		/* get the .kde-ci.yml for dependencies */
		gf := &gitlab.GetFileOptions{
			Ref: gitlab.String(data.MetaData["branch-rules"]["branch"].(string)),
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
					data.MetaData["kde-ci"] = make(map[string]interface{})
					data.MetaData["kde-ci"]["dependencies"] = deps
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
			/* appstream */
			content, err := base64.StdEncoding.DecodeString(f.Content)
			if err != nil {
				utils.Logger.Error("Failed to decode appstream", utils.Logger.Args("error", err))
			} else {
				if as, err := parsers.GetParser("appstream"); err != nil {
					utils.Logger.Error("Failed to get appstream parser", utils.Logger.Args("error", err))
				} else {
					if data.MetaData["appstream"], err = as.Parse(strings.NewReader(string(content))); err != nil {
						utils.Logger.Error("Failed to parse appstream", utils.Logger.Args("error", err))
					}
				}

			}
		}

		if _, ok := l.pr[data.Identifier]; ok {
			utils.Logger.Error("Duplicate identifier", utils.Logger.Args("identifier", data.Identifier))
			continue
		}
		data.Source = l.GetName()
		l.pr[data.Identifier] = data
	}

	cache, err := json.Marshal(l.pr)
	if err != nil {
		utils.Logger.Error("Failed to marshal metadata", utils.Logger.Args("error", err))
		os.Exit(-1)
	}
	err = ioutil.WriteFile(utils.Config.BaseDir + "/" + l.GetName() + "-cache.json", cache, 0644)
	if err != nil {
		utils.Logger.Error("Failed to write metadata", utils.Logger.Args("error", err))
		os.Exit(-1)
	}
	brcache, err := json.Marshal(l.br)
	if err != nil {
		utils.Logger.Error("Failed to marshal branch metadata", utils.Logger.Args("error", err))
		os.Exit(-1)
	}
	err = ioutil.WriteFile(utils.Config.BaseDir + "/" + l.GetName() + "-branch-cache.json", brcache, 0644)
	if err != nil {
		utils.Logger.Error("Failed to write branch metadata", utils.Logger.Args("error", err))
		os.Exit(-1)
	}

	utils.Logger.Trace("Parsed metadata", utils.Logger.Args("layers", len(l.pr), "branches", len(l.br)));


	return nil
}

func (l *KDEBe) LoadCache() (err error) {
	utils.Logger.Trace("Loading KDE Cache")
	cache, err := ioutil.ReadFile(utils.Config.BaseDir + "/" + l.GetName() + "-cache.json")
	if err != nil {
		utils.Logger.Error("Failed to read cache", utils.Logger.Args("error", err))
	} else {
		err = json.Unmarshal(cache, &l.pr)
		if err != nil {
			utils.Logger.Error("Failed to unmarshal cache", utils.Logger.Args("error", err))
		}
	}
	brcache, err := ioutil.ReadFile(utils.Config.BaseDir + "/" + l.GetName() + "-branch-cache.json")
	if err != nil {
		utils.Logger.Error("Failed to read branch cache", utils.Logger.Args("error", err))
	} else {
		err = json.Unmarshal(brcache, &l.br)
		if err != nil {
			utils.Logger.Error("Failed to unmarshal branch cache", utils.Logger.Args("error", err))
		}
	}

	utils.Logger.Trace("KDE Cache Loaded", utils.Logger.Args("layers", len(l.pr), "branches", len(l.br)))
	return nil
}

func (l *KDEBe) SearchSource(keywords string) (source []source.RecipeSource, err error) {
	utils.Logger.Trace("Searching KDE Source", utils.Logger.Args("keyword", keywords))

	p, _ := pterm.DefaultProgressbar.WithTotal(len(l.pr)).WithTitle("Searching KDE...").Start()

	for _, data := range l.pr {
		p.Increment()
		if strings.Contains(strings.ToLower(data.Name), strings.ToLower(keywords)) {
			source = append(source, data.RecipeSource)
		}
		if strings.Contains(strings.ToLower(data.Description), strings.ToLower(keywords)) {
			source = append(source, data.RecipeSource)
		}
	}
	return source, nil;
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