package model

import (
	"net/http"

	"github.com/tidwall/gjson"

	"alice-bot-go/src/core/alice"
)

type Up struct {
	UID      int64 `gorm:"primarykey"`
	NickName string
}

func (up *Up) GetLatestDynamic() (*Dynamic, error) {
	api, err := alice.NewAPI(
		"bilibili",
		"user",
		"info.dynamic",
	)
	if err != nil {
		return nil, err
	}
	api.Params = map[string]interface{}{
		"host_uid":          up.UID,
		"offset_dynamic_id": 0,
		"need_top":          false,
	}
	data, err := api.DoRequest(&http.Client{})
	if err != nil {
		return nil, err
	}
	d := gjson.GetBytes(data, "data.cards.0")
	card := d.Get("card").String()
	dynamic := NewDynamic(
		card,
		d.Get("desc.type").Int(),
		d.Get("desc.timestamp").Int(),
	)
	return dynamic, nil
}
