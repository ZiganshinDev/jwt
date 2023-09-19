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
	"github.com/ZiganshinDev/medods/internal/http-server/handlers"
	"github.com/ZiganshinDev/medods/internal/lib/logger/sl"
	"github.com/ZiganshinDev/medods/internal/repository"
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

		return
	}

	db := mongoClient.Database(cfg.Mongo.Name)

	//TODO handlers && routers && middlewaver && JWT && refresh etc
	tokenManager, err := auth.NewManager(cfg.Auth.JWT.SigningKey)
	if err != nil {
		log.Error("failed to init auth", sl.Err(err))

		return
	}

	repos := repository.NewRepositories(db)
	services := service.NewServices(service.Deps{
		Repos:           repos,
		TokenManager:    tokenManager,
		AccessTokenTTL:  cfg.Auth.JWT.AccessTokenTTL,
		RefreshTokenTTL: cfg.Auth.JWT.RefreshTokenTTL,
	})

	handlers := handlers.NewHandler(services, tokenManager)

	srv := server.New(cfg, handlers.Init(cfg))

	go func() {
		if err := srv.Run(); !errors.Is(err, http.ErrServerClosed) {
			log.Error("failed to init http server", sl.Err(err))
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
	}

	if err := mongoClient.Disconnect(context.Background()); err != nil {
		log.Error(err.Error())
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
