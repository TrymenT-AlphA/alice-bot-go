package github

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"

	"alice-bot-go/src/core/alice"
	"alice-bot-go/src/core/config"
	"alice-bot-go/src/core/database/localSqlite3"
	"alice-bot-go/src/core/util"
	"alice-bot-go/src/plugin/github/model"
)

var (
	initComplete = make(chan bool, 1)
	plugin       = "github"
	cache        string
	db           *gorm.DB
)

func init() {
	alice.Init.Register(func() {
		fn := "initialize"
		alice.CommandWapper(nil, true, plugin, fn, func() error {
			return initialize(initComplete)
		})
	}, 1)
	// usage: <NickName> 跟踪 <Repo.Owner> <Repo.Name> <local>
	// example: 兔兔 跟踪 Kengxxiao ArknightsGameData <local>
	zero.OnRegex(`^跟踪 (.+) (.+) (.+)$`, zero.OnlyToMe, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "follow"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return follow(ctx, db)
		})
	})
	// usage: <NickName> 跟踪 <Repo.Owner> <Repo.Name>
	// example: 兔兔 跟踪 Kengxxiao ArknightsGameData
	zero.OnRegex(`^跟踪 (.+) (.+)$`, zero.OnlyToMe, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "defaultFollow"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return defaultFollow(ctx, db)
		})
	})
	// usage: <NickName> 停止跟踪 <Repo.Owner> <Repo.Name>
	// example: 兔兔 停止跟踪 Kengxxiao ArknightsGameData
	zero.OnRegex(`^停止跟踪 (.+) (.+)$`, zero.OnlyToMe, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "unfollow"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return unfollow(ctx, db)
		})
	})
	go func() {
		<-initComplete
		fn := "polling"
		alice.TickerWapper(time.Second*10, false, plugin, fn, func(ctx *zero.Ctx) error {
			return polling(ctx, db)
		})
	}()
}

func initialize(initComplete chan bool) error {
	cache = filepath.Join(config.Global.CacheDir, plugin)
	if err := os.MkdirAll(cache, 0666); err != nil {
		return err
	}
	var err error
	db, err = localSqlite3.Init(fmt.Sprintf("%s.db", plugin))
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(&model.Task{}); err != nil {
		return err
	}
	if err := initGithubYml(); err != nil {
		return err
	}
	initComplete <- true
	return nil
}

func initGithubYml() error {
	fn := "initGithubYml"
	githubYml := filepath.Join(config.Global.ConfigDir, fmt.Sprintf("%s.yml", plugin))
	if util.IsNotExist(githubYml) {
		logrus.Infof("[%s][%s] %s", plugin, fn, fmt.Sprintf("no `%s.yml` was found, start configuring...", plugin))
		for { // config.Github.Token
			fmt.Printf("[%s][%s] %s", plugin, fn, fmt.Sprintf("enter %s:", "config.Github.Token"))
			n, err := fmt.Scan(&config.Github.Token)
			if n != 1 || err != nil {
				if errors.Is(err, io.EOF) {
					os.Exit(0)
				}
				fmt.Printf("[%s][%s] %s", plugin, fn, fmt.Sprintf("input err: %s, try again.\n", err))
				continue
			}
			break
		}
		data, err := yaml.Marshal(&config.Github)
		if err != nil {
			return err
		}
		if err = util.Write(githubYml, data); err != nil {
			return err
		}
		logrus.Infof("[%s][%s] %s", plugin, fn, "configure complete!")
	} else {
		logrus.Infof("[%s][%s] %s", plugin, fn, fmt.Sprintf("found `%s.yml`", plugin))
		data, err := os.ReadFile(githubYml)
		if err != nil {
			return err
		}
		if err = yaml.Unmarshal(data, &config.Github); err != nil {
			return err
		}
	}
	return nil
}

func defaultFollow(ctx *zero.Ctx, db *gorm.DB) error {
	args := ctx.State["regex_matched"].([]string)
	task := &model.Task{
		Repo: model.Repo{
			Owner: args[1],
			Name:  args[2],
			Local: "",
		},
		GroupID:   ctx.Event.GroupID,
		Timestamp: 0,
	}
	if err := task.CreateOrUpdate(db); err != nil {
		return err
	}
	ctx.Send(message.Text("跟踪成功"))
	return nil
}

func follow(ctx *zero.Ctx, db *gorm.DB) error {
	args := ctx.State["regex_matched"].([]string)
	task := &model.Task{
		Repo: model.Repo{
			Owner: args[1],
			Name:  args[2],
			Local: args[3],
		},
		GroupID:   ctx.Event.GroupID,
		Timestamp: 0,
	}
	if err := task.CreateOrUpdate(db); err != nil {
		return err
	}
	ctx.Send(message.Text("跟踪成功"))
	return nil
}

func unfollow(ctx *zero.Ctx, db *gorm.DB) error {
	args := ctx.State["regex_matched"].([]string)
	task := &model.Task{
		Repo: model.Repo{
			Owner: args[1],
			Name:  args[2],
		},
		GroupID: ctx.Event.GroupID,
	}
	if err := task.Delete(db); err != nil {
		return err
	}
	ctx.Send(message.Text("成功停止跟踪"))
	return nil
}

func polling(ctx *zero.Ctx, db *gorm.DB) error {
	tasks, err := (&model.Task{}).ReadAll(db)
	if err != nil {
		return err
	}
	for _, task := range tasks {
		commit, err := task.Repo.GetLatestCommit(fmt.Sprintf("token %s", config.Github.Token))
		if err != nil {
			return err
		}
		if commit.Timestamp <= task.Timestamp {
			continue
		}
		logrus.Infof("[github][Polling] Cloning/Pulling ...")
		ctx.SendGroupMessage(task.GroupID, message.Text(
			fmt.Sprintf("检测到数据更新，开始同步\n%s\\%s\n%s\n%s", task.Repo.Owner, task.Repo.Name, commit.Date, commit.Message),
		))
		var local string
		if task.Repo.Local == "" {
			local, err = task.Repo.DefaultLocal()
			if err != nil {
				return err
			}
		} else {
			local = filepath.Join(config.Global.Cwd, filepath.FromSlash(task.Repo.Local))
		}
		if err = task.Repo.CloneOrPull(local); err != nil {
			return err
		}
		logrus.Infof("[github][Polling] Cloning/Pulling Complete!")
		ctx.SendGroupMessage(task.GroupID, message.Text("数据更新完毕"))
		task.Timestamp = commit.Timestamp
		if err = task.Update(db); err != nil {
			return err
		}
	}
	return nil
}
