// Package postgres implements connecting and pinging db
package postgres

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/devathh/staffy-sso/internal/infrastructure/config"
	"github.com/devathh/staffy-sso/internal/infrastructure/persistence"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectToDB(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Secrets.Postgres.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect pg: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql db: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.Secrets.Postgres.MaxIdleConn)
	sqlDB.SetMaxOpenConns(cfg.Secrets.Postgres.MaxOpenConn)

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping: %w", err)
	}

	return db, nil
}

func MustLoadDB(cfg *config.Config) *gorm.DB {
	db, err := ConnectToDB(cfg)
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	return db
}

func SyncDB(db *gorm.DB) error {
	return db.AutoMigrate(&persistence.UserModel{})
}
