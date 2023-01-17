package netease

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"

	"alice-bot-go/src/core/alice"
	"alice-bot-go/src/core/config"
	"alice-bot-go/src/core/database/localSqlite3"
	"alice-bot-go/src/plugin/netease/model"
)

// Context guess music context, each one for each group
type Context struct {
	session     *model.Session
	playlist    []model.Play
	currPlay    *model.Play
	tracklist   []model.Track
	currTrack   *model.Track
	nextSegment int
	segmentLen  int
}

var (
	plugin   = "netease"
	db       *gorm.DB
	cache    string
	contexts = make(map[int64]*Context)
)

func init() {
	alice.Init.Register(func() {
		fn := "initialize"
		alice.CommandWapper(nil, true, plugin, fn, func() error {
			return initialize()
		})
	}, 1)
	// usage: 注册 <Account.NickName> <Account.Phone> <Account.Password>
	// example: 注册 AliceRemake <Account.Phone> <Account.Password>
	zero.OnRegex(`^注册 (.+) (.+) (.+)$`, zero.OnlyToMe, zero.OnlyPrivate).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "register"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return register(ctx, db)
		})
	})
	// usage: 注销 <Account.NickName>
	// example: 注销 AliceRemake
	zero.OnRegex(`^注销 (.+)$`, zero.OnlyToMe, zero.OnlyPrivate).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "revoke"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return revoke(ctx, db)
		})
	})
	// usage/example: 账号列表
	zero.OnRegex(`^账号列表$`, zero.OnlyToMe, zero.OnlyPrivate).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "accountList"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return accountList(ctx, db)
		})
	})
	// usage: <NickName> 登录 <Account.NickName>
	// example: 兔兔 登录 AliceRemake
	zero.OnRegex(`^登录 (.+)$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "login"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return login(ctx, db)
		})
	})
	// usage: <NickName> 登出
	// example: 兔兔 登出
	zero.OnRegex(`^登出$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "logout"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return logout(ctx)
		})
	})
	// usage: <NickName> 登录信息
	// example: 兔兔 登录信息
	zero.OnRegex(`^登录信息$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "loginStatus"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return loginStatus(ctx)
		})
	})
	// usage: <NickName> 歌单列表
	// example: 兔兔 歌单列表
	zero.OnRegex(`^歌单列表$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "playList"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return playList(ctx)
		})
	})
	// usage: <NickName> 切换歌单  <Play.Name>
	// example: 兔兔 切换歌单  测试歌单
	zero.OnRegex(`^切换歌单 (.+)$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "switchPlay"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return switchPlay(ctx)
		})
	})
	// usage: <NickName> 歌单信息
	// example: 兔兔 歌单信息
	zero.OnRegex(`^歌曲列表$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "trackList"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return trackList(ctx)
		})
	})
	// usage: <NickName> 开始猜歌
	// example: 兔兔 开始猜歌
	zero.OnRegex(`^开始猜歌$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "startGuess"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return startGuess(ctx)
		})
	})
	// usage: <NickName> 提示
	// example: 兔兔 提示
	zero.OnRegex(`^提示$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "tip"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return tip(ctx)
		})
	})
	// usage: <NickName> 答案
	// example: 兔兔 答案
	zero.OnRegex(`^答案$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "answer"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return answer(ctx)
		})
	})
	// usage: <NickName> 猜 <Track.Name/Track.Tns[i]>
	// example: 兔兔 猜 离去之原
	zero.OnRegex(`^猜 (.+)$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		fn := "guess"
		alice.CommandWapper(ctx, true, plugin, fn, func() error {
			return guess(ctx)
		})
	})
}

func initialize() error {
	cache = filepath.Join(config.Global.CacheDir, plugin)
	if err := os.MkdirAll(cache, 0666); err != nil {
		return err
	}
	var err error
	db, err = localSqlite3.Init(fmt.Sprintf("%s.db", plugin))
	if err != nil {
		return err
	}
	if err := db.AutoMigrate(&model.Account{}); err != nil {
		return err
	}
	return nil
}

func register(ctx *zero.Ctx, db *gorm.DB) error {
	args := ctx.State["regex_matched"].([]string)
	account := model.Account{
		UserID:   ctx.Event.UserID,
		NickName: args[1],
		Phone:    args[2],
		Password: args[3],
	}
	if err := account.CreateOrUpdate(db); err != nil {
		return err
	}
	ctx.Send(message.Text("注册成功"))
	return nil
}

func revoke(ctx *zero.Ctx, db *gorm.DB) error {
	args := ctx.State["regex_matched"].([]string)
	account := model.Account{
		UserID:   ctx.Event.UserID,
		NickName: args[1],
	}
	if err := account.Delete(db); err != nil {
		return err
	}
	ctx.Send(message.Text("注销成功"))
	return nil
}

func accountList(ctx *zero.Ctx, db *gorm.DB) error {
	accounts, err := (&model.Account{
		UserID: ctx.Event.UserID,
	}).ReadAll(db)
	if err != nil {
		return err
	}
	var msg []message.MessageSegment
	msg = append(msg, message.Text("#### 账号列表 ####"))
	for index, account := range accounts {
		msg = append(msg, message.Text(fmt.Sprintf("\n[%v] %s", index+1, account.NickName)))
	}
	ctx.Send((message.Message)(msg))
	return nil
}

