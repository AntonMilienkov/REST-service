package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/AntonMilienkov/REST-service/internal/config"
	"github.com/AntonMilienkov/REST-service/internal/handler"
	"github.com/AntonMilienkov/REST-service/internal/repository"
	"github.com/AntonMilienkov/REST-service/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	if err := repository.Migrate(cfg.DSN(), "migrations"); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	pool, err := repository.NewPool(context.Background(), cfg.DSN())
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer pool.Close()

	repo := repository.NewSubscriptionRepository(pool)
	svc := service.NewSubscriptionService(repo)
	h := handler.NewSubscriptionHandler(svc)
	router := handler.NewRouter(h, logger)

	logger.Info("starting server", "port", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
