package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/driver"
	"gopkg.in/yaml.v2"

	"alice-bot-go/src/core/alice"
	"alice-bot-go/src/core/config"
	"alice-bot-go/src/core/util"

	_ "alice-bot-go/src/plugin/bilibili"
	_ "alice-bot-go/src/plugin/github"
	_ "alice-bot-go/src/plugin/heartbeat"
	_ "alice-bot-go/src/plugin/meta"
	_ "alice-bot-go/src/plugin/netease"
)

var (
	initComplete = make(chan bool, 1)
	plugin       = "main"
)

func init() {
	alice.Init.Register(func() {
		fn := "initialize"
		alice.CommandWapper(nil, true, plugin, fn, func() error {
			return initialize(initComplete)
		})
	}, 0)
}

func main() {
	alice.Init.Initialize()
	zero.RunAndBlock(&zero.Config{
		NickName:      config.Bot.NickName,
		CommandPrefix: config.Bot.CommandPrefix,
		SuperUsers:    config.Bot.SuperUsers,
		Driver: func() []zero.Driver {
			var Driver []zero.Driver
			for _, botDriver := range config.Bot.Driver {
				if botDriver.Type == 1 {
					Driver = append(Driver, driver.NewWebSocketClient(botDriver.Url, botDriver.AccessToken))
				}
				if botDriver.Type == 2 {
					Driver = append(Driver, driver.NewWebSocketServer(botDriver.Waitn, botDriver.Url, botDriver.AccessToken))
				}
			}
			return Driver
		}(),
	}, nil)
}

func initialize(initComplete chan bool) error {
	if err := initLogrus(); err != nil {
		return err
	}
	if err := initGlobal(); err != nil {
		return err
	}
	if err := initBot(); err != nil {
		return err
	}
	initComplete <- true
	return nil
}

// initLogrus set logrus format and level
func initLogrus() error {
	logrus.SetFormatter(&easy.Formatter{
		TimestampFormat: "2006-01-02 15:04:05",
		LogFormat:       "[%time%][%lvl%][alice-bot]%msg% \n",
	})
	logrus.SetLevel(logrus.InfoLevel)
	return nil
}

// initGlobal make dirs: `data/cache` `data/config` `data/database` and set config.Global
func initGlobal() error {
	var err error
	config.Global.UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36 Edg/108.0.1462.54"
	config.Global.Cwd, err = os.Getwd()
	if err != nil {
		return err
	}
	config.Global.CacheDir = filepath.Join(config.Global.Cwd, "..", "data", "cache")
	if util.IsNotExist(config.Global.CacheDir) {
		err = os.MkdirAll(config.Global.CacheDir, 0666)
		if err != nil {
			return err
		}
	}
	config.Global.ConfigDir = filepath.Join(config.Global.Cwd, "..", "data", "config")
	if util.IsNotExist(config.Global.ConfigDir) {
		err = os.MkdirAll(config.Global.ConfigDir, 0666)
		if err != nil {
			return err
		}
	}
	config.Global.DatabaseDir = filepath.Join(config.Global.Cwd, "..", "data", "database")
	if util.IsNotExist(config.Global.DatabaseDir) {
		err = os.MkdirAll(config.Global.DatabaseDir, 0666)
		if err != nil {
			return err
		}
	}
	//	now we should have `data/cache` `data/config` `data/database`
	return nil
}

