package server

import (
	"context"
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/north-fy/levelup/internal/config"
	"github.com/north-fy/levelup/internal/middleware"
	"github.com/north-fy/levelup/internal/pkg/database"
)

// Server wraps the Gin engine and the underlying http.Server.
type Server struct {
	engine *gin.Engine
	http   *http.Server
	pg     *gorm.DB
	ch     *sql.DB
	redis  *redis.Client
	log    *zap.Logger
}

// New builds the HTTP server with middleware and registered routes.
func New(cfg *config.Config, log *zap.Logger, pg *gorm.DB, ch *sql.DB, rdb *redis.Client) *Server {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	engine.Use(
		middleware.Recovery(log),
		middleware.RequestID(),
		middleware.Logger(log),
		middleware.Metrics(),
	)

	s := &Server{engine: engine, pg: pg, ch: ch, redis: rdb, log: log}
	s.routes()

	httpServer := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.App.Port),
		Handler:      engine,
		ReadTimeout:  cfg.App.ReadTimeout,
		WriteTimeout: cfg.App.WriteTimeout,
		IdleTimeout:  cfg.App.IdleTimeout,
	}
	s.http = httpServer

	return s
}

func (s *Server) routes() {
	s.engine.GET("/healthz", s.healthz)
	s.engine.GET("/readyz", s.readyz)
	s.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := s.engine.Group("/api/v1")
	_ = api
}

func (s *Server) healthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) readyz(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
	defer cancel()

	checks := map[string]string{}

	if err := database.PingPostgres(ctx, s.pg); err != nil {
		checks["postgres"] = err.Error()
	} else {
		checks["postgres"] = "ok"
	}

	if err := database.PingClickHouse(ctx, s.ch); err != nil {
		checks["clickhouse"] = err.Error()
	} else {
		checks["clickhouse"] = "ok"
	}

	if err := s.redis.Ping(ctx).Err(); err != nil {
		checks["redis"] = err.Error()
	} else {
		checks["redis"] = "ok"
	}

	status := http.StatusOK
	body := gin.H{"status": "ok", "checks": checks}
	for _, v := range checks {
		if v != "ok" {
			status = http.StatusServiceUnavailable
			body["status"] = "unavailable"
			break
		}
	}

	c.JSON(status, body)
}

// ListenAndServe starts the HTTP server and blocks until it fails.
func (s *Server) ListenAndServe() error {
	return s.http.ListenAndServe()
}

// Addr returns the address the server listens on.
func (s *Server) Addr() string {
	return s.http.Addr
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}
