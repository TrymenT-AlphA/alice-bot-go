package model

type Play struct {
	Name string `gorm:"primarykey"`
	PID  int64  `gorm:"primarykey"`
}
