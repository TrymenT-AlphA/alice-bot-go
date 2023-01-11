package model

import (
	"alice-bot-go/src/types"
	"github.com/tidwall/gjson"
	"net/http"
	"os"
	"path/filepath"
)

type Session struct {
	NickName string
	UID      int64
	Client   *http.Client
}

func (session *Session) GetPlayList() ([]Play, error) {
	neteaseAPI, err := types.NewAPI(
		"netease",
		"user",
		"playlist",
	)
	if err != nil {
		return nil, err
	}

	neteaseAPI.Params["uid"] = session.UID

	data, err := neteaseAPI.DoRequest(session.Client)
	if err != nil {
		return nil, err
	}

	var playlist []Play

	gjson.GetBytes(data, "playlist").ForEach(
		func(key, value gjson.Result) bool {
			play := Play{
				Name: value.Get("name").String(),
				PID:  value.Get("id").Int(),
			}
			playlist = append(playlist, play)
			return true
		})

	return playlist, nil
}

func (session *Session) GetTrackList(play *Play) ([]Track, error) {
	neteaseAPI, err := types.NewAPI(
		"netease",
		"playlist",
		"detail",
	)
	if err != nil {
		return nil, err
	}

	neteaseAPI.Params["id"] = play.PID

	data, err := neteaseAPI.DoRequest(session.Client)
	if err != nil {
		return nil, err
	}

	var tracklist []Track

	gjson.GetBytes(data, "playlist.tracks").ForEach(
		func(key, value gjson.Result) bool {
			var tns []string
			value.Get("tns").ForEach(func(key, value gjson.Result) bool {
				tns = append(tns, value.String())
				return true
			})
			track := Track{
				TID:  value.Get("id").Int(),
				Name: value.Get("name").String(),
				Tns:  tns,
			}
			tracklist = append(tracklist, track)
			return true
		})

	return tracklist, nil
}

func (session *Session) GetTask(track *Track) (*Task, error) {
	neteaseAPI, err := types.NewAPI(
		"netease",
		"song",
		"url",
	)
	if err != nil {
		return nil, err
	}

	neteaseAPI.Params["id"] = track.TID

	data, err := neteaseAPI.DoRequest(session.Client)
	if err != nil {
		return nil, err
	}

	task := &Task{
		Name: track.Name,
		Url:  gjson.GetBytes(data, "data.0.url").String(),
		Type: gjson.GetBytes(data, "data.0.type").String(),
	}

	return task, nil
}

func (session *Session) DownloadTask(task *Task, dir string) error {
	err := os.MkdirAll(dir, 0666)
	if err != nil {
		return err
	}

	err = task.Download(dir)
	if err != nil {
		return err
	}

	return nil
}

func (session *Session) DownloadTrack(track *Track, dir string) error {
	neteaseAPI, err := types.NewAPI(
		"netease",
		"song",
		"url",
	)
	if err != nil {
		return err
	}

	neteaseAPI.Params["id"] = track.TID
	neteaseAPI.Params["br"] = 128000

	data, err := neteaseAPI.DoRequest(session.Client)
	if err != nil {
		return err
	}

	task := &Task{
		Name: track.Name,
		Url:  gjson.GetBytes(data, "data.0.url").String(),
		Type: gjson.GetBytes(data, "data.0.type").String(),
	}

	err = os.MkdirAll(dir, 0666)
	if err != nil {
		return err
	}

	err = task.Download(dir)
	if err != nil {
		return err
	}

	return nil
}

func (session *Session) DownloadTrackList(tracklist []Track, dir string) error {
	err := os.MkdirAll(dir, 0666)
	if err != nil {
		return err
	}

	for _, track := range tracklist {
		err := session.DownloadTrack(&track, dir)
		if err != nil {
			return err
		}
	}

	return nil
}

func (session *Session) DownloadPlay(play *Play, dir string) error {
	tracklist, err := session.GetTrackList(play)
	if err != nil {
		return err
	}

	err = os.MkdirAll(dir, 0666)
	if err != nil {
		return err
	}

	err = session.DownloadTrackList(tracklist, filepath.Join(dir, play.Name))
	if err != nil {
		return err
	}

	return nil
}
