package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"

	"github.com/north-fy/levelup/internal/config"
)

// NewClickHouse opens a connection pool to ClickHouse via the native driver.
func NewClickHouse(cfg config.CH) (*sql.DB, error) {
	conn := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)},
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.User,
			Password: cfg.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		DialTimeout: 10 * time.Second,
	})

	conn.SetMaxIdleConns(5)
	conn.SetMaxOpenConns(10)
	conn.SetConnMaxLifetime(time.Hour)

	return conn, nil
}

// PingClickHouse verifies connectivity to ClickHouse.
func PingClickHouse(ctx context.Context, db *sql.DB) error {
	return db.PingContext(ctx)
}
