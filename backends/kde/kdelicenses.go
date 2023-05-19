package kde

import (
	"strings"

	"github.com/Fishwaldo/go-yocto/utils"
	"github.com/xanzy/go-gitlab"
)



func GetLicense(pr Project) (license []string, err error) {
	utils.Logger.Trace("Getting License", utils.Logger.Args("project", pr.Name))
	gl, err := gitlab.NewClient(utils.Config.KDEConfig.AccessToken, gitlab.WithBaseURL(utils.Config.KDEConfig.KDEGitLabURL+"/api/v4"))

	gf := &gitlab.ListTreeOptions {
		Path: gitlab.String("LICENSES"),
		Ref: gitlab.String(pr.MetaData["branch-rules"]["branch"].(string)),
	}
	f, res, err := gl.Repositories.ListTree(pr.Repopath, gf)
	if err != nil {
		utils.Logger.Error("Failed to get License", utils.Logger.Args("error", err))
		return nil, err
	}
	if res.StatusCode != 200 {
		utils.Logger.Error("Failed to get License", utils.Logger.Args("error", res.Status))
		return nil, err
	}
	for _, file := range f {
		license = append(license, strings.TrimSuffix(file.Name, ".txt"))
	}
	return license, nil
}