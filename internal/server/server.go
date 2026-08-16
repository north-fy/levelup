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
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"gorm.io/gorm"

	_ "github.com/north-fy/levelup/docs"

	"github.com/north-fy/levelup/internal/config"
	"github.com/north-fy/levelup/internal/handlers"
	"github.com/north-fy/levelup/internal/middleware"
	"github.com/north-fy/levelup/internal/outbox"
	"github.com/north-fy/levelup/internal/pkg/cache"
	"github.com/north-fy/levelup/internal/pkg/database"
	"github.com/north-fy/levelup/internal/pkg/jwt"
	"github.com/north-fy/levelup/internal/pkg/ratelimit"
	"github.com/north-fy/levelup/internal/repositories"
	"github.com/north-fy/levelup/internal/services"
)

// Server wraps the Gin engine and the underlying http.Server.
type Server struct {
	engine           *gin.Engine
	http             *http.Server
	pg               *gorm.DB
	ch               *sql.DB
	redis            *redis.Client
	log              *zap.Logger
	auth             *handlers.AuthHandler
	users            *handlers.UserHandler
	branches         *handlers.BranchHandler
	quests           *handlers.QuestHandler
	shop             *handlers.ShopHandler
	roadmaps         *handlers.RoadmapHandler
	workshop         *handlers.WorkshopHandler
	stats            *handlers.StatsHandler
	authSvc          *services.AuthService
	jwtMgr           *jwt.Manager
	flusher          *outbox.Flusher
	limiter          ratelimit.Limiter
	rateLimitPerUser int
	rateLimitWindow  time.Duration
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

	limiter := ratelimit.NewRedisLimiter(rdb)
	engine.Use(middleware.RateLimitGlobal(limiter, cfg.RateLimit.Global, cfg.RateLimit.Window))

	userRepo := repositories.NewUserRepository(pg)
	tokenStore := repositories.NewTokenStore(rdb)
	branchRepo := repositories.NewBranchRepository(pg)
	questRepo := repositories.NewQuestRepository(pg)
	shopItemRepo := repositories.NewShopItemRepository(pg)
	purchaseRepo := repositories.NewPurchaseRepository(pg)
	roadmapRepo := repositories.NewRoadmapRepository(pg)
	workshopRepo := repositories.NewWorkshopRepository(pg)
	outboxRepo := repositories.NewOutboxRepository(pg)
	jwtMgr := jwt.New(cfg.JWT)

	cacheStore := cache.NewRedisCache(rdb)

	authService := services.NewAuthService(userRepo, tokenStore, jwtMgr, cfg.GitHub)
	userService := services.NewUserService(userRepo, cacheStore)
	branchService := services.NewBranchService(branchRepo)
	questService := services.NewQuestService(questRepo, branchRepo, userRepo, services.NewOutboxQuestPublisher(outboxRepo), cacheStore)
	shopService := services.NewShopService(shopItemRepo, purchaseRepo, userRepo, services.NewOutboxPurchasePublisher(outboxRepo), cacheStore)
	roadmapService := services.NewRoadmapService(roadmapRepo, userRepo, services.NewOutboxQuestPublisher(outboxRepo), cacheStore)
	workshopService := services.NewWorkshopService(workshopRepo, roadmapRepo, cacheStore)
	statsService := services.NewStatsService(ch, userRepo, cacheStore)

	s := &Server{
		engine:           engine,
		pg:               pg,
		ch:               ch,
		redis:            rdb,
		log:              log,
		auth:             handlers.NewAuthHandler(authService),
		users:            handlers.NewUserHandler(userService),
		branches:         handlers.NewBranchHandler(branchService),
		quests:           handlers.NewQuestHandler(questService),
		shop:             handlers.NewShopHandler(shopService),
		roadmaps:         handlers.NewRoadmapHandler(roadmapService),
		workshop:         handlers.NewWorkshopHandler(workshopService),
		stats:            handlers.NewStatsHandler(statsService),
		authSvc:          authService,
		jwtMgr:           jwtMgr,
		flusher:          outbox.NewFlusher(outboxRepo, ch, log),
		limiter:          limiter,
		rateLimitPerUser: cfg.RateLimit.PerUser,
		rateLimitWindow:  cfg.RateLimit.Window,
	}
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
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	api := s.engine.Group("/api/v1")

	auth := api.Group("/auth")
	auth.POST("/register", s.auth.Register)
	auth.POST("/login", s.auth.Login)
	auth.POST("/refresh", s.auth.Refresh)
	auth.POST("/logout", s.auth.Logout)
	auth.GET("/github/redirect", s.auth.GitHubRedirect)
	auth.GET("/github/callback", s.auth.GitHubCallback)

	protected := api.Group("")
	protected.Use(middleware.Authenticate(s.jwtMgr, s.isBlacklisted))
	protected.Use(middleware.RateLimitUser(s.limiter, s.rateLimitPerUser, s.rateLimitWindow))
	protected.GET("/users/me", s.users.Me)
	protected.PATCH("/users/me", s.users.UpdateMe)

	protected.POST("/branches", s.branches.Create)
	protected.GET("/branches", s.branches.List)
	protected.GET("/branches/:id", s.branches.Get)
	protected.PATCH("/branches/:id", s.branches.Update)
	protected.DELETE("/branches/:id", s.branches.Delete)

	protected.POST("/branches/:branch_id/quests", s.quests.Create)
	protected.GET("/branches/:branch_id/quests", s.quests.List)
	protected.GET("/quests/:id", s.quests.Get)
	protected.PATCH("/quests/:id", s.quests.Update)
	protected.DELETE("/quests/:id", s.quests.Delete)
	protected.POST("/quests/:id/complete", s.quests.Complete)
	protected.POST("/quests/:id/start", s.quests.Start)
	protected.POST("/quests/:id/stop", s.quests.Stop)

	protected.POST("/shop/items", s.shop.Create)
	protected.GET("/shop/items", s.shop.List)
	protected.PATCH("/shop/items/:id", s.shop.Update)
	protected.DELETE("/shop/items/:id", s.shop.Delete)
	protected.POST("/shop/items/:id/buy", s.shop.Buy)
	protected.GET("/shop/purchases", s.shop.Purchases)

	protected.POST("/roadmaps", s.roadmaps.Create)
	protected.GET("/roadmaps", s.roadmaps.List)
	protected.GET("/roadmaps/:id", s.roadmaps.Get)
	protected.PATCH("/roadmaps/:id", s.roadmaps.Update)
	protected.DELETE("/roadmaps/:id", s.roadmaps.Delete)
	protected.POST("/roadmaps/:id/nodes", s.roadmaps.AddNode)
	protected.PATCH("/roadmaps/:id/nodes/:nodeId", s.roadmaps.UpdateNode)
	protected.POST("/roadmaps/:id/nodes/:nodeId/complete", s.roadmaps.CompleteNode)

	protected.POST("/workshop/roadmaps", s.workshop.Create)
	protected.GET("/workshop/roadmaps", s.workshop.List)
	protected.PATCH("/workshop/roadmaps/:id", s.workshop.Update)
	protected.POST("/workshop/roadmaps/:id/install", s.workshop.Install)

	protected.GET("/stats/overview", s.stats.Overview)
	protected.GET("/stats/branches", s.stats.Branches)
	protected.GET("/stats/roadmaps", s.stats.Roadmaps)
	protected.GET("/stats/quests", s.stats.Quests)
}

// StartBackground launches the outbox flusher.
func (s *Server) StartBackground(ctx context.Context) {
	if s.flusher != nil {
		go s.flusher.Run(ctx)
	}
}

func (s *Server) isBlacklisted(c *gin.Context, tokenID string) (bool, error) {
	return s.authSvc.IsTokenBlacklisted(c.Request.Context(), tokenID)
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
