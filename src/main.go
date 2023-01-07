package main

import (
	// plugin
	_ "bot-go/src/plugin/bilibili"
	_ "bot-go/src/plugin/meta"
	_ "bot-go/src/plugin/netease"
	"os"
	"path/filepath"

	// import
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
)

func init() {
	logrus.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[zerobot][%time%][%lvl%]: %msg% \n",
	})
	logrus.SetLevel(logrus.DebugLevel)
}

func main() {

	cwd, err := os.Getwd()
	if err != nil {
		logrus.Fatalf("[TOP] %s", err)
	}

	cacheDir := filepath.Join(cwd, "..", "data", "cache")
	databaseDir := filepath.Join(cwd, "..", "data", "database")

	err = os.MkdirAll(cacheDir, 0666)
	if err != nil {
		logrus.Fatalf("[TOP] %s", err)
	}
	err = os.MkdirAll(databaseDir, 0666)
	if err != nil {
		logrus.Fatalf("[TOP] %s", err)
	}

	zero.RunAndBlock(&zero.Config{
		NickName:      []string{"兔兔"},
		CommandPrefix: "",
		SuperUsers:    []int64{1302393176},
		Driver: []zero.Driver{
			driver.NewWebSocketClient("ws://127.0.0.1:6700", ""),
		},
	}, nil)
}
