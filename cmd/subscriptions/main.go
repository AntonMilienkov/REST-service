package main

import (
	"log"

	"github.com/AntonMilienkov/REST-service/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	log.Printf("config loaded: server_port=%s db_host=%s db_name=%s", cfg.ServerPort, cfg.DBHost, cfg.DBName)
}