func login(ctx *zero.Ctx, db *gorm.DB) error {
	args := ctx.State["regex_matched"].([]string)
	account := model.Account{
		UserID:   ctx.Event.UserID,
		NickName: args[1],
	}
	if err := account.Read(db); err != nil {
		return err
	}
	session, err := account.Login()
	if err != nil {
		return err
	}
	playlist, err := session.GetPlayList()
	if err != nil {
		return err
	}
	contexts[ctx.Event.GroupID] = &Context{
		session:     session,
		playlist:    playlist,
		currPlay:    nil,
		tracklist:   nil,
		currTrack:   nil,
		nextSegment: 1,
		segmentLen:  3,
	}
	ctx.Send(message.Text("登录成功"))
	return nil
}

func logout(ctx *zero.Ctx) error {
	contexts[ctx.Event.GroupID] = nil
	ctx.Send(message.Text("已登出"))
	return nil
}

func loginStatus(ctx *zero.Ctx) error {
	if contexts[ctx.Event.GroupID] == nil {
		ctx.Send(message.Text("暂未登录"))
		return nil
	}
	ctx.Send(message.Text(
		fmt.Sprintf("已登陆 %s", contexts[ctx.Event.GroupID].session.NickName),
	))
	return nil
}

func playList(ctx *zero.Ctx) error {
	if contexts[ctx.Event.GroupID] == nil {
		ctx.Send(message.Text("请先登录"))
		return nil
	}
	var msg []message.MessageSegment
	msg = append(msg, message.Text("#### 歌单列表 ####"))
	for index, play := range contexts[ctx.Event.GroupID].playlist {
		msg = append(msg, message.Text(fmt.Sprintf("\n[%v] %s", index+1, play.Name)))
	}
	ctx.Send((message.Message)(msg))
	return nil
}

func switchPlay(ctx *zero.Ctx) error {
	if contexts[ctx.Event.GroupID] == nil {
		ctx.Send(message.Text("请先登录"))
		return nil
	}
	args := ctx.State["regex_matched"].([]string)
	for _, play := range contexts[ctx.Event.GroupID].playlist {
		if play.Name == args[1] {
			contexts[ctx.Event.GroupID].currPlay = &play
			break
		}
	}
	var err error
	contexts[ctx.Event.GroupID].tracklist, err = contexts[ctx.Event.GroupID].session.GetTrackList(contexts[ctx.Event.GroupID].currPlay)
	if err != nil {
		return err
	}
	ctx.Send(fmt.Sprintf("切换成功"))
	return nil
}

func trackList(ctx *zero.Ctx) error {
	if contexts[ctx.Event.GroupID] == nil {
		ctx.Send(message.Text("请先登录"))
		return nil
	}
	if contexts[ctx.Event.GroupID].currPlay == nil {
		ctx.Send(message.Text("暂未选择歌单"))
		return nil
	}
	var msg []message.MessageSegment
	msg = append(msg, message.Text("### 歌曲列表 ###"))
	for index, track := range contexts[ctx.Event.GroupID].tracklist {
		msg = append(msg, message.Text(fmt.Sprintf("\n[%v] %s", index+1, track.Name)))
	}
	ctx.Send((message.Message)(msg))
	return nil
}

