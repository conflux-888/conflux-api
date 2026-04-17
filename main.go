package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/conflux-888/conflux-api/swagger"
	"github.com/conflux-888/conflux-api/internal/adminauth"
	"github.com/conflux-888/conflux-api/internal/common/gemini"
	"github.com/conflux-888/conflux-api/internal/common/logger"
	"github.com/conflux-888/conflux-api/internal/common/middleware"
	"github.com/conflux-888/conflux-api/internal/config"
	"github.com/conflux-888/conflux-api/internal/event"
	"github.com/conflux-888/conflux-api/internal/infrastructure/database"
	"github.com/conflux-888/conflux-api/internal/infrastructure/server"
	"github.com/conflux-888/conflux-api/internal/infrastructure/staticfs"
	"github.com/conflux-888/conflux-api/internal/notification"
	"github.com/conflux-888/conflux-api/internal/preferences"
	"github.com/conflux-888/conflux-api/internal/report"
	"github.com/conflux-888/conflux-api/internal/summary"
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

	// Auth middlewares
	authMW := middleware.Auth(cfg.JWTSecret)
	adminAuthMW := middleware.AdminAuth(cfg.JWTSecret)

	// Admin auth domain (ADMIN_USER / ADMIN_PASSWORD)
	adminAuthSvc := adminauth.NewService(cfg.AdminUser, cfg.AdminPasswordHash, cfg.JWTSecret)
	adminAuthHandler := adminauth.NewHandler(adminAuthSvc)
	if !adminAuthSvc.Configured() {
		log.Warn().Msg("[main] ADMIN_USER / ADMIN_PASSWORD not set — admin endpoints will reject all traffic")
	}

	// User domain
	userRepo := user.NewRepository(db)
	userSvc := user.NewService(userRepo, cfg.JWTSecret)
	userHandler := user.NewHandler(userSvc)

	// Event domain
	eventRepo := event.NewRepository(db)
	eventSvc := event.NewService(eventRepo)
	eventHandler := event.NewHandler(eventSvc)

	// Report domain
	reportSvc := report.NewService(eventRepo)
	reportHandler := report.NewHandler(reportSvc)

	// Sync domain
	syncClient := sync.NewClient()
	syncStateRepo := sync.NewStateRepository(db)
	syncSvc := sync.NewService(syncClient, eventRepo, syncStateRepo, cfg.SyncIntervalMinutes)
	syncHandler := sync.NewHandler(syncSvc)

	// Preferences domain
	prefsRepo := preferences.NewRepository(db)
	prefsSvc := preferences.NewService(prefsRepo)
	prefsHandler := preferences.NewHandler(prefsSvc)

	// Notification domain
	notifRepo := notification.NewRepository(db)
	notifSvc := notification.NewService(notifRepo, prefsRepo)
	notifHandler := notification.NewHandler(notifSvc)

	// Hook notification service into sync + event (for admin seed)
	syncSvc.SetNotifier(notifSvc)
	eventSvc.SetNotifier(notifSvc)

	// Router
	router, v1 := server.NewRouter(server.RouterOptions{
		CORSAllowLocalhost: cfg.CORSAllowLocalhost,
	})
	adminauth.RegisterRoutes(v1, adminAuthHandler)
	user.RegisterRoutes(v1, userHandler, authMW)
	event.RegisterRoutes(v1, eventHandler, authMW, adminAuthMW)
	report.RegisterRoutes(v1, reportHandler, authMW)
	sync.RegisterRoutes(v1, syncHandler, adminAuthMW)
	preferences.RegisterRoutes(v1, prefsHandler, authMW)
	notification.RegisterRoutes(v1, notifHandler, authMW)

	// Summary domain (optional — requires GEMINI_API_KEY)
	var summaryScheduler *summary.Scheduler
	if cfg.GeminiAPIKey != "" {
		geminiClient, err := gemini.NewClient(ctx, cfg.GeminiAPIKey)
		if err != nil {
			log.Fatal().Err(err).Msg("[main] failed to create Gemini client")
		}
		defer geminiClient.Close()
		summaryRepo := summary.NewRepository(db)
		summarySvc := summary.NewService(summaryRepo, eventRepo, geminiClient)
		summarySvc.SetNotifier(notifSvc)
		summaryHandler := summary.NewHandler(summarySvc)
		summaryScheduler = summary.NewScheduler(summarySvc, cfg.SummaryCheckIntervalMin, cfg.SummaryBackfillDays)
		summary.RegisterRoutes(v1, summaryHandler, authMW, adminAuthMW)
	} else {
		log.Warn().Msg("[main] GEMINI_API_KEY not set, summary feature disabled")
	}

	// Swagger (outside /api/v1)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// Admin UI (outside /api/v1), served from embedded dist
	if cfg.AdminUIEnabled {
		staticfs.Register(router, adminFS(), "/admin")
	}

	// Start background jobs
	go syncSvc.Start(ctx)
	if summaryScheduler != nil {
		go summaryScheduler.Start(ctx)
	}

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
