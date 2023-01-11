package localSqlite3

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path/filepath"
)

func Init(dbFile string) (*gorm.DB, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	localSqlite3Dir := filepath.Join(cwd, "..", "data", "database", "localSqlite3")

	err = os.MkdirAll(localSqlite3Dir, 0666)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(filepath.Join(localSqlite3Dir, dbFile)), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
