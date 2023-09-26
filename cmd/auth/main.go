package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ZiganshinDev/medods/internal/auth"
	"github.com/ZiganshinDev/medods/internal/config"
	"github.com/ZiganshinDev/medods/internal/http-server/handler"
	"github.com/ZiganshinDev/medods/internal/http-server/middleware/logger"
	"github.com/ZiganshinDev/medods/internal/lib/logger/sl"
	"github.com/ZiganshinDev/medods/internal/server"
	"github.com/ZiganshinDev/medods/internal/service"
	"github.com/ZiganshinDev/medods/internal/storage/mongodb"
	"github.com/joho/godotenv"
	"golang.org/x/exp/slog"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	err := godotenv.Load("config.env")
	if err != nil {
		log.Fatal("error loading .env file")
	}

	cfg := config.MustLoad()

	log := setupLogger(cfg.Env)

	log.Info(
		"Starting AuthApp",
		slog.String("env", cfg.Env))
	log.Debug("debug messages are enabled")

	mongoClient, err := mongodb.NewClient(cfg.Mongo.URI, cfg.Mongo.User, cfg.Mongo.Password)
	if err != nil {
		log.Error("failed to init storage", sl.Err(err))
		os.Exit(1)
	}

	mongoDatabase := mongodb.NewStorage(mongoClient, cfg.Mongo.Database)
	mongoRefreshRepo := mongoDatabase.NewRefreshRepo()

	tokenManager, err := auth.NewManager(cfg.JWT.SigningKey)
	if err != nil {
		log.Error("failed to init auth", sl.Err(err))
		os.Exit(1)
	}

	service, err := service.New(cfg, mongoRefreshRepo, tokenManager)
	if err != nil {
		log.Error("failed to init service", sl.Err(err))
		os.Exit(1)
	}

	logger := logger.Log

	h := handler.New(cfg, service, logger)

	srv := server.New(cfg, h.NewRouter())

	go func() {
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to init http server", sl.Err(err))
			os.Exit(1)
		}
	}()

	log.Info("Server started")

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	const timeout = 5 * time.Second

	ctx, shutdown := context.WithTimeout(context.Background(), timeout)
	defer shutdown()

	if err := srv.Stop(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))
		os.Exit(1)
	}

	if err := mongoClient.Disconnect(context.Background()); err != nil {
		log.Error("failed to stop mongo client", sl.Err(err))
		os.Exit(1)
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
