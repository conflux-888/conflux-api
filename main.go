package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/conflux-888/conflux-api/swagger"
	"github.com/conflux-888/conflux-api/internal/common/logger"
	"github.com/conflux-888/conflux-api/internal/common/middleware"
	"github.com/conflux-888/conflux-api/internal/config"
	"github.com/conflux-888/conflux-api/internal/event"
	"github.com/conflux-888/conflux-api/internal/infrastructure/database"
	"github.com/conflux-888/conflux-api/internal/infrastructure/server"
	"github.com/conflux-888/conflux-api/internal/report"
	"github.com/conflux-888/conflux-api/internal/sync"
	"github.com/conflux-888/conflux-api/internal/user"
	"github.com/rs/zerolog/log"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Conflux API
// @version         1.0
// @description     Global Threat Report API — aggregates conflict events from GDELT and user reports.
// @host            localhost:8080
// @BasePath        /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your Bearer token as: Bearer <token>
func main() {
	cfg := config.Load()
	logger.Init(cfg.LogLevel)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.Connect(ctx, cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		log.Fatal().Err(err).Msg("[main] failed to connect to database")
	}

	// Auth middleware
	authMW := middleware.Auth(cfg.JWTSecret)

	// User domain
	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo, cfg.JWTSecret)
	userHandler := user.NewHandler(userSvc)

	// Event domain
	eventRepo := event.NewRepository(db)
	eventSvc := event.NewService(eventRepo)
	eventHandler := event.NewHandler(eventSvc)

	// Report domain
	reportClusterRepo := report.NewRepository(db)
	reportSvc := report.NewService(eventRepo, reportClusterRepo)
	reportHandler := report.NewHandler(reportSvc)

	// Sync domain
	syncClient := sync.NewClient()
	syncStateRepo := sync.NewStateRepository(db)
	syncSvc := sync.NewService(syncClient, eventRepo, syncStateRepo, cfg.SyncIntervalMinutes)
	syncHandler := sync.NewHandler(syncSvc)

	// Router
	router, v1 := server.NewRouter()
	user.RegisterRoutes(v1, userHandler, authMW)
	event.RegisterRoutes(v1, eventHandler, authMW)
	report.RegisterRoutes(v1, reportHandler, authMW)
	sync.RegisterRoutes(v1, syncHandler, authMW)

	// Swagger (outside /api/v1)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Start background sync
	go syncSvc.Start(ctx)

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		log.Info().Str("port", cfg.Port).Msg("[main] server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("[main] server error")
		}
	}()

	<-ctx.Done()
	log.Info().Msg("[main] shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal().Err(err).Msg("[main] shutdown error")
	}
	log.Info().Msg("[main] server stopped")
}
