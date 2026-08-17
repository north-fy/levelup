package database

import (
	"context"
	"time"

	"github.com/north-fy/levelup/internal/config"
	"github.com/north-fy/levelup/internal/pkg/metrics"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

const metricsStartKey = "levelup:metrics:start"

// NewPostgres opens a GORM connection pool to PostgreSQL.
func NewPostgres(cfg config.DB) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	})
	if err != nil {
		return nil, err
	}

	registerMetricsCallbacks(db)

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(25)
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

// registerMetricsCallbacks instruments GORM operations with Prometheus timing.
func registerMetricsCallbacks(db *gorm.DB) {
	for name, operation := range map[string]string{
		"query":  "query",
		"create": "create",
		"update": "update",
		"delete": "delete",
	} {
		before := db.Callback().Query().Before("gorm:" + name)
		after := db.Callback().Query().After("gorm:" + name)
		switch name {
		case "create":
			before = db.Callback().Create().Before("gorm:" + name)
			after = db.Callback().Create().After("gorm:" + name)
		case "update":
			before = db.Callback().Update().Before("gorm:" + name)
			after = db.Callback().Update().After("gorm:" + name)
		case "delete":
			before = db.Callback().Delete().Before("gorm:" + name)
			after = db.Callback().Delete().After("gorm:" + name)
		}
		_ = before.Register("levelup:metrics:start", func(db *gorm.DB) {
			db.Statement.Settings.Store(metricsStartKey, time.Now())
		})
		_ = after.Register("levelup:metrics:done", func(db *gorm.DB) {
			start, ok := db.Statement.Settings.Load(metricsStartKey)
			if !ok {
				return
			}
			if t, ok := start.(time.Time); ok {
				metrics.DBQueryDuration.WithLabelValues(operation).Observe(time.Since(t).Seconds())
			}
		})
	}
}

// PingPostgres verifies connectivity to PostgreSQL.
func PingPostgres(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
