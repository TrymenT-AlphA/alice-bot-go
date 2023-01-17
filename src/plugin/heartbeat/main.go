package heartbeat

import (
	"fmt"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"gorm.io/gorm"

	"alice-bot-go/src/core/alice"
	"alice-bot-go/src/core/config"
	"alice-bot-go/src/core/database/localSqlite3"
	"alice-bot-go/src/plugin/heartbeat/model"
)

var (
	initComplete = make(chan bool, 1)
	plugin       = "heartbeat"
	db           *gorm.DB
)

func init() {
	alice.Init.Register(func() {
		fn := "initialize"
		alice.CommandWapper(nil, true, plugin, fn, func() error {
			return initialize(initComplete)
		})
	}, 1)
	// usage: <NickName> 开启心跳
	// example: 兔兔 开启心跳
	zero.OnRegex(`^开启心跳$`, zero.OnlyToMe, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "turnon"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return turnon(ctx, db)
		})
	})
	// usage: <NickName> 关闭心跳
	// example: 兔兔 关闭心跳
	zero.OnRegex(`^关闭心跳$`, zero.OnlyToMe, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "turnoff"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return turnoff(ctx, db)
		})
	})
	go func() {
		<-initComplete
		fn := "heartbeat"
		alice.TickerWapper(time.Minute, false, plugin, fn, func(ctx *zero.Ctx) error {
			return heartbeat(ctx, db)
		})
	}()
}

func initialize(initComplete chan bool) error {
	var err error
	db, err = localSqlite3.Init(fmt.Sprintf("%s.db", plugin))
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(&model.Group{}); err != nil {
		return err
	}
	initComplete <- true
	return nil
}

func turnon(ctx *zero.Ctx, db *gorm.DB) error {
	if err := (&model.Group{
		ID: ctx.Event.GroupID,
	}).CreateOrUpdate(db); err != nil {
		return err
	}
	return nil
}

func turnoff(ctx *zero.Ctx, db *gorm.DB) error {
	if err := (&model.Group{
		ID: ctx.Event.GroupID,
	}).Delete(db); err != nil {
		return err
	}
	ctx.SetGroupCard(
		ctx.Event.GroupID, config.Bot.ID,
		fmt.Sprintf("%s", config.Bot.NickName[0]),
	)
	return nil
}

func heartbeat(ctx *zero.Ctx, db *gorm.DB) error {
	groups, err := (&model.Group{}).ReadAll(db)
	if err != nil {
		return err
	}
	for _, group := range groups {
		ctx.SetGroupCard(
			group.ID, config.Bot.ID,
			fmt.Sprintf("%s %s", config.Bot.NickName[0], time.Now().Format("01-02 15:04")),
		)
	}
	return nil
}
