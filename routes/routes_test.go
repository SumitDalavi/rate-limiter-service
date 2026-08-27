package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"

	"github.com/SumitDalavi/rate-limiter-service/config"
	"github.com/SumitDalavi/rate-limiter-service/middleware"
)

func setupRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return mr, client
}

func TestRegister(t *testing.T) {
	mr, client := setupRedis(t)
	defer mr.Close()
	defer client.Close()

	rl := middleware.New(client)
	cfg := &config.Config{
		DefaultLimit:   100,
		DefaultWindowS: 60,
		Routes: []config.RouteLimit{
			{Path: "/api/auth", Limit: 1, WindowS: 60},
		},
	}

	mux := http.NewServeMux()
	Register(mux, rl, cfg)

	// Test /health
	reqHealth, _ := http.NewRequest("GET", "/health", nil)
	rrHealth := httptest.NewRecorder()
	mux.ServeHTTP(rrHealth, reqHealth)
	if rrHealth.Code != http.StatusOK {
		t.Errorf("expected 200 OK for /health, got %d", rrHealth.Code)
	}

	// Test /api/auth (route-specific limit)
	reqAuth, _ := http.NewRequest("GET", "/api/auth", nil)
	reqAuth.RemoteAddr = "10.0.0.1:1234"

	// First request - OK
	rrAuth1 := httptest.NewRecorder()
	mux.ServeHTTP(rrAuth1, reqAuth)
	if rrAuth1.Code != http.StatusOK {
		t.Errorf("expected 200 OK for /api/auth, got %d", rrAuth1.Code)
	}

	// Second request - Rate limited
	rrAuth2 := httptest.NewRecorder()
	mux.ServeHTTP(rrAuth2, reqAuth)
	if rrAuth2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 for /api/auth, got %d", rrAuth2.Code)
	}

	// Test catch-all /api/unknown
	reqApi, _ := http.NewRequest("GET", "/api/unknown", nil)
	rrApi := httptest.NewRecorder()
	mux.ServeHTTP(rrApi, reqApi)
	if rrApi.Code != http.StatusOK {
		t.Errorf("expected 200 OK for /api/unknown, got %d", rrApi.Code)
	}
}
