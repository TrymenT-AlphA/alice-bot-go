package bilibili

import (
	"bot-go/src/database/localSqlite3"
	"bot-go/src/plugin/bilibili/model"
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

func init() {
	// 数据库连接迁移
	cwd, err := os.Getwd()
	if err != nil {
		logrus.Fatalf("[bilibili][database] %s", err)
	}
	db, err := localSqlite3.Init(filepath.Join(cwd, "..", "data", "database", "localSqlite3", "bilibili.db"))
	if err != nil {
		logrus.Fatalf("[bilibili][database] %s", err)
	}
	if err := db.AutoMigrate(&model.User{}); err != nil {
		logrus.Fatalf("[bilibili][database] %s", err)
	}
	// 指令定义
	zero.OnCommand("关注", zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			args := ctx.State["args"].(string)
			uid, err := strconv.ParseUint(args, 10, 64)
			if err != nil {
				logrus.Errorf("[bilibili][关注] %s", err)
				ctx.Send(message.Text("兔兔不懂"))
				return
			}
			user := model.User{
				Uid:   uid,
				Group: (uint64)(ctx.Event.GroupID),
			}
			if err := user.Create(db); err != nil {
				logrus.Errorf("[bilibili][关注] %s", err)
				ctx.Send(message.Text("兔兔坏掉了"))
				return
			}
			// success
			logrus.Infof("[bilibili][关注] 成功关注%s", args)
			ctx.Send(message.Text(fmt.Sprintf("兔兔记住了%s", args)))
		})
	zero.OnCommand("取关", zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).
		Handle(func(ctx *zero.Ctx) {
			args := ctx.State["args"].(string)
			uid, err := strconv.ParseUint(args, 10, 64)
			if err != nil {
				logrus.Errorf("[bilibili][取关] %s", err)
				ctx.Send(message.Text("兔兔不懂"))
				return
			}
			user := model.User{
				Uid:   uid,
				Group: (uint64)(ctx.Event.GroupID),
			}
			if err := user.Delete(db); err != nil {
				logrus.Errorf("[bilibili][取关] %s", err)
				ctx.Send(message.Text("兔兔坏掉了"))
				return
			}
			// success
			logrus.Infof("[bilibili][关注] 成功取关%s", args)
			ctx.Send(message.Text(fmt.Sprintf("兔兔忘记了%s", args)))
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
				PollingHandler(db)
				timer.Reset(interval)
			}
		}
	}()
}

func PollingHandler(db *gorm.DB) {
	ctx := zero.GetBot(2245788922)
	if ctx == nil {
		logrus.Warnf("[bilibili][轮询] %s", "无法获取ctx")
		return
	}
	users, err := model.User{}.ReadAll(db)
	if err != nil {
		logrus.Errorf("[bilibili][轮询] %s", "无法获取users")
		return
	}

	for _, user := range users {
		timestamp, description, pictures, err := user.GetLatestDynamic()
		if err != nil {
			logrus.Errorf("[bilibili][轮询] %s", err)
			ctx.Send(message.Text("兔兔被识破了"))
			return
		}

		if timestamp == user.Timestamp {
			continue
		}

		ctx.SendGroupMessage((int64)(user.Group), message.Text(description))

		for _, src := range pictures {
			ctx.SendGroupMessage((int64)(user.Group), message.Image(src))
		}

		user.Timestamp = timestamp
		if err := user.Update(db); err != nil {
			logrus.Errorf("[bilibili][轮询] %s", err)
			ctx.Send(message.Text("兔兔坏掉了"))
			return
		}
		if err := user.Update(db); err != nil {
			logrus.Errorf("[bilibili][轮询] %s", err)
			ctx.Send(message.Text("兔兔坏掉了"))
			return
		}
	}
}
