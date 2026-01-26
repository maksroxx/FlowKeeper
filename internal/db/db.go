package db

import (
	"github.com/maksroxx/flowkeeper/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	switch cfg.Driver {
	case "postgres":
		db, err = gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
		if err == nil {
			db.Exec("PRAGMA foreign_keys = ON")
		}
	default:
		db, err = gorm.Open(sqlite.Open("local.db"), &gorm.Config{})
		if err == nil {
			db.Exec("PRAGMA foreign_keys = ON")
		}
	}

	return db, err
}
