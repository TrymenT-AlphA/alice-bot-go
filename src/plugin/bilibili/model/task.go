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

func (*Task) TableName() string {
	return "Task"
}

func (*Task) ReadAll(db *gorm.DB) ([]Task, error) {
	var result []Task
	if err := db.Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (task *Task) Create(db *gorm.DB) error {
	if err := db.Create(task).Error; err != nil {
		return err
	}
	return nil
}

func (task *Task) Read(db *gorm.DB) error {
	if err := db.Where(&Task{
		Up: Up{
			UID: task.Up.UID,
		},
		GroupID: task.GroupID,
	}).First(task).Error; err != nil {
		return err
	}
	return nil
}

func (task *Task) Update(db *gorm.DB) error {
	if err := db.Where(&Task{
		Up: Up{
			UID: task.Up.UID,
		},
		GroupID: task.GroupID,
	}).Updates(task).Error; err != nil {
		return err
	}
	return nil
}

func (task *Task) Delete(db *gorm.DB) error {
	if err := db.Delete(&Task{
		Up: Up{
			UID: task.Up.UID,
		},
		GroupID: task.GroupID,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (task *Task) CreateOrUpdate(db *gorm.DB) error {
	if err := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(task).Error; err != nil {
		return err
	}
	return nil
}
