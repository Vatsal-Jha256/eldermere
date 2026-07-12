package config

import "testing"

func TestFromEnvDefaultsLocalToDevelopment(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("RENDER", "")

	cfg := FromEnv()

	if cfg.AppEnv != "development" {
		t.Fatalf("AppEnv = %q, want development", cfg.AppEnv)
	}
}

func TestFromEnvDefaultsRenderToProduction(t *testing.T) {
	t.Setenv("APP_ENV", "")
	t.Setenv("RENDER", "true")

	cfg := FromEnv()

	if cfg.AppEnv != "production" {
		t.Fatalf("AppEnv = %q, want production", cfg.AppEnv)
	}
}

func TestFromEnvUsesExplicitAppEnv(t *testing.T) {
	t.Setenv("APP_ENV", "staging")
	t.Setenv("RENDER", "true")

	cfg := FromEnv()

	if cfg.AppEnv != "staging" {
		t.Fatalf("AppEnv = %q, want staging", cfg.AppEnv)
	}
}

func TestFromEnvParsesAllowedOrigins(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", " https://eldermere.pages.dev, http://localhost:5173 ")

	cfg := FromEnv()

	if len(cfg.AllowedOrigins) != 2 {
		t.Fatalf("AllowedOrigins = %#v", cfg.AllowedOrigins)
	}
	if cfg.AllowedOrigins[0] != "https://eldermere.pages.dev" || cfg.AllowedOrigins[1] != "http://localhost:5173" {
		t.Fatalf("AllowedOrigins = %#v", cfg.AllowedOrigins)
	}
}

func TestFromEnvDefaultsAllowedOriginsToWildcard(t *testing.T) {
	t.Setenv("ALLOWED_ORIGINS", "")

	cfg := FromEnv()

	if len(cfg.AllowedOrigins) != 1 || cfg.AllowedOrigins[0] != "*" {
		t.Fatalf("AllowedOrigins = %#v, want wildcard", cfg.AllowedOrigins)
	}
}
