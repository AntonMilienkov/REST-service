package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config хранит все настройки приложения, читается один раз при старте.
type Config struct {
	ServerPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	LogLevel string
}

// Load читает .env (если он есть) и переменные окружения, возвращает готовый Config.
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("load .env: %w", err)
		}
	}

	cfg := &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),

		DBHost:     getEnv("DB_HOST", "localhost"),
		DBPort:     getEnv("DB_PORT", "5432"),
		DBUser:     getEnv("DB_USER", "postgres"),
		DBPassword: getEnv("DB_PASSWORD", "postgres"),
		DBName:     getEnv("DB_NAME", "subscriptions"),
		DBSSLMode:  getEnv("DB_SSLMODE", "disable"),

		LogLevel: getEnv("LOG_LEVEL", "info"),
	}

	return cfg, nil
}

// DSN собирает строку подключения к Postgres из полей Config.
func (c *Config) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName, c.DBSSLMode,
	)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return fallback
}
