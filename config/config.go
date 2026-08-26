package config

import (
	"os"
	"strconv"
	"strings"
)

type RouteLimit struct {
	Path    string
	Limit   int
	WindowS int
}

type Config struct {
	RedisURL    string
	Port        string
	DefaultLimit   int
	DefaultWindowS int
	Routes      []RouteLimit
}

func Load() *Config {
	c := &Config{
		RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
		Port:           getEnv("PORT", "8080"),
		DefaultLimit:   getInt("DEFAULT_LIMIT", 100),
		DefaultWindowS: getInt("DEFAULT_WINDOW_S", 60),
		Routes: []RouteLimit{
			{Path: "/api/auth", Limit: 10, WindowS: 60},
			{Path: "/api/search", Limit: 30, WindowS: 60},
			{Path: "/api/export", Limit: 5, WindowS: 60},
		},
	}
	return c
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" { return v }
	return def
}

func getInt(k string, def int) int {
	if v := os.Getenv(k); v != "" {
		if i, err := strconv.Atoi(strings.TrimSpace(v)); err == nil { return i }
	}
	return def
}
