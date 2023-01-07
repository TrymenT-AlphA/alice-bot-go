package meta

import (
	"bot-go/src/util"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

func init() {
	zero.OnRegex(`^版本$`, zero.OnlyToMe).Handle(func(ctx *zero.Ctx) {
		Version(ctx)
		logrus.Infof("[meta][Version][success]")
	})
	zero.OnRegex(`^在吗$`, zero.OnlyToMe).Handle(func(ctx *zero.Ctx) {
		Alive(ctx)
		logrus.Infof("[meta][Alive][success]")
	})
}

func Version(ctx *zero.Ctx) {
	ctx.Send(message.Text("0.1"))
}

func Alive(ctx *zero.Ctx) {
	dice := util.GetDice(100)
	if dice > 0 {
		ctx.Send(message.Text("兔兔在哦~"))
	} else {
		ctx.Send(message.Text("在吗起手，必定小丑🤡"))
	}
}
