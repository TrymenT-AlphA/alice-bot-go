package model

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/tidwall/gjson"

	"alice-bot-go/src/core/alice"
	"alice-bot-go/src/core/util"
)

type Repo struct {
	Owner string `gorm:"primarykey"`
	Name  string `gorm:"primarykey"`
	Local string `gorm:"primarykey"`
}

func (*Repo) TableName() string {
	return "Repo"
}

func (repo *Repo) Url() string {
	return fmt.Sprintf("https://github.com/%s/%s.git", repo.Owner, repo.Name)
}

func (repo *Repo) HtmlUrl() string {
	return fmt.Sprintf("https://github.com/%s/%s", repo.Owner, repo.Name)
}

func (repo *Repo) DefaultLocal() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	defaultLocal := filepath.Join(cwd, "..", "data", "cache", "github", fmt.Sprintf("%s/%s", repo.Owner, repo.Name))
	return defaultLocal, nil
}

func (repo *Repo) GetLatestCommit(auth string) (*Commit, error) {
	api, err := alice.NewAPI("github", "repos", "commits")
	if err != nil {
		return nil, err
	}
	api.UrlParams = []interface{}{repo.Owner, repo.Name}
	api.Params = map[string]interface{}{
		"page":     1,
		"per_page": 1,
	}
	api.Header = map[string]string{
		"Authorization": auth,
	}
	data, err := api.DoRequest(&http.Client{})
	if err != nil {
		return nil, err
	}
	c := gjson.GetBytes(data, "0.commit")
	commitDate := c.Get("committer.date").String()
	commitMsg := c.Get("message").String()
	commitTime, err := time.Parse("2006-01-02T15:04:05Z", commitDate)
	if err != nil {
		return nil, err
	}
	commit := &Commit{
		Date:      commitDate,
		Message:   commitMsg,
		Timestamp: commitTime.Unix(),
	}
	return commit, nil
}

func (repo *Repo) Clone(local string) error {
	_, err := git.PlainClone(local, false, &git.CloneOptions{
		URL:               repo.Url(),
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if err != nil && !errors.Is(err, git.ErrRepositoryAlreadyExists) {
		return err
	}
	return nil
}

func (repo *Repo) Pull(local string) error {
	repository, err := git.PlainOpen(local)
	if err != nil {
		return err
	}
	worktree, err := repository.Worktree()
	if err != nil {
		return err
	}
	err = worktree.Pull(&git.PullOptions{RemoteName: "origin"})
	if err != nil && !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return err
	}
	return nil
}

func (repo *Repo) CloneOrPull(local string) error {
	if util.IsNotExist(local) {
		if err := repo.Clone(local); err != nil {
			return err
		}
	} else {
		if err := repo.Pull(local); err != nil {
			return err
		}
	}
	return nil
}
