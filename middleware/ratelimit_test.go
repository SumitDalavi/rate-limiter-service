package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
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

func TestAllow(t *testing.T) {
	mr, client := setupRedis(t)
	defer mr.Close()
	defer client.Close()

	rl := New(client)
	ctx := context.Background()

	// Should allow first request
	allowed, err := rl.Allow(ctx, "testkey", 1, 60)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Errorf("expected allowed=true")
	}

	// Should deny second request
	allowed, err = rl.Allow(ctx, "testkey", 1, 60)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Errorf("expected allowed=false")
	}
}

func TestRemainingTTL(t *testing.T) {
	mr, client := setupRedis(t)
	defer mr.Close()
	defer client.Close()

	rl := New(client)
	ctx := context.Background()
	rl.Allow(ctx, "ttlkey", 1, 60)

	ttl := rl.RemainingTTL(ctx, "ttlkey")
	if ttl <= 0 || ttl > 60*time.Second {
		t.Errorf("unexpected ttl: %v", ttl)
	}
}

func TestMiddleware(t *testing.T) {
	mr, client := setupRedis(t)
	defer mr.Close()
	defer client.Close()

	rl := New(client)
	mw := rl.Middleware(1, 60, IPKeyFn)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "1.2.3.4:1234"

	// Request 1 - OK
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", rr.Code)
	}

	// Request 2 - Rate limited
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("expected 429 Too Many Requests, got %d", rr2.Code)
	}
}

func TestMiddlewareRedisError(t *testing.T) {
	mr, client := setupRedis(t)
	mr.Close() // Close immediately to trigger error
	defer client.Close()

	rl := New(client)
	mw := rl.Middleware(1, 60, IPKeyFn)

	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	// Should fail open
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200 OK (fail open), got %d", rr.Code)
	}
}

func TestAPIKeyFn(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Set("X-API-Key", "my-key")

	key := APIKeyFn(req)
	if key != "my-key" {
		t.Errorf("expected my-key, got %s", key)
	}

	req2, _ := http.NewRequest("GET", "/", nil)
	req2.RemoteAddr = "10.0.0.1:80"

	key2 := APIKeyFn(req2)
	if key2 != "10.0.0.1" {
		t.Errorf("expected fallback 10.0.0.1, got %s", key2)
	}
}
