package alice

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"

	"alice-bot-go/src/core/config"
)

func CommandWapper(ctx *zero.Ctx, loginfo bool, plugin, fn string, f func() error) {
	if err := f(); err != nil {
		logrus.Errorf("[%s][%s] %s", plugin, fn, err)
		if ctx != nil {
			ctx.Send(message.Text(fmt.Sprintf("[%s][%s] %s", plugin, fn, err)))
		}
	} else if loginfo {
		logrus.Infof("[%s][%s] %s", plugin, fn, "success")
	}
}

func TickerWapper(duration time.Duration, loginfo bool, plugin string, fn string, f func(ctx *zero.Ctx) error) {
	ticker := time.NewTicker(duration)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	for {
		select {
		case <-quit:
			os.Exit(0)
		case <-ticker.C:
			ctx := zero.GetBot(config.Bot.ID)
			if ctx == nil {
				logrus.Warnf("[%s][%s] %s", plugin, "zero.GetBot", fmt.Sprintf("cannot get bot: %v", config.Bot.ID))
				continue
			}
			CommandWapper(ctx, loginfo, plugin, fn, func() error {
				return f(ctx)
			})
		}
	}
}
