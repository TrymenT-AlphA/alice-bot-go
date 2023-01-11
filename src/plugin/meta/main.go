package meta

import (
	"alice-bot-go/src/util"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	zero.OnRegex(`^版本$`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		Version(ctx)
		logrus.Infof("[meta][Version][success]")
	})
	zero.OnRegex(`^在吗$`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		Alive(ctx)
		logrus.Infof("[meta][Alive][success]")
	})
	zero.OnRegex(`^源码$`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		SourceCode(ctx)
		logrus.Infof("[meta][SourceCode][success]")
	})
	zero.OnRegex(`^许可证$`, zero.OnlyToMe).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		License(ctx)
		logrus.Infof("[meta][License][success]")
	})
}

func Version(ctx *zero.Ctx) {
	ctx.Send(message.Text("v0.2"))
}

func Alive(ctx *zero.Ctx) {
	dice := util.GetDice(100)
	if dice > 0 {
		ctx.Send(message.Text("兔兔在哦~"))
	} else {
		ctx.Send(message.Text("在吗起手，必定小丑🤡"))
	}
}

func SourceCode(ctx *zero.Ctx) {
	ctx.Send(message.Text())
}

func License(ctx *zero.Ctx) {
	ctx.Send(message.Text())
}
