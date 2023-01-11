package github

import (
	"alice-bot-go/src/config"
	"alice-bot-go/src/database/localSqlite3"
	"alice-bot-go/src/plugin/github/model"
	"fmt"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"path/filepath"
	"time"
)

var (
	db      *gorm.DB
	repoDir string
)

func init() {
	// initialize
	err := Initialize()
	if err != nil {
		logrus.Fatalf("[github][Initialize] %s", err)
	} else {
		logrus.Infof("[github][Initialize][success]")
	}

	// example: 兔兔 跟踪 Kengxxiao ArknightsGameData
	zero.OnRegex(`^跟踪 (.+) (.+)$`, zero.OnlyToMe, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := DefaultFollow(ctx)
		if err != nil {
			logrus.Errorf("[github][DefaultFollow] %s", err)
			ctx.Send(message.Text(fmt.Sprintf("[github][DefaultFollow] %s", err)))
		} else {
			logrus.Infof("[github][DefaultFollow][success]")
		}
	})

	// example: 兔兔 跟踪 Kengxxiao ArknightsGameData <path to your local repo>
	zero.OnRegex(`^跟踪 (.+) (.+) (.+)$`, zero.OnlyToMe, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := Follow(ctx)
		if err != nil {
			logrus.Errorf("[github][Follow] %s", err)
			ctx.Send(message.Text(fmt.Sprintf("[github][Follow] %s", err)))
		} else {
			logrus.Infof("[github][Follow][success]")
		}
	})

	// example: 兔兔 停止跟踪 Kengxxiao ArknightsGameData
	zero.OnRegex(`^停止跟踪 (.+) (.+)$`, zero.OnlyToMe, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := Unfollow(ctx)
		if err != nil {
			logrus.Errorf("[github][Unfollow] %s", err)
			ctx.Send(message.Text(fmt.Sprintf("[github][Unfollow] %s", err)))
		} else {
			logrus.Infof("[github][Unfollow][success]")
		}
	})

	// 轮询
	go func() {
		interval := time.Second * 2
		timer := time.NewTimer(interval)
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, os.Kill)

		for {
			select {
			case <-quit: // handle quit first
				os.Exit(0)

			case <-timer.C:
				ctx := zero.GetBot(2245788922)
				if ctx == nil {
					logrus.Warnf("[github][Polling] GetBot failed")
					continue
				}

				err := Polling(ctx)
				if err != nil {
					logrus.Errorf("[github][Polling] %s", err)
				} else {
					logrus.Infof("[github][Polling][success]")
				}

				timer.Reset(interval)
			}
		}
	}()
}

func Initialize() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	repoDir = filepath.Join(cwd, "..", "data", "repo")

	err = os.MkdirAll(repoDir, 0666)
	if err != nil {
		return err
	}

	db, err = localSqlite3.Init("github.db")
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&model.Task{})
	if err != nil {
		return err
	}

	return nil
}

func DefaultFollow(ctx *zero.Ctx) error {
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

	err := task.CreateOrUpdate(db)
	if err != nil {
		return err
	}

	ctx.Send(message.Text("跟踪成功"))

	return nil
}

func Follow(ctx *zero.Ctx) error {
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

	err := task.CreateOrUpdate(db)
	if err != nil {
		return err
	}

	ctx.Send(message.Text("跟踪成功"))

	return nil
}

func Unfollow(ctx *zero.Ctx) error {
	args := ctx.State["regex_matched"].([]string)

	task := &model.Task{
		Repo: model.Repo{
			Owner: args[1],
			Name:  args[2],
		},
		GroupID: ctx.Event.GroupID,
	}

	err := task.Delete(db)
	if err != nil {
		return err
	}

	ctx.Send(message.Text("成功停止跟踪"))

	return nil
}

func Polling(ctx *zero.Ctx) error {
	tasks, err := (&model.Task{}).ReadAll(db)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		commit, err := task.Repo.GetLatestCommit(fmt.Sprintf("token %s", config.GITHUBCONFIG.Token))
		if err != nil {
			return err
		}

		if commit.Timestamp <= task.Timestamp {
			continue
		}

		ctx.SendGroupMessage(task.GroupID, message.Text(
			fmt.Sprintf("%s\\%s\n%s\n%s", task.Repo.Owner, task.Repo.Name, commit.Date, commit.Message),
		))

		ctx.SendGroupMessage(task.GroupID, message.Text("检测到数据更新，开始同步"))

		logrus.Infof("[github][Polling] Cloning/Pulling ...")
		if task.Repo.Local == "" {
			local, err := task.Repo.DefaultLocal()
			if err != nil {
				return err
			}
			err = task.Repo.CloneOrPull(local)
			if err != nil {
				return err
			}
		} else {
			err = task.Repo.CloneOrPull(task.Repo.Local)
			if err != nil {
				return err
			}
		}
		logrus.Infof("[github][Polling] Cloning/Pulling Complete!")

		ctx.SendGroupMessage(task.GroupID, message.Text("数据更新完毕"))

		task.Timestamp = commit.Timestamp

		err = task.Update(db)
		if err != nil {
			return err
		}
	}

	return nil
}
