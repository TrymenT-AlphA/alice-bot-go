package bilibili

import (
	"fmt"
	"strconv"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"

	"alice-bot-go/src/core/alice"
	"alice-bot-go/src/core/database/localSqlite3"
	"alice-bot-go/src/plugin/bilibili/model"
)

var (
	initComplete = make(chan bool, 1)
	plugin       = "bilibili"
	db           *gorm.DB
)

func init() {
	alice.Initializer.Register(func() {
		fn := "initialize"
		alice.CommandWapper(nil, true, plugin, fn, func() error {
			return initialize(initComplete)
		})
	}, 1)
	// usage: <NickName> 关注 <Up.NickName> <Up.UID>
	// example: 兔兔 关注 明日方舟 161775300
	zero.OnRegex("^关注 (.+) (.+)$", zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "subscribe"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return subscribe(ctx, db)
		})
	})
	// usage: <NickName> 取关 <Up.NickName>
	// example: 兔兔 取关 明日方舟
	zero.OnRegex("^取关 (.+)$", zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "unsubscribe"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return unsubscribe(ctx, db)
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
	var err error
	db, err = localSqlite3.Init(fmt.Sprintf("%s.db", plugin))
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(&model.Task{}); err != nil {
		return err
	}
	initComplete <- true
	return nil
}

func subscribe(ctx *zero.Ctx, db *gorm.DB) error {
	args := ctx.State["regex_matched"].([]string)
	uid, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return err
	}
	task := &model.Task{
		Up: model.Up{
			NickName: args[1],
			UID:      uid,
		},
		GroupID:   ctx.Event.GroupID,
		Timestamp: 0,
	}
	if err := task.CreateOrUpdate(db); err != nil {
		return err
	}
	ctx.Send(message.Text("关注成功"))
	return nil
}

func unsubscribe(ctx *zero.Ctx, db *gorm.DB) error {
	args := ctx.State["regex_matched"].([]string)
	task := &model.Task{
		Up: model.Up{
			NickName: args[1],
		},
		GroupID: ctx.Event.GroupID,
	}
	if err := task.Delete(db); err != nil {
		return err
	}
	ctx.Send(message.Text("取关成功"))
	return nil
}

func polling(ctx *zero.Ctx, db *gorm.DB) error {
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
		if dynamic.Origin != nil {
			ctx.SendGroupMessage(task.GroupID, message.Text(dynamic.Origin.Description))
			for _, src := range dynamic.Origin.Pictures {
				ctx.SendGroupMessage(task.GroupID, message.Image(src))
			}
		}
		task.Timestamp = dynamic.Timestamp
		if err := task.Update(db); err != nil {
			return err
		}
	}
	return nil
}
