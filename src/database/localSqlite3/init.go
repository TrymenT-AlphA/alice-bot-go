package localSqlite3

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Init(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
