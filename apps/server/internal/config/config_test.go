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
