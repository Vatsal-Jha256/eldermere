package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	AppEnv      string
	ServerAddr  string
	DatabaseURL string
	LogLevel    slog.Level
}

func FromEnv() Config {
	return Config{
		AppEnv:      getEnv("APP_ENV", "development"),
		ServerAddr:  getEnv("SERVER_ADDR", ":8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://eldermere:eldermere@localhost:5432/eldermere?sslmode=disable"),
		LogLevel:    parseLogLevel(getEnv("LOG_LEVEL", "info")),
	}
}

func getEnv(key, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func parseLogLevel(value string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