// initBot read config and set config.Bot, if no config file, auto generate
func initBot() error {
	botYml := filepath.Join(config.Global.ConfigDir, "bot.yml")
	if util.IsNotExist(botYml) {
		logrus.Infof("[initBot] no `bot.yml` was found, start configuring...")
		for { // config.Bot.ID
			fmt.Printf("[initBot] enter config.Bot.ID:")
			n, err := fmt.Scan(&config.Bot.ID)
			if n != 1 || err != nil {
				if errors.Is(err, io.EOF) {
					os.Exit(0)
				}
				fmt.Printf("[initBot] input err: %s, try again.\n", err)
				continue
			}
			break
		}
		// config.Bot.NickName
		fmt.Printf("[initBot] configuring config.Bot.NickName(enter `0` to finish):\n")
		for {
			fmt.Printf("[initBot] enter NO.%v NickName:", len(config.Bot.SuperUsers)+1)
			var nickName string
			n, err := fmt.Scan(&nickName)
			if n != 1 || err != nil {
				if errors.Is(err, io.EOF) {
					os.Exit(0)
				}
				fmt.Printf("[initBot] input err: %s, try again.\n", err)
				continue
			}
			if nickName == "0" {
				break
			}
			config.Bot.NickName = append(config.Bot.NickName, nickName)
		}
		for { // config.Bot.CommandPrefix
			fmt.Printf("[initBot] enter config.Bot.CommandPrefix(`empty` for no commandPrefix):")
			n, err := fmt.Scan(&config.Bot.CommandPrefix)
			if n != 1 || err != nil {
				if errors.Is(err, io.EOF) {
					os.Exit(0)
				}
				fmt.Printf("[initBot] input err: %s, try again.\n", err)
				continue
			}
			if config.Bot.CommandPrefix == "empty" {
				config.Bot.CommandPrefix = ""
			}
			break
		}
		// config.Bot.SuperUsers
		fmt.Printf("[initBot] configuring config.Bot.SuperUsers(enter `0` to finish):\n")
		for {
			fmt.Printf("[initBot] enter NO.%v SuperUser:", len(config.Bot.SuperUsers)+1)
			var superUser int64
			n, err := fmt.Scan(&superUser)
			if n != 1 || err != nil {
				if errors.Is(err, io.EOF) {
					os.Exit(0)
				}
				fmt.Printf("[initBot] input err: %s, try again.\n", err)
				continue
			}
			if superUser == 0 {
				break
			}
			config.Bot.SuperUsers = append(config.Bot.SuperUsers, superUser)
		}
		// config.Bot.Dirver
		fmt.Printf("[initBot] configuring config.Bot.Driver(enter `0` to finish):\n")
		for {
			fmt.Printf("[initBot] configuring NO.%v Driver:\n", len(config.Bot.Driver)+1)
			var botDriver config.Driver
			for { // botDriver.Type
				fmt.Printf("[initBot] choose websocket type(`1` for `ws`, `2` for `ws-reverse`):")
				var wsType int
				n, err := fmt.Scan(&wsType)
				if n != 1 || err != nil {
					if errors.Is(err, io.EOF) {
						os.Exit(0)
					}
					fmt.Printf("[initBot] input err: %s, try again.\n", err)
					continue
				}
				if wsType != 1 && wsType != 2 {
					fmt.Printf("[initBot] range err, try again.\n")
					continue
				}
				botDriver.Type = wsType
				break
			}
			if botDriver.Type == 2 { // botDriver.Waitn
				for {
					fmt.Printf("[initBot] using `ws-reverse`, enter botDriver.Waitn`:")
					n, err := fmt.Scan(&botDriver.Waitn)
					if n != 1 || err != nil {
						if errors.Is(err, io.EOF) {
							os.Exit(0)
						}
						fmt.Printf("[initBot] input err: %s, try again.\n", err)
						continue
					}
					break
				}
			}
			for { // botDriver.Url
				fmt.Printf("[initBot] enter botDriver.Url(`default` for default url):")
				n, err := fmt.Scan(&botDriver.Url)
				if n != 1 || err != nil {
					if errors.Is(err, io.EOF) {
						os.Exit(0)
					}
					fmt.Printf("[initBot] input err: %s, try again.\n", err)
					continue
				}
				if botDriver.Url == "default" {
					if botDriver.Type == 1 {
						botDriver.Url = "ws://127.0.0.1:6700"
					}
					if botDriver.Type == 2 {
						botDriver.Url = "ws://127.0.0.1:6701"
					}
				}
				break
			}
			for { // botDriver.AccessToken
				fmt.Printf("[initBot] enter botDriver.AccessToken(`empty for no accessToken`):")
				n, err := fmt.Scan(&botDriver.AccessToken)
				if n != 1 || err != nil {
					if errors.Is(err, io.EOF) {
						os.Exit(0)
					}
					fmt.Printf("[initBot] input err: %s, try again.\n", err)
					continue
				}
				if botDriver.AccessToken == "empty" {
					botDriver.AccessToken = ""
				}
				break
			}
			config.Bot.Driver = append(config.Bot.Driver, botDriver)
			break
		}
		data, err := yaml.Marshal(&config.Bot)
		if err != nil {
			logrus.Fatalf("[initBot] %s", err)
		}
		err = util.Write(botYml, data)
		if err != nil {
			logrus.Fatalf("[initBot] %s", err)
		}
		logrus.Infof("[initBot] configure complete!")
	} else {
		logrus.Infof("[initBot] found `bot.yml`")
		data, err := os.ReadFile(botYml)
		if err != nil {
			logrus.Fatalf("[initBot] %s", err)
		}
		err = yaml.Unmarshal(data, &config.Bot)
		if err != nil {
			logrus.Fatalf("[initBot] %s", err)
		}
	}
	logrus.Infof("[initBot] config.Bot:%+v", config.Bot)
	return nil
}
