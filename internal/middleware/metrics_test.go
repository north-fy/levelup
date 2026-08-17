package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"

	"github.com/north-fy/levelup/internal/pkg/metrics"
)

func TestMetricsMiddlewareRecordsHTTP(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(Metrics())
	router.GET("/ok", func(c *gin.Context) { c.Status(http.StatusOK) })
	router.POST("/nope", func(c *gin.Context) { c.Status(http.StatusInternalServerError) })

	for _, m := range []string{http.MethodGet, http.MethodPost} {
		req := httptest.NewRequest(m, map[string]string{http.MethodGet: "/ok", http.MethodPost: "/nope"}[m], nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
	}

	if got := testutil.ToFloat64(metrics.HTTPRequests.WithLabelValues("GET", "/ok", "200")); got != 1 {
		t.Fatalf("GET /ok counter = %v, want 1", got)
	}
	if got := testutil.ToFloat64(metrics.HTTPRequests.WithLabelValues("POST", "/nope", "500")); got != 1 {
		t.Fatalf("POST /nope counter = %v, want 1", got)
	}
	if got := testutil.ToFloat64(metrics.HTTPInFlight); got != 0 {
		t.Fatalf("in-flight after requests = %v, want 0", got)
	}

	var m dto.Metric
	hist, ok := metrics.HTTPDuration.WithLabelValues("GET", "/ok").(interface{ Write(*dto.Metric) error })
	if !ok {
		t.Fatal("expected histogram to support Write")
	}
	_ = hist.Write(&m)
	if m.GetHistogram() == nil || m.GetHistogram().GetSampleCount() < 1 {
		t.Fatal("expected at least one duration sample")
	}
}
