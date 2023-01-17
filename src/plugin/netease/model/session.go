package model

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/tidwall/gjson"

	"alice-bot-go/src/core/alice"
)

type Session struct {
	NickName string
	UID      int64
	Client   *http.Client
}

func (session *Session) GetPlayList() ([]Play, error) {
	api, err := alice.NewAPI(
		"netease",
		"user",
		"playlist",
	)
	if err != nil {
		return nil, err
	}
	api.Params = map[string]interface{}{
		"uid": session.UID,
	}
	data, err := api.DoRequest(session.Client)
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
	api, err := alice.NewAPI(
		"netease",
		"playlist",
		"detail",
	)
	if err != nil {
		return nil, err
	}
	api.Params = map[string]interface{}{
		"id": play.PID,
	}
	data, err := api.DoRequest(session.Client)
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
	api, err := alice.NewAPI(
		"netease",
		"song",
		"url",
	)
	if err != nil {
		return nil, err
	}
	api.Params = map[string]interface{}{
		"id": track.TID,
	}
	data, err := api.DoRequest(session.Client)
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
	if err := os.MkdirAll(dir, 0666); err != nil {
		return err
	}
	if err := task.Download(dir); err != nil {
		return err
	}
	return nil
}

func (session *Session) DownloadTrack(track *Track, dir string) error {
	api, err := alice.NewAPI(
		"netease",
		"song",
		"url",
	)
	if err != nil {
		return err
	}
	api.Params = map[string]interface{}{
		"id": track.TID,
		"br": 128000,
	}
	data, err := api.DoRequest(session.Client)
	if err != nil {
		return err
	}
	task := &Task{
		Name: track.Name,
		Url:  gjson.GetBytes(data, "data.0.url").String(),
		Type: gjson.GetBytes(data, "data.0.type").String(),
	}
	if err = os.MkdirAll(dir, 0666); err != nil {
		return err
	}
	if err = task.Download(dir); err != nil {
		return err
	}
	return nil
}

func (session *Session) DownloadTrackList(tracklist []Track, dir string) error {
	if err := os.MkdirAll(dir, 0666); err != nil {
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
	if err = os.MkdirAll(dir, 0666); err != nil {
		return err
	}
	if err = session.DownloadTrackList(tracklist, filepath.Join(dir, play.Name)); err != nil {
		return err
	}
	return nil
}
