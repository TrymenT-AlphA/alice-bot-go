package model

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Group struct {
	ID int64 `gorm:"primarykey"`
}

func (*Group) TableName() string {
	return "Group"
}

func (*Group) ReadAll(db *gorm.DB) ([]Group, error) {
	var result []Group
	if err := db.Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (group *Group) Create(db *gorm.DB) error {
	if err := db.Create(group).Error; err != nil {
		return err
	}
	return nil
}

func (group *Group) Read(db *gorm.DB) error {
	if err := db.Where(&Group{
		ID: group.ID,
	}).First(group).Error; err != nil {
		return err
	}
	return nil
}

func (group *Group) Update(db *gorm.DB) error {
	if err := db.Where(&Group{
		ID: group.ID,
	}).Updates(group).Error; err != nil {
		return nil
	}
	return nil
}

func (group *Group) Delete(db *gorm.DB) error {
	if err := db.Delete(&Group{
		ID: group.ID,
	}).Error; err != nil {
		return err
	}
	return nil
}

func (group *Group) CreateOrUpdate(db *gorm.DB) error {
	if err := db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(group).Error; err != nil {
		return err
	}
	return nil
}
