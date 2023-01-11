package netease

import (
	"alice-bot-go/src/database/localSqlite3"
	"alice-bot-go/src/plugin/netease/model"
	"alice-bot-go/src/util"
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
	"gorm.io/gorm"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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
	db       *gorm.DB
	cache    string
	contexts map[int64]*Context
)

func init() {
	// Initialize
	err := Initialize()
	if err != nil {
		logrus.Fatalf("[netease][Initialize] %s", err)
	} else {
		logrus.Infof("[netease][Initialize][success]")
	}

	// OnlyPrivate Context Free
	// example: 注册 AliceRemake <Phone> <Password>
	zero.OnRegex(`^注册 (.+) (.+) (.+)$`, zero.OnlyToMe, zero.OnlyPrivate).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := Register(ctx)
		if err != nil {
			logrus.Errorf("[netease][Register] %s", err)
			ctx.Send(message.Text(fmt.Sprintf("[netease][Register] %s", err)))
		} else {
			logrus.Infof("[netease][Register][success]")
		}
	})
	// example: 注销 AliceRemake
	zero.OnRegex(`^注销 (.+)$`, zero.OnlyToMe, zero.OnlyPrivate).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := Revoke(ctx)
		if err != nil {
			logrus.Errorf("[netease][Revoke] %s", err)
			ctx.Send(message.Text(fmt.Sprintf("[netease][Revoke] %s", err)))
		} else {
			logrus.Infof("[netease][Revoke][success]")
		}
	})
	// example: 账号列表
	zero.OnRegex(`^账号列表$`, zero.OnlyToMe, zero.OnlyPrivate).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := AccountList(ctx)
		if err != nil {
			logrus.Errorf("[netease][AccountList] %s", err)
			ctx.Send(message.Text(fmt.Sprintf("[netease][AccountList] %s", err)))
		} else {
			logrus.Infof("[netease][AccountList][success]")
		}
	})

	// OnlyGroup Create a new context for a group after login
	// in the new context, we should have `session` and `playlist`
	// example: 兔兔 登录 AliceRemake
	zero.OnRegex(`^登录 (.+)$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := Login(ctx)
		if err != nil {
			logrus.Errorf("[netease][Login] %s", err)
			ctx.Send(message.Text(fmt.Sprintf("[netease][Login] %s", err)))
		} else {
			logrus.Infof("[netease][Login][success]")
		}
	})
	// example: 兔兔 登出
	zero.OnRegex(`^登出$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		Logout(ctx)
		logrus.Infof("[netease][Logout][success]")
	})
	// example: 兔兔 登录信息
	zero.OnRegex(`^登录信息$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		LoginStatus(ctx)
		logrus.Infof("[netease][LoginStatus][success]")
	})
	// example: 兔兔 歌单列表
	zero.OnRegex(`^歌单列表$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		PlayList(ctx)
		logrus.Infof("[netease][PlayList][success]")
	})

	// OnlyGroup Use the context created in login
	// after `SwitchPlay`, we should have `currPlay` and `tracklist` in the context
	// example: 兔兔 切换歌单  测试歌单
	zero.OnRegex(`^切换歌单 (.+)$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := SwitchPlay(ctx)
		if err != nil {
			logrus.Errorf("[netease][SwitchPlay] %s", err)
			ctx.Send(message.Text(fmt.Sprintf("[netease][SwitchPlay] %s", err)))
		} else {
			logrus.Infof("[netease][SwitchPlay][success]")
		}
	})
	// example: 兔兔 歌单信息
	zero.OnRegex(`^歌曲列表$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		TrackList(ctx)
		logrus.Infof("[netease][PlayStatus][success]")
	})

	// OnlyGroup Use the context created in login
	// after `StartGuess`, we should have `currTrack`
	// example: 兔兔 开始猜歌
	zero.OnRegex(`^开始猜歌$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := StartGuess(ctx)
		if err != nil {
			logrus.Errorf("[netease][StartGuess] %s", err)
			ctx.Send(message.Text(fmt.Sprintf("[netease][StartGuess] %s", err)))
		} else {
			logrus.Infof("[netease][StartGuess][success]")
		}
	})
	// example: 兔兔 提示
	zero.OnRegex(`^提示$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := Tip(ctx)
		if err != nil {
			logrus.Errorf("[netease][Tip] %s", err)
			ctx.Send(message.Text(fmt.Sprintf("[netease][Tip] %s", err)))
		} else {
			logrus.Infof("[netease][Tip][success]")
		}
	})
	// example: 兔兔 答案
	zero.OnRegex(`^答案$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := Answer(ctx)
		if err != nil {
			logrus.Errorf("[netease][Answer] %s", err)
			ctx.Send(message.Text(fmt.Sprintf("[netease][Answer] %s", err)))
		} else {
			logrus.Infof("[netease][Answer][success]")
		}
	})
	// example: 兔兔 猜 离去之原
	zero.OnRegex(`^猜 (.+)$`, zero.OnlyToMe, zero.OnlyGroup).SetBlock(true).Handle(func(ctx *zero.Ctx) {
		err := Guess(ctx)
		if err != nil {
			logrus.Errorf("[netease][Guess] %s", err)
			ctx.Send(message.Text(fmt.Sprintf("[netease][Guess] %s", err)))
		} else {
			logrus.Infof("[netease][Guess][success]")
		}
	})
}

func Initialize() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cache = filepath.Join(cwd, "..", "data", "cache", "netease")
	err = os.MkdirAll(cache, 0666)
	if err != nil {
		return err
	}

	db, err = localSqlite3.Init("netease.db")
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&model.Account{})
	if err != nil {
		return err
	}

	contexts = make(map[int64]*Context)

	return nil
}

// Register example: 注册 AliceBot 18852000505 dongwoo1217
func Register(ctx *zero.Ctx) error {
	args := ctx.State["regex_matched"].([]string)

	account := model.Account{
		UserID:   ctx.Event.UserID,
		NickName: args[1],
		Phone:    args[2],
		Password: args[3],
	}

	err := account.CreateOrUpdate(db)
	if err != nil {
		return err
	}

	ctx.Send(message.Text("注册成功"))

	return nil
}

// Revoke example: 注销 AliceBot
func Revoke(ctx *zero.Ctx) error {
	args := ctx.State["regex_matched"].([]string)

	account := model.Account{
		UserID:   ctx.Event.UserID,
		NickName: args[1],
	}

	err := account.Delete(db)
	if err != nil {
		return err
	}

	ctx.Send(message.Text("注销成功"))

	return nil
}

// AccountList example: 账号列表
func AccountList(ctx *zero.Ctx) error {
	accounts, err := (&model.Account{
		UserID: ctx.Event.UserID,
	}).ReadAll(db)
	if err != nil {
		return err
	}

	var msg []message.MessageSegment
	msg = append(msg, message.Text("### 账号列表 ###"))
	for index, account := range accounts {
		msg = append(msg, message.Text(fmt.Sprintf("\n[%v] %s", index+1, account.NickName)))
	}

	ctx.Send((message.Message)(msg))

	return nil
}

// Login example: 兔兔 登录 AliceRemake
func Login(ctx *zero.Ctx) error {
	args := ctx.State["regex_matched"].([]string)

	account := model.Account{
		UserID:   ctx.Event.UserID,
		NickName: args[1],
	}

	err := account.Read(db)
	if err != nil {
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

// Logout example: 兔兔 登出
func Logout(ctx *zero.Ctx) {
	contexts[ctx.Event.GroupID] = nil
	ctx.Send(message.Text("已登出"))
}

// LoginStatus example: 兔兔 登录信息
func LoginStatus(ctx *zero.Ctx) {
	if contexts[ctx.Event.GroupID] == nil {
		ctx.Send(message.Text("暂未登录"))
		return
	}

	ctx.Send(message.Text(
		fmt.Sprintf("已登陆 %s", contexts[ctx.Event.GroupID].session.NickName),
	))
}

// PlayList example: 兔兔 歌单列表
func PlayList(ctx *zero.Ctx) {
	if contexts[ctx.Event.GroupID] == nil {
		ctx.Send(message.Text("请先登录"))
		return
	}

	var msg []message.MessageSegment
	msg = append(msg, message.Text("### 歌单列表 ###"))
	for index, play := range contexts[ctx.Event.GroupID].playlist {
		msg = append(msg, message.Text(fmt.Sprintf("\n[%v] %s", index+1, play.Name)))
	}

	ctx.Send((message.Message)(msg))
}

// SwitchPlay example: 兔兔 切换歌单 AliceBot
func SwitchPlay(ctx *zero.Ctx) error {
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

// TrackList example: 兔兔 歌曲列表
func TrackList(ctx *zero.Ctx) {
	if contexts[ctx.Event.GroupID] == nil {
		ctx.Send(message.Text("请先登录"))
		return
	}

	if contexts[ctx.Event.GroupID].currPlay == nil {
		ctx.Send(message.Text("暂未选择歌单"))
		return
	}

	var msg []message.MessageSegment
	msg = append(msg, message.Text("### 歌曲列表 ###"))
	for index, track := range contexts[ctx.Event.GroupID].tracklist {
		msg = append(msg, message.Text(fmt.Sprintf("\n[%v] %s", index+1, track.Name)))
	}

	ctx.Send((message.Message)(msg))
}

// StartGuess example: 兔兔 开始猜歌
func StartGuess(ctx *zero.Ctx) error {
	if contexts[ctx.Event.GroupID] == nil {
		ctx.Send(message.Text("请先登录"))
		return nil
	}

	if contexts[ctx.Event.GroupID].tracklist == nil {
		ctx.Send(message.Text("请先选择歌单"))
		return nil
	}

	contexts[ctx.Event.GroupID].nextSegment = 1

	dice := util.GetDice(len(contexts[ctx.Event.GroupID].tracklist))

	contexts[ctx.Event.GroupID].currTrack = &contexts[ctx.Event.GroupID].tracklist[dice]

	tracklistDir := filepath.Join(cache, "tracklist")
	err := os.MkdirAll(tracklistDir, 0666)
	if err != nil {
		return err
	}

	guessingDir := filepath.Join(cache, "guessing", fmt.Sprintf("%v", ctx.Event.GroupID))
	err = os.MkdirAll(guessingDir, 0666)
	if err != nil {
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

	dice = util.GetDice(segment - contexts[ctx.Event.GroupID].segmentLen)
	sp1 := fmt.Sprintf("%v", dice)
	dice = util.GetDice(segment - contexts[ctx.Event.GroupID].segmentLen)
	sp2 := fmt.Sprintf("%v", dice+segment)
	dice = util.GetDice(segment - contexts[ctx.Event.GroupID].segmentLen)
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

// Tip example: 兔兔 提示
func Tip(ctx *zero.Ctx) error {
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

// Answer example: 兔兔 答案
func Answer(ctx *zero.Ctx) error {
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

// Guess example: 兔兔 猜 离去之原
func Guess(ctx *zero.Ctx) error {
	if contexts[ctx.Event.GroupID] == nil {
		ctx.Send(message.Text("请先登录"))
		return nil
	}

	if contexts[ctx.Event.GroupID].currTrack == nil {
		ctx.Send(message.Text("请先开始猜歌"))
		return nil
	}

	args := ctx.State["regex_matched"].([]string)

	if GuessCheck(args[1], contexts[ctx.Event.GroupID].currTrack.Name) {
		ctx.Send("おめでとう")
		contexts[ctx.Event.GroupID].currTrack = nil
		contexts[ctx.Event.GroupID].nextSegment = 1
		return nil
	}
	for _, tn := range contexts[ctx.Event.GroupID].currTrack.Tns {
		if GuessCheck(args[1], tn) {
			ctx.Send("おめでとう")
			contexts[ctx.Event.GroupID].currTrack = nil
			contexts[ctx.Event.GroupID].nextSegment = 1
			return nil
		}
	}

	ctx.Send(message.Text("残念"))

	return nil
}

func GuessCheck(guess string, answer string) bool {
	newGuess := strings.ToLower(guess)
	newAnswer := strings.ToLower(answer)
	guessLen := len(newGuess)
	answerLen := len(newAnswer)
	if guessLen > answerLen/2 && strings.Contains(newAnswer, newGuess) {
		return true
	}
	return false
}
