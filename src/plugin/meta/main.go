package meta

import (
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"alice-bot-go/src/core/alice"
)

var (
	plugin = "meta"
)

func init() {
	// usage/example: <NickName> 版本
	zero.OnRegex(`^版本$`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "version"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return version(ctx)
		})
	})
	// usage/example: <NickName> 源码
	zero.OnRegex(`^源码$`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "code"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return code(ctx)
		})
	})
	// usage/example: <NickName> 许可证
	zero.OnRegex(`^许可证$`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "license"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return license(ctx)
		})
	})
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
