package model

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Task struct {
	Up
	GroupID   int64 `gorm:"primarykey"`
	Timestamp int64
}

func (Task) TableName() string {
	return "Task"
}

func (Task) ReadAll(db *gorm.DB) ([]Task, error) {
	var result []Task
	err := db.Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (user *Task) Create(db *gorm.DB) error {
	err := db.Create(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (user *Task) Read(db *gorm.DB) error {
	err := db.Where(&Task{
		Up: Up{
			UID: user.Up.UID,
		},
		GroupID: user.GroupID,
	}).First(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (user *Task) Update(db *gorm.DB) error {
	err := db.Where(&Task{
		Up: Up{
			UID: user.Up.UID,
		},
		GroupID: user.GroupID,
	}).Updates(user).Error
	if err != nil {
		return err
	}
	return nil
}

func (user *Task) Delete(db *gorm.DB) error {
	err := db.Delete(&Task{
		Up: Up{
			UID: user.Up.UID,
		},
		GroupID: user.GroupID,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (user *Task) CreateOrUpdate(db *gorm.DB) error {
	err := db.Clauses(
		clause.OnConflict{
			UpdateAll: true,
		}).Create(user).Error
	if err != nil {
		return err
	}
	return nil
}
