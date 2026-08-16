package middleware

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/north-fy/levelup/internal/pkg/ratelimit"
)

// RateLimitGlobal limits the overall number of requests from an IP within a
// window. It is meant to run on the engine before authentication.
func RateLimitGlobal(limiter ratelimit.Limiter, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		bucket := time.Now().Unix() / int64(window/time.Second)
		key := "rl:ip:" + hashScope(c.ClientIP()) + ":" + strconv.FormatInt(bucket, 10)
		allow(c, limiter, key, limit, window)
	}
}

// RateLimitUser limits per-user requests within a window. It must run after
// authentication so that the user id is present in the context.
func RateLimitUser(limiter ratelimit.Limiter, limit int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		bucket := time.Now().Unix() / int64(window/time.Second)
		var key string
		if id := UserID(c); id != 0 {
			key = "rl:user:" + strconv.FormatUint(uint64(id), 10) + ":" + strconv.FormatInt(bucket, 10)
		} else {
			key = "rl:ip:" + hashScope(c.ClientIP()) + ":" + strconv.FormatInt(bucket, 10)
		}
		allow(c, limiter, key, limit, window)
	}
}

func allow(c *gin.Context, limiter ratelimit.Limiter, key string, limit int, window time.Duration) {
	ok, err := limiter.Allow(c.Request.Context(), key, limit, window)
	if err != nil || !ok {
		c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		return
	}
	c.Next()
}

func hashScope(s string) string {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return fmt.Sprintf("%x", h.Sum64())
}
