package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/conflux-888/conflux-api/internal/common/logger"
	"github.com/conflux-888/conflux-api/internal/common/middleware"
	"github.com/conflux-888/conflux-api/internal/config"
	"github.com/conflux-888/conflux-api/internal/infrastructure/database"
	"github.com/conflux-888/conflux-api/internal/infrastructure/server"
	"github.com/conflux-888/conflux-api/internal/user"
	"github.com/rs/zerolog/log"
)

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

	// Router
	router := server.NewRouter()
	user.RegisterRoutes(router, userHandler, authMW)

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
