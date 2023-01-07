package localSqlite3

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

func Init(dsn string) (*gorm.DB, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	err = os.MkdirAll(filepath.Join(cwd, "..", "data", "database", "localSqlite3"), 0666)

	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	return db, nil
}
