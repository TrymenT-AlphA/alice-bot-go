package model

import (
	"net/http"
	"net/http/cookiejar"

	"github.com/tidwall/gjson"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"alice-bot-go/src/core/alice"
	"alice-bot-go/src/core/config"
)

type Account struct {
	UserID   int64  `gorm:"primarykey"`
	NickName string `gorm:"primarykey"`
	Phone    string
	Password string
}

func (*Account) TableName() string {
	return "Account"
}

func (account *Account) ReadAll(db *gorm.DB) ([]Account, error) {
	var result []Account
	if err := db.Where(account).Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (account *Account) Create(db *gorm.DB) error {
	if err := db.Create(account).Error; err != nil {
		return err
	}
	return nil
}

func (account *Account) Read(db *gorm.DB) error {
	if err := db.Where(&Account{
		UserID:   account.UserID,
		NickName: account.NickName,
	}).First(account).Error; err != nil {
		return err
	}
	return nil
}

func (account *Account) Update(db *gorm.DB) error {
	if err := db.Where(&Account{
		UserID:   account.UserID,
		NickName: account.NickName,
	}).Updates(account).Error; err != nil {
		return err
	}
	return nil
}

func (account *Account) Delete(db *gorm.DB) error {
	if err := db.Delete(&Account{
		UserID:   account.UserID,
		NickName: account.NickName,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (account *Account) CreateOrUpdate(db *gorm.DB) error {
	if err := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(account).Error; err != nil {
		return err
	}
	return nil
}

func (account *Account) Login() (*Session, error) {
	api, err := alice.NewAPI(
		"netease",
		"login",
		"cellphone",
	)
	if err != nil {
		return nil, err
	}
	api.UrlParams = []interface{}{config.Netease.Server}
	api.Params = map[string]interface{}{
		"phone":    account.Phone,
		"password": account.Password,
	}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	session := &Session{
		NickName: account.NickName,
		Client: &http.Client{
			Jar: jar,
		},
	}
	data, err := api.DoRequest(session.Client)
	if err != nil {
		return nil, err
	}
	session.UID = gjson.GetBytes(data, "account.id").Int()
	return session, nil
}
