package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/north-fy/levelup/internal/config"
	"github.com/north-fy/levelup/internal/pkg/database"
	"github.com/north-fy/levelup/internal/pkg/logger"
	"github.com/north-fy/levelup/internal/pkg/redis"
	"github.com/north-fy/levelup/internal/server"
)

//	@title			LevelUp Tracker API
//	@version		1.0
//	@description	RPG-style task tracker: quests, gold, roadmaps and stats.
//	@host			localhost:8080
//	@BasePath		/api/v1
//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and the JWT access token.

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "server exited with error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log, err := logger.New(cfg.Log.Level)
	if err != nil {
		return err
	}
	defer func() { _ = log.Sync() }()

	pg, err := database.NewPostgres(cfg.DB)
	if err != nil {
		return err
	}

	ch, err := database.NewClickHouse(cfg.CH)
	if err != nil {
		return err
	}

	rdb, err := redis.New(cfg.Redis)
	if err != nil {
		return err
	}

	srv := server.New(cfg, log, pg, ch, rdb)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, 1)
	go func() {
		log.Info("server starting", zap.String("addr", srv.Addr()))
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return err
	}
	log.Info("server stopped gracefully")
	return nil
}
