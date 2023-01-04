package model

import (
	"bot-go/src/model"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
)

type User struct {
	Uid       uint64 `gorm:"primarykey"`
	Group     uint64 `gorm:"primarykey"`
	Timestamp uint64 `gorm:"default:0"`
}

func (User) TableName() string {
	return "User"
}

func (user *User) Create(db *gorm.DB) error {
	if err := db.Create(&User{
		Uid:   user.Uid,
		Group: user.Group,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (user *User) Read(db *gorm.DB) error {
	if err := db.Where(&User{
		Uid:   user.Uid,
		Group: user.Group,
	}).First(user).Error; err != nil {
		return err
	}
	return nil
}

func (user *User) Update(db *gorm.DB) error {
	if err := db.Where(&User{
		Uid:   user.Uid,
		Group: user.Group,
	}).Updates(&User{
		Timestamp: user.Timestamp,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (user *User) Delete(db *gorm.DB) error {
	if err := db.Delete(&User{
		Uid:   user.Uid,
		Group: user.Group,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (User) ReadAll(db *gorm.DB) ([]User, error) {
	var result []User
	if err := db.Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (user *User) GetLatestDynamic() (uint64, string, []string, error) {
	var timestamp uint64
	var description string
	var pictures []string

	var userInfoDynamicAPI model.API

	if err := userInfoDynamicAPI.GetAPI("user", "info.dynamic"); err != nil {
		return timestamp, description, pictures, err
	}

	userInfoDynamicAPI.Params = make(map[string]interface{})
	userInfoDynamicAPI.Params["host_uid"] = user.Uid
	userInfoDynamicAPI.Params["offset_dynamic_id"] = 0
	userInfoDynamicAPI.Params["need_top"] = false

	data, err := userInfoDynamicAPI.Request()
	if err != nil {
		return timestamp, description, pictures, err
	}

	timestamp = gjson.GetBytes(data, "data.cards.0.desc.timestamp").Uint()
	card := gjson.GetBytes(data, "data.cards.0.card").String()
	description = gjson.Get(card, "item.description").String()
	gjson.Get(card, "item.pictures.#.img_src").
		ForEach(func(key, value gjson.Result) bool {
			pictures = append(pictures, value.String())
			return true
		})

	return timestamp, description, pictures, nil
}
