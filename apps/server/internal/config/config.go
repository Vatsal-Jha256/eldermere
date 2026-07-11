package config

import (
	"log/slog"
	"os"
	"strings"
)

type Config struct {
	AppEnv          string
	ServerAddr      string
	DatabaseURL     string
	ContentPacksDir string
	LogLevel        slog.Level
}

func FromEnv() Config {
	return Config{
		AppEnv:          getEnv("APP_ENV", defaultAppEnv()),
		ServerAddr:      getEnv("SERVER_ADDR", ":8080"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://eldermere:eldermere@localhost:5432/eldermere?sslmode=disable"),
		ContentPacksDir: getEnv("CONTENT_PACKS_DIR", ""),
		LogLevel:        parseLogLevel(getEnv("LOG_LEVEL", "info")),
	}
}

func defaultAppEnv() string {
	if strings.TrimSpace(os.Getenv("RENDER")) != "" {
		return "production"
	}
	return "development"
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
