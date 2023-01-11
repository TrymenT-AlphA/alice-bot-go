package main

import (
	// plugin
	_ "alice-bot-go/src/plugin/bilibili"
	_ "alice-bot-go/src/plugin/github"
	_ "alice-bot-go/src/plugin/meta"
	_ "alice-bot-go/src/plugin/netease"
	"os"
	"path/filepath"

	// import
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
)

func init() {
	// config logger
	logrus.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[%time%][%lvl%]%msg% \n",
	})
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {

	zero.RunAndBlock(&zero.Config{
		NickName:      []string{"兔兔"},
		CommandPrefix: "",
		SuperUsers:    []int64{1302393176},
		Driver: []zero.Driver{
			driver.NewWebSocketClient("ws://127.0.0.1:6700", ""),
		},
	}, func() { // preBlock func
		cwd, err := os.Getwd()
		if err != nil {
			logrus.Fatalf("[PreBlock] %s", err)
		}

		cacheDir := filepath.Join(cwd, "..", "data", "cache")
		err = os.MkdirAll(cacheDir, 0666)
		if err != nil {
			logrus.Fatalf("[PreBlock] %s", err)
		}

		databaseDir := filepath.Join(cwd, "..", "data", "database")
		err = os.MkdirAll(databaseDir, 0666)
		if err != nil {
			logrus.Fatalf("[PreBlock] %s", err)
		}
		//	now we should have `data/cache` and `data/database`
	})
}
