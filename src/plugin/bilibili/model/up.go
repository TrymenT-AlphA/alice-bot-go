package model

import (
	"bot-go/src/types"
	"github.com/tidwall/gjson"
)

type Up struct {
	UID     int64 `gorm:"primarykey"`
	NicName string
}

func (up *Up) GetLatestDynamic() (*Dynamic, error) {
	bilibiliAPI, err := types.NewAPI(
		"bilibili",
		"user",
		"info.dynamic",
	)
	if err != nil {
		return nil, err
	}

	bilibiliAPI.Params = make(map[string]interface{})
	bilibiliAPI.Params["host_uid"] = up.UID
	bilibiliAPI.Params["offset_dynamic_id"] = 0
	bilibiliAPI.Params["need_top"] = false

	data, err := bilibiliAPI.Request()
	if err != nil {
		return nil, err
	}

	card := gjson.GetBytes(data, "data.cards.0.card").String()
	dynamic := &Dynamic{
		Description: gjson.Get(card, "item.description").String(),
		Pictures:    nil,
		Timestamp:   gjson.GetBytes(data, "data.cards.0.desc.timestamp").Int(),
	}
	gjson.Get(card, "item.pictures.#.img_src").ForEach(
		func(key, value gjson.Result) bool {
			dynamic.Pictures = append(dynamic.Pictures, value.String())
			return true
		},
	)

	return dynamic, nil
}
