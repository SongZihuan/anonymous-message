package database

import (
	"fmt"
	"github.com/SongZihuan/anonymous-message/src/flagparser"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

func InitSQLite() error {
	if flagparser.SQLitePath == "" {
		return nil
	}

	_db, err := gorm.Open(sqlite.Open(flagparser.SQLitePath), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("connect to sqlite (%s) failed: %s", flagparser.SQLitePath, err)
	}

	err = _db.AutoMigrate(&Mail{})
	if err != nil {
		return fmt.Errorf("migrate sqlite (%s) failed: %s", flagparser.SQLitePath, err)
	}

	db = _db
	return nil
}

func CloseSQLite() {
	if db == nil {
		return
	}

	defer func() {
		db = nil
	}()

	if flagparser.SQLiteActiveClose {
		// https://github.com/go-gorm/gorm/issues/3145
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
}
