package model

import (
	"alice-bot-go/src/types"
	"github.com/tidwall/gjson"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"net/http"
	"net/http/cookiejar"
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
	err := db.Where(account).Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (account *Account) Create(db *gorm.DB) error {
	err := db.Create(account).Error
	if err != nil {
		return err
	}
	return nil
}

func (account *Account) Read(db *gorm.DB) error {
	err := db.Where(&Account{
		UserID:   account.UserID,
		NickName: account.NickName,
	}).First(account).Error
	if err != nil {
		return err
	}
	return nil
}

func (account *Account) Update(db *gorm.DB) error {
	err := db.Where(&Account{
		UserID:   account.UserID,
		NickName: account.NickName,
	}).Updates(account).Error
	if err != nil {
		return err
	}
	return nil
}

func (account *Account) Delete(db *gorm.DB) error {
	err := db.Delete(&Account{
		UserID:   account.UserID,
		NickName: account.NickName,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (account *Account) CreateOrUpdate(db *gorm.DB) error {
	err := db.Clauses(
		clause.OnConflict{
			UpdateAll: true,
		}).Create(account).Error
	if err != nil {
		return err
	}
	return nil
}

func (account *Account) Login() (*Session, error) {
	neteaseAPI, err := types.NewAPI(
		"netease",
		"login",
		"cellphone",
	)
	if err != nil {
		return nil, err
	}

	neteaseAPI.Params["phone"] = account.Phone
	neteaseAPI.Params["password"] = account.Password

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

	data, err := neteaseAPI.DoRequest(session.Client)
	if err != nil {
		return nil, err
	}

	session.UID = gjson.GetBytes(data, "account.id").Int()

	return session, nil
}
