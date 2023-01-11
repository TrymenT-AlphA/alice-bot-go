package bilibili

import (
	"alice-bot-go/src/database/localSqlite3"
	"alice-bot-go/src/plugin/bilibili/model"
	"fmt"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"
)

var (
	db    *gorm.DB
	cache string
)

func init() {
	// initialize
	err := Initialize()
	if err != nil {
		logrus.Fatalf("[bilibili][Initialize] %s", err)
	} else {
		logrus.Infof("[bilibili][Initialize][success]")
	}

	// example: 兔兔 关注 明日方舟 161775300
	zero.OnRegex("^关注 (.+) (.+)$", zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := Subscribe(ctx, db)
		if err != nil {
			logrus.Errorf("[bilibili][Subscribe] %s", err)
			ctx.Send(message.Text(fmt.Sprintf("[bilibili][Subscribe] %s", err)))
		} else {
			logrus.Infof("[bilibili][Subscribe][success]")
		}
	})

	// example: 兔兔 取关 明日方舟
	zero.OnRegex("^取关 (.+)$", zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := Unsubscribe(ctx, db)
		if err != nil {
			logrus.Errorf("[bilibili][Unsubscribe] %s", err)
			ctx.Send(message.Text(fmt.Sprintf("[bilibili][Unsubscribe] %s", err)))
		} else {
			logrus.Infof("[bilibili][Unsubscribe][success]")
		}
	})

	// 轮询
	go func() {
		interval := time.Second * 10
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
					logrus.Warnf("[bilibili][Polling] GetBot failed")
					continue
				}

				err := Polling(ctx)
				if err != nil {
					logrus.Errorf("[bilibili][Polling] %s", err)
				} else {
					logrus.Infof("[bilibili][Polling][success]")
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

	cache = filepath.Join(cwd, "..", "data", "cache", "bilibili")
	err = os.MkdirAll(cache, 0666)
	if err != nil {
		return err
	}

	db, err = localSqlite3.Init("bilibili.db")
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&model.Task{})
	if err != nil {
		return err
	}

	return nil
}

func Subscribe(ctx *zero.Ctx, db *gorm.DB) error {
	args := ctx.State["regex_matched"].([]string)

	uid, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return err
	}

	task := &model.Task{
		Up: model.Up{
			NicName: args[1],
			UID:     uid,
		},
		GroupID:   ctx.Event.GroupID,
		Timestamp: 0,
	}

	err = task.CreateOrUpdate(db)
	if err != nil {
		return err
	}

	ctx.Send(message.Text("关注成功"))

	return nil
}

func Unsubscribe(ctx *zero.Ctx, db *gorm.DB) error {
	args := ctx.State["regex_matched"].([]string)

	task := &model.Task{
		Up: model.Up{
			NicName: args[1],
		},
		GroupID: ctx.Event.GroupID,
	}

	err := task.Delete(db)
	if err != nil {
		return err
	}

	ctx.Send(message.Text("取关成功"))

	return nil
}

func Polling(ctx *zero.Ctx) error {
	tasks, err := (&model.Task{}).ReadAll(db)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		dynamic, err := task.Up.GetLatestDynamic()
		if err != nil {
			return err
		}

		if dynamic.Timestamp <= task.Timestamp {
			continue
		}

		ctx.SendGroupMessage(task.GroupID, message.Text(dynamic.Description))
		for _, src := range dynamic.Pictures {
			ctx.SendGroupMessage(task.GroupID, message.Image(src))
		}

		task.Timestamp = dynamic.Timestamp

		err = task.Update(db)
		if err != nil {
			return err
		}
	}

	return nil
}
