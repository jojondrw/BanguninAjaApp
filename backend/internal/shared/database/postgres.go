package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/jojondrw/BanguninAjaApp/backend/internal/shared/config"
)

const (
	maxIdleConnections     = 10
	maxOpenConnections     = 50
	connectionMaxLifetime  = time.Hour
	connectionMaxIdleTime  = 10 * time.Minute
	slowQueryThresholdInMs = 200
)

func Open(cfg config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.Database.DSN()), &gorm.Config{
		Logger:                 gormLogger(cfg.App),
		SkipDefaultTransaction: true,
		TranslateError:         true,
	})
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	pool, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("read connection pool: %w", err)
	}

	pool.SetMaxIdleConns(maxIdleConnections)
	pool.SetMaxOpenConns(maxOpenConnections)
	pool.SetConnMaxLifetime(connectionMaxLifetime)
	pool.SetConnMaxIdleTime(connectionMaxIdleTime)

	return db, nil
}

func Close(db *gorm.DB) error {
	pool, err := db.DB()
	if err != nil {
		return err
	}
	return pool.Close()
}

func gormLogger(app config.App) logger.Interface {
	level := logger.Info
	if app.IsProduction() {
		level = logger.Warn
	}

	return logger.New(log.New(os.Stdout, "", log.LstdFlags), logger.Config{
		SlowThreshold:             slowQueryThresholdInMs * time.Millisecond,
		LogLevel:                  level,
		IgnoreRecordNotFoundError: true,
		Colorful:                  !app.IsProduction(),
	})
}
