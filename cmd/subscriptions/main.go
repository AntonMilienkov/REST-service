package main

import (
	"context"
	"log"

	"github.com/AntonMilienkov/REST-service/internal/config"
	"github.com/AntonMilienkov/REST-service/internal/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if err := repository.Migrate(cfg.DSN(), "migrations"); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	pool, err := repository.NewPool(context.Background(), cfg.DSN())
	if err != nil {
		log.Fatalf("connect to db: %v", err)
	}
	defer pool.Close()

	log.Printf("connected to db, server_port=%s", cfg.ServerPort)
}
