package alive

import (
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"math/rand"
	"time"
)

func init() {
	// 设定随机数种子
	rand.Seed(time.Now().Unix())
	// 指令定义
	zero.OnCommand("在吗", zero.OnlyToMe).Handle(func(ctx *zero.Ctx) {
		dice := rand.Intn(100)
		logrus.Infof("[alive][在吗] 判定dice: %v", dice)
		if dice > 0 {
			ctx.Send(message.Text("兔兔在哦~"))
		} else {
			ctx.Send(message.Text("在吗起手，必定小丑🤡"))
		}
	})
}
