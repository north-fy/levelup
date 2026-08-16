package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// HTTPRequests counts HTTP requests by method, route and status code.
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
)