func startGuess(ctx *zero.Ctx) error {
	if contexts[ctx.Event.GroupID] == nil {
		ctx.Send(message.Text("请先登录"))
		return nil
	}
	if contexts[ctx.Event.GroupID].tracklist == nil {
		ctx.Send(message.Text("请先选择歌单"))
		return nil
	}
	contexts[ctx.Event.GroupID].nextSegment = 1
	dice := rand.Intn(len(contexts[ctx.Event.GroupID].tracklist))
	contexts[ctx.Event.GroupID].currTrack = &contexts[ctx.Event.GroupID].tracklist[dice]
	tracklistDir := filepath.Join(cache, "tracklist")
	if err := os.MkdirAll(tracklistDir, 0666); err != nil {
		return err
	}
	guessingDir := filepath.Join(cache, "guessing", fmt.Sprintf("%v", ctx.Event.GroupID))
	if err := os.MkdirAll(guessingDir, 0666); err != nil {
		return err
	}
	task, err := contexts[ctx.Event.GroupID].session.GetTask(contexts[ctx.Event.GroupID].currTrack)
	if err != nil {
		return err
	}
	_, err = os.Stat(filepath.Join(tracklistDir, fmt.Sprintf("%s.%s", task.Name, task.Type)))
	if os.IsNotExist(err) {
		err = contexts[ctx.Event.GroupID].session.DownloadTask(task, tracklistDir)
		if err != nil {
			return err
		}
	}
	var cmdout bytes.Buffer
	var cmderr bytes.Buffer
	cmd := exec.Command(
		"ffprobe",
		"-show_format",
		filepath.Join(tracklistDir, fmt.Sprintf("%s.%s", task.Name, task.Type)),
	)
	cmd.Stdout = &cmdout
	cmd.Stderr = &cmderr
	err = cmd.Run()
	if err != nil {
		return err
	}
	re := regexp.MustCompile("duration=([0-9]+)")
	duration, err := strconv.Atoi((string)(re.FindSubmatch(cmdout.Bytes())[1]))
	if err != nil {
		return err
	}
	segment := duration / 3
	dice = rand.Intn(segment - contexts[ctx.Event.GroupID].segmentLen)
	sp1 := fmt.Sprintf("%v", dice)
	dice = rand.Intn(segment - contexts[ctx.Event.GroupID].segmentLen)
	sp2 := fmt.Sprintf("%v", dice+segment)
	dice = rand.Intn(segment - contexts[ctx.Event.GroupID].segmentLen)
	sp3 := fmt.Sprintf("%v", dice+2*segment)
	cmd = exec.Command(
		"ffmpeg",
		"-y", "-i",
		filepath.Join(tracklistDir, fmt.Sprintf("%s.%s", task.Name, task.Type)),
		"-ss", sp1, "-t", fmt.Sprintf("%v", contexts[ctx.Event.GroupID].segmentLen), filepath.Join(guessingDir, "1.mp3"),
		"-ss", sp2, "-t", fmt.Sprintf("%v", contexts[ctx.Event.GroupID].segmentLen), filepath.Join(guessingDir, "2.mp3"),
		"-ss", sp3, "-t", fmt.Sprintf("%v", contexts[ctx.Event.GroupID].segmentLen), filepath.Join(guessingDir, "3.mp3"),
	)
	err = cmd.Run()
	if err != nil {
		return err
	}
	ctx.Send(message.Record(
		fmt.Sprintf("file:///%s", filepath.Join(guessingDir, fmt.Sprintf("%v.mp3", contexts[ctx.Event.GroupID].nextSegment))),
	))
	contexts[ctx.Event.GroupID].nextSegment++
	return nil
}

func tip(ctx *zero.Ctx) error {
	if contexts[ctx.Event.GroupID] == nil {
		ctx.Send(message.Text("请先登录"))
		return nil
	}
	if contexts[ctx.Event.GroupID].currTrack == nil {
		ctx.Send(message.Text("请先开始猜歌"))
		return nil
	}
	if contexts[ctx.Event.GroupID].nextSegment > 3 {
		ctx.Send(message.Text("提示次数已用完"))
		return nil
	}
	guessingDir := filepath.Join(cache, "guessing", fmt.Sprintf("%v", ctx.Event.GroupID))
	ctx.Send(message.Text(fmt.Sprintf("剩余提示次数：%v", 3-contexts[ctx.Event.GroupID].nextSegment)))
	ctx.Send(message.Record(
		fmt.Sprintf("file:///%s", filepath.Join(guessingDir, fmt.Sprintf("%v.mp3", contexts[ctx.Event.GroupID].nextSegment))),
	))
	contexts[ctx.Event.GroupID].nextSegment++
	return nil
}

func answer(ctx *zero.Ctx) error {
	if contexts[ctx.Event.GroupID] == nil {
		ctx.Send(message.Text("请先登录"))
		return nil
	}
	if contexts[ctx.Event.GroupID].currTrack == nil {
		ctx.Send(message.Text("请先开始猜歌"))
		return nil
	}
	ctx.Send(message.Text(fmt.Sprintf("答案是 %s", contexts[ctx.Event.GroupID].currTrack.Name)))
	contexts[ctx.Event.GroupID].currTrack = nil
	contexts[ctx.Event.GroupID].nextSegment = 1
	return nil
}

func guess(ctx *zero.Ctx) error {
	if contexts[ctx.Event.GroupID] == nil {
		ctx.Send(message.Text("请先登录"))
		return nil
	}
	if contexts[ctx.Event.GroupID].currTrack == nil {
		ctx.Send(message.Text("请先开始猜歌"))
		return nil
	}
	args := ctx.State["regex_matched"].([]string)
	if guessCheck(args[1], contexts[ctx.Event.GroupID].currTrack.Name) {
		ctx.Send("おめでとう")
		contexts[ctx.Event.GroupID].currTrack = nil
		contexts[ctx.Event.GroupID].nextSegment = 1
		return nil
	}
	for _, tn := range contexts[ctx.Event.GroupID].currTrack.Tns {
		if guessCheck(args[1], tn) {
			ctx.Send("おめでとう")
			contexts[ctx.Event.GroupID].currTrack = nil
			contexts[ctx.Event.GroupID].nextSegment = 1
			return nil
		}
	}
	ctx.Send(message.Text("残念"))
	return nil
}

func guessCheck(guess string, answer string) bool {
	newGuess := strings.ToLower(guess)
	newAnswer := strings.ToLower(answer)
	guessLen := len(newGuess)
	answerLen := len(newAnswer)
	if guessLen > answerLen/2 && strings.Contains(newAnswer, newGuess) {
		return true
	}
	return false
}
