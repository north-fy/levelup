package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/pkg/metrics"
)

// Metrics records Prometheus counters and histograms for every request.
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unmatched"
		}

		c.Next()

		metrics.HTTPRequests.WithLabelValues(
			c.Request.Method, path, strconv.Itoa(c.Writer.Status()),
		).Inc()
		metrics.HTTPDuration.WithLabelValues(
			c.Request.Method, path,
		).Observe(time.Since(start).Seconds())
	}
}
