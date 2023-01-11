package model

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Task struct {
	Repo
	GroupID   int64 `gorm:"primarykey"`
	Timestamp int64
}

func (*Task) TableName() string {
	return "Task"
}

func (*Task) ReadAll(db *gorm.DB) ([]Task, error) {
	var result []Task
	err := db.Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (task *Task) Create(db *gorm.DB) error {
	err := db.Create(task).Error
	if err != nil {
		return err
	}
	return nil
}

func (task *Task) Read(db *gorm.DB) error {
	err := db.Where(&Task{
		Repo: Repo{
			Owner: task.Repo.Owner,
			Name:  task.Repo.Name,
			Local: task.Repo.Local,
		},
		GroupID: task.GroupID,
	}).First(task).Error
	if err != nil {
		return err
	}
	return nil
}

func (task *Task) Update(db *gorm.DB) error {
	err := db.Where(&Task{
		Repo: Repo{
			Owner: task.Repo.Owner,
			Name:  task.Repo.Name,
			Local: task.Repo.Local,
		},
		GroupID: task.GroupID,
	}).Updates(task).Error
	if err != nil {
		return err
	}
	return nil
}

func (task *Task) Delete(db *gorm.DB) error {
	err := db.Delete(&Task{
		Repo: Repo{
			Owner: task.Repo.Owner,
			Name:  task.Repo.Name,
			Local: task.Repo.Local,
		},
		GroupID: task.GroupID,
	}).Error
	if err != nil {
		return err
	}
	return nil
}

func (task *Task) CreateOrUpdate(db *gorm.DB) error {
	err := db.Clauses(
		clause.OnConflict{
			UpdateAll: true,
		}).Create(task).Error
	if err != nil {
		return err
	}
	return nil
}
