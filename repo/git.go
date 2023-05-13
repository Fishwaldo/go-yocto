package repo

import (
	"fmt"
	"os"

	"github.com/Fishwaldo/go-yocto/utils"

	"github.com/go-git/go-git/v5"
)

type Repo struct {
	Url string
	Name string
	Repo *git.Repository
}

func (r *Repo) CheckRepo() (err error) {
	path := fmt.Sprintf("%s/%s", utils.Config.BaseDir, r.Name)

	utils.Logger.Trace("Opening Repo", utils.Logger.Args("path", path))
	r.Repo, err = git.PlainOpen(path)
	if err != nil {
		utils.Logger.Error("Failed to open repo", utils.Logger.Args("error", err))
		return err
	}
	// Print the latest commit that was just pulled
	ref, err := r.Repo.Head()
	if err != nil {
		utils.Logger.Error("Failed to get head", utils.Logger.Args("error", err))
		os.Exit(-1)
	}
	commit, err := r.Repo.CommitObject(ref.Hash())
	if err != nil {
		utils.Logger.Error("Failed to get commit", utils.Logger.Args("error", err))
		os.Exit(-1)
	}
	utils.Logger.Info("Repo Stats", utils.Logger.Args("commit", commit.Hash, "message", commit.Message, "author", commit.Committer.Email, "date", commit.Author.When, "path", path))
	return nil
}

func (r *Repo) CloneRepo() (err error)  {
	path := fmt.Sprintf("%s/%s", utils.Config.BaseDir, r.Name)
	r.Repo, err = git.PlainClone(path, false, &git.CloneOptions{
		URL: r.Url,
	})
	if err != nil {
		return err
	}
	utils.Logger.Info("Cloned repo", utils.Logger.Args("repo", r))
	return nil;
}
