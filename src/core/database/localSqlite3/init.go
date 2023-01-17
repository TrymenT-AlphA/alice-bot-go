package localSqlite3

import (
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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
	db, err := gorm.Open(sqlite.Open(filepath.Join(localSqlite3Dir, dbFile)), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
			NoLowerCase:   true,
		},
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}
