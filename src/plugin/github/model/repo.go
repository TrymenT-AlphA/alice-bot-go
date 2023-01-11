package model

import (
	"alice-bot-go/src/types"
	"errors"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/tidwall/gjson"
	"net/http"
	"os"
	"path/filepath"
	"time"
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

	return filepath.Join(cwd, "..", "data", "repo", fmt.Sprintf("%s/%s", repo.Owner, repo.Name)), nil
}

func (repo *Repo) GetLatestCommit(token string) (*Commit, error) {
	githubAPI, err := types.NewRestAPI("github", "repos", "commits")
	if err != nil {
		return nil, err
	}

	githubAPI.UrlParams = []interface{}{repo.Owner, repo.Name}

	githubAPI.Params["page"] = 1
	githubAPI.Params["per_page"] = 1

	data, err := githubAPI.DoRequestAuth(&http.Client{}, token)
	if err != nil {
		return nil, err
	}

	temp := gjson.GetBytes(data, "0.commit")
	commitDate := temp.Get("committer.date").String()
	commitMsg := temp.Get("message").String()
	commitTimestamp, err := time.Parse("2006-01-02T15:04:05Z", commitDate)
	if err != nil {
		return nil, err
	}

	commit := &Commit{
		Date:      commitDate,
		Message:   commitMsg,
		Timestamp: commitTimestamp.Unix(),
	}

	return commit, nil
}

func (repo *Repo) Clone(local string) error {
	_, err := git.PlainClone(local, false, &git.CloneOptions{
		URL:               repo.Url(),
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})
	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		return nil
	} else if err != nil {
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
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil
	} else if err != nil {
		return err
	}

	return nil
}

func (repo *Repo) CloneOrPull(local string) error {
	_, err := os.Stat(local)
	if err == nil {
		err = repo.Pull(local)
		if err != nil {
			return err
		}
	} else if os.IsNotExist(err) {
		err = repo.Clone(local)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return nil
}
