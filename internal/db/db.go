package db

import (
	"github.com/maksroxx/flowkeeper/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Connect(cfg config.DatabaseConfig) (*gorm.DB, error) {
	switch cfg.Driver {
	case "postgres":
		return gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	case "sqlite":
		return gorm.Open(sqlite.Open(cfg.DSN), &gorm.Config{})
	default:
		return gorm.Open(sqlite.Open("local.db"), &gorm.Config{})
	}
}
