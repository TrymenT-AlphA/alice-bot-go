package bilibili

import (
	"bot-go/src/database/localSqlite3"
	"bot-go/src/plugin/bilibili/model"
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
	db *gorm.DB
)

func init() {
	err := InitAndMigrate()
	if err != nil {
		logrus.Fatalf("[bilibili][InitAndMigrate] %s", err)
	}
	// example: 兔兔 关注 明日方舟 161775300
	zero.OnRegex("^关注 (.+) (.+)$", zero.OnlyToMe, zero.OnlyGroup).
		SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := Subscribe(ctx, db)
		if err != nil {
			logrus.Errorf("[bilibili][Subscribe] %s", err)
		} else {
			logrus.Infof("[bilibili][Subscribe][Success]")
		}
	})
	// example: 兔兔 取关 明日方舟
	zero.OnRegex("取关 (.+)", zero.OnlyToMe, zero.OnlyGroup).
		SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := Unsubscribe(ctx, db)
		if err != nil {
			logrus.Errorf("[bilibili][Subscribe] %s", err)
		} else {
			logrus.Infof("[bilibili][Unsubscribe][Success]")
		}
	})
	// 轮询
	go func() {
		interval := time.Second * 3
		timer := time.NewTimer(interval)
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, os.Kill)

		for {
			select {
			case <-quit: // handle quit first
				os.Exit(0)
			case <-timer.C:
				err := Polling(db)
				if err != nil {
					logrus.Errorf("[bilibili][Polling] %s", err)
				}
				timer.Reset(interval)
			}
		}
	}()
}

func InitAndMigrate() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	db, err = localSqlite3.Init(
		filepath.Join(cwd, "..", "data", "database", "localSqlite3", "bilibili.db"),
	)
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

	ctx.Send("兔兔记住了")

	return nil
}

func Unsubscribe(ctx *zero.Ctx, db *gorm.DB) error {
	args := ctx.State["regex_matched"].([]string)

	user := &model.Task{
		Up: model.Up{
			NicName: args[1],
		},
		GroupID: ctx.Event.GroupID,
	}

	err := user.Delete(db)
	if err != nil {
		return err
	}

	ctx.Send("兔兔忘记了")

	return nil
}

func Polling(db *gorm.DB) error {
	ctx := zero.GetBot(2245788922)
	if ctx == nil {
		logrus.Warnf("[bilibili][Polling] %s", "GetBot failed")
		return nil
	}

	tasks, err := model.Task{}.ReadAll(db)
	if err != nil {
		return err
	}

	for _, task := range tasks {
		dynamic, err := task.Up.GetLatestDynamic()
		if err != nil {
			return err
		}

		if dynamic.Timestamp == task.Timestamp {
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
