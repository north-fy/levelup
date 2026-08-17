package metrics

import (
	"context"
	"database/sql"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTP metrics.
var (
	HTTPRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "levelup_http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)

	HTTPDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "levelup_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	HTTPInFlight = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "levelup_http_requests_in_flight",
			Help: "Number of HTTP requests currently being served.",
		},
	)
)

// Database metrics.
var (
	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "levelup_db_query_duration_seconds",
			Help:    "Database query duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	RedisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "levelup_redis_operation_duration_seconds",
			Help:    "Redis operation duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	CHQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "levelup_clickhouse_query_duration_seconds",
			Help:    "ClickHouse query duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

// Business metrics.
var (
	UsersRegistered = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "levelup_users_registered_total",
			Help: "Total number of registered users.",
		},
	)

	QuestsCompleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "levelup_quests_completed_total",
			Help: "Total number of completed quests.",
		},
	)

	NodesCompleted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "levelup_nodes_completed_total",
			Help: "Total number of completed roadmap nodes.",
		},
	)

	Purchases = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "levelup_purchases_total",
			Help: "Total number of completed purchases.",
		},
	)

	GoldSpent = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "levelup_gold_spent_total",
			Help: "Total gold spent by buyers.",
		},
	)

	RoadmapsInstalled = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "levelup_roadmaps_installed_total",
			Help: "Total number of workshop roadmaps installed.",
		},
	)
)

// DB wraps a database/sql handle and records query durations as ClickHouse
// metrics. It satisfies the query interfaces used by the stats and outbox
// services.
type DB struct {
	*sql.DB
}

// NewDB wraps a database/sql handle with duration instrumentation.
func NewDB(db *sql.DB) *DB {
	return &DB{DB: db}
}

// QueryContext records a ClickHouse query.
func (d *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	start := time.Now()
	rows, err := d.DB.QueryContext(ctx, query, args...)
	CHQueryDuration.WithLabelValues("query").Observe(time.Since(start).Seconds())
	return rows, err
}

// QueryRowContext records a ClickHouse row query.
func (d *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	start := time.Now()
	row := d.DB.QueryRowContext(ctx, query, args...)
	CHQueryDuration.WithLabelValues("queryrow").Observe(time.Since(start).Seconds())
	return row
}

// ExecContext records a ClickHouse insert.
func (d *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	res, err := d.DB.ExecContext(ctx, query, args...)
	CHQueryDuration.WithLabelValues("exec").Observe(time.Since(start).Seconds())
	return res, err
}
