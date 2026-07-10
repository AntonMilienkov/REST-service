package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/AntonMilienkov/REST-service/internal/config"
	_ "github.com/AntonMilienkov/REST-service/docs"
	"github.com/AntonMilienkov/REST-service/internal/handler"
	"github.com/AntonMilienkov/REST-service/internal/repository"
	"github.com/AntonMilienkov/REST-service/internal/service"
)

// @title Subscriptions API
// @version 1.0
// @description REST-сервис агрегации данных об онлайн-подписках пользователей.
// @host localhost:8080
// @BasePath /
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: parseLogLevel(cfg.LogLevel)}))

	if err := repository.Migrate(cfg.DSN(), "migrations"); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	pool, err := repository.NewPool(context.Background(), cfg.DSN())
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer pool.Close()

	repo := repository.NewSubscriptionRepository(pool, logger)
	svc := service.NewSubscriptionService(repo, logger)
	h := handler.NewSubscriptionHandler(svc)
	router := handler.NewRouter(h, logger)

	logger.Info("starting server", "port", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
