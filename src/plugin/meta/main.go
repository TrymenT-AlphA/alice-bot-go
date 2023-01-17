package meta

import (
	"fmt"
	"time"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"

	"alice-bot-go/src/core/alice"
	"alice-bot-go/src/core/config"
	"alice-bot-go/src/core/database/localSqlite3"
	"alice-bot-go/src/plugin/meta/model"
)

var (
	initComplete = make(chan bool, 1)
	plugin       = "meta"
	db           *gorm.DB
)

func init() {
	alice.Initializer.Register(func() {
		fn := "initialize"
		alice.CommandWapper(nil, true, plugin, fn, func() error {
			return initialize(initComplete)
		})
	}, 1)
	engine := zero.New()
	engine.SetBlock(true)
	// usage: <NickName> 版本
	// example: 兔兔 版本
	engine.OnRegex(`^版本$`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "version"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return version(ctx)
		})
	})
	// usage: <NickName> 源码
	// example: 兔兔 源码
	engine.OnRegex(`^源码$`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "code"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return code(ctx)
		})
	})
	// usage: <NickName> 许可证
	// example: 兔兔 许可证
	engine.OnRegex(`^许可证$`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "license"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return license(ctx)
		})
	})
	// usage: <NickName> 启动心跳模式
	// example: 兔兔 启动心跳模式
	zero.OnRegex(`^启动心跳模式$`, zero.OnlyToMe, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "turnonHeartbeat"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return turnonHeartbeat(ctx, db)
		})
	})
	// usage: <NickName> 关闭心跳模式
	// example: 兔兔 关闭心跳模式
	zero.OnRegex(`^关闭心跳模式$`, zero.OnlyToMe, zero.OnlyGroup, zero.SuperUserPermission).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "turnoffHeartbeat"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return turnoffHeartbeat(ctx, db)
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

func version(ctx *zero.Ctx) error {
	ctx.Send(message.Text("v0.3"))
	return nil
}

func code(ctx *zero.Ctx) error {
	ctx.Send(message.Text("https://github.com/TrymenT-AlphA/alice-bot-go"))
	return nil
}

func license(ctx *zero.Ctx) error {
	ctx.Send(message.Text("GNU General Public License v3.0"))
	return nil
}

func turnonHeartbeat(ctx *zero.Ctx, db *gorm.DB) error {
	if err := (&model.Group{
		ID: ctx.Event.GroupID,
	}).CreateOrUpdate(db); err != nil {
		return err
	}
	return nil
}

func turnoffHeartbeat(ctx *zero.Ctx, db *gorm.DB) error {
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
