package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/north-fy/levelup/internal/config"
	"github.com/north-fy/levelup/internal/pkg/metrics"
)

// New creates a Redis client and verifies connectivity.
func New(cfg config.Redis) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	client.AddHook(metricsHook{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

// metricsHook times every Redis command and reports it to Prometheus.
type metricsHook struct{}

func (metricsHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h metricsHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmd)
		metrics.RedisOperationDuration.WithLabelValues(cmd.Name()).Observe(time.Since(start).Seconds())
		return err
	}
}

func (h metricsHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		start := time.Now()
		err := next(ctx, cmds)
		metrics.RedisOperationDuration.WithLabelValues("pipeline").Observe(time.Since(start).Seconds())
		return err
	}
}
