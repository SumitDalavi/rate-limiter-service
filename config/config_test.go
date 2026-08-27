package config

import (
	"os"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	// Ensure env vars are clear
	os.Unsetenv("REDIS_URL")
	os.Unsetenv("PORT")
	os.Unsetenv("DEFAULT_LIMIT")
	os.Unsetenv("DEFAULT_WINDOW_S")

	cfg := Load()

	if cfg.RedisURL != "redis://localhost:6379" {
		t.Errorf("expected default redis url, got %s", cfg.RedisURL)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected default port 8080, got %s", cfg.Port)
	}
	if cfg.DefaultLimit != 100 {
		t.Errorf("expected default limit 100, got %d", cfg.DefaultLimit)
	}
	if cfg.DefaultWindowS != 60 {
		t.Errorf("expected default window 60, got %d", cfg.DefaultWindowS)
	}
}

func TestLoadWithEnvVars(t *testing.T) {
	os.Setenv("REDIS_URL", "redis://remote:6379")
	os.Setenv("PORT", "9090")
	os.Setenv("DEFAULT_LIMIT", "50")
	os.Setenv("DEFAULT_WINDOW_S", "120")

	defer func() {
		os.Unsetenv("REDIS_URL")
		os.Unsetenv("PORT")
		os.Unsetenv("DEFAULT_LIMIT")
		os.Unsetenv("DEFAULT_WINDOW_S")
	}()

	cfg := Load()

	if cfg.RedisURL != "redis://remote:6379" {
		t.Errorf("expected redis url redis://remote:6379, got %s", cfg.RedisURL)
	}
	if cfg.Port != "9090" {
		t.Errorf("expected port 9090, got %s", cfg.Port)
	}
	if cfg.DefaultLimit != 50 {
		t.Errorf("expected limit 50, got %d", cfg.DefaultLimit)
	}
	if cfg.DefaultWindowS != 120 {
		t.Errorf("expected window 120, got %d", cfg.DefaultWindowS)
	}
}

func TestGetIntInvalid(t *testing.T) {
	os.Setenv("DEFAULT_LIMIT", "invalid")
	defer os.Unsetenv("DEFAULT_LIMIT")

	// Should fallback to default if parsing fails
	cfg := Load()
	if cfg.DefaultLimit != 100 {
		t.Errorf("expected fallback limit 100 on invalid input, got %d", cfg.DefaultLimit)
	}
}
