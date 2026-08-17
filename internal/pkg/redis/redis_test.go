package redis

import (
	"context"
	"testing"
	"time"

	dto "github.com/prometheus/client_model/go"

	"github.com/north-fy/levelup/internal/config"
	"github.com/north-fy/levelup/internal/pkg/metrics"
)

func TestMetricsHookObservesCommands(t *testing.T) {
	rdb, err := New(config.Redis{Host: "localhost", Port: 6379})
	if err != nil {
		t.Skipf("redis unavailable: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Close() })

	ctx := context.Background()
	if err := rdb.Set(ctx, "test:metrics:hook", "v", time.Minute).Err(); err != nil {
		t.Fatalf("set: %v", err)
	}
	t.Cleanup(func() { _ = rdb.Del(ctx, "test:metrics:hook") })

	obs := metrics.RedisOperationDuration.WithLabelValues("set")
	hist, ok := obs.(interface{ Write(*dto.Metric) error })
	if !ok {
		t.Fatal("expected histogram to support Write")
	}
	var m dto.Metric
	_ = hist.Write(&m)
	if m.GetHistogram() == nil || m.GetHistogram().GetSampleCount() < 1 {
		t.Fatal("expected at least one 'set' sample")
	}
}
