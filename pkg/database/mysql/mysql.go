package mysql

import (
	"fmt"
	"github.com/anton1ks96/college-core-api/internal/config"
	"github.com/anton1ks96/college-core-api/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func NewClient(cfg *config.Config) (*sqlx.DB, error) {
	db, err := sqlx.Open("mysql", cfg.Database.DSN)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	db.SetMaxOpenConns(cfg.Database.MaxConnections)
	db.SetMaxIdleConns(cfg.Database.MaxIdle)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("connected to MySQL successfully")
	return db, nil
}
