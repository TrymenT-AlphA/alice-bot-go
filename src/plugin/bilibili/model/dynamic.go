package model

import (
	"fmt"

	"github.com/tidwall/gjson"
)

// Dynamic Type==1 forward Type==2 text+image Type==8 video
type Dynamic struct {
	Description string
	Pictures    []string
	Timestamp   int64
	Origin      *Dynamic
}

func NewDynamic(card string, t int64, timestamp int64) *Dynamic {
	var dynamic *Dynamic
	switch t {
	case 1: // forward
		dynamic = &Dynamic{
			Description: gjson.Get(card, "item.content").String(),
			Pictures:    nil,
			Timestamp:   timestamp,
			Origin: NewDynamic(
				gjson.Get(card, "origin").String(),
				gjson.Get(card, "item.orig_type").Int(),
				0,
			),
		}
	case 2: // text + image
		dynamic = &Dynamic{
			Description: gjson.Get(card, "item.description").String(),
			Timestamp:   timestamp,
			Origin:      nil,
		}
		gjson.Get(card, "item.pictures.#.img_src").ForEach(
			func(key, value gjson.Result) bool {
				dynamic.Pictures = append(dynamic.Pictures, value.String())
				return true
			},
		)
	case 8: // video
		dynamic = &Dynamic{
			Description: fmt.Sprintf(
				"%s\n➥%s",
				gjson.Get(card, "title").String(),
				gjson.Get(card, "short_link").String(),
			),
			Pictures:  []string{gjson.Get(card, "pic").String()},
			Timestamp: timestamp,
			Origin:    nil,
		}
	default:
		dynamic = nil
	}
	return dynamic
}
