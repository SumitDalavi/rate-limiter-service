package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SumitDalavi/rate-limiter-service/middleware"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

func newTestRL(t *testing.T) (*middleware.RateLimiter, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	return middleware.New(client), mr
}

func TestAllow_UnderLimit(t *testing.T) {
	rl, _ := newTestRL(t)
	for i := 0; i < 5; i++ {
		allowed, err := rl.Allow(t.Context(), "test:ip", 10, 60)
		if err != nil { t.Fatalf("unexpected error: %v", err) }
		if !allowed { t.Fatalf("expected allowed on request %d", i+1) }
	}
}

func TestAllow_OverLimit(t *testing.T) {
	rl, _ := newTestRL(t)
	for i := 0; i < 10; i++ {
		rl.Allow(t.Context(), "test:burst", 5, 60)
	}
	allowed, _ := rl.Allow(t.Context(), "test:burst", 5, 60)
	if allowed { t.Fatal("expected rate limited after exceeding limit") }
}

func TestAllow_DifferentKeysDontInterfere(t *testing.T) {
	rl, _ := newTestRL(t)
	for i := 0; i < 5; i++ {
		rl.Allow(t.Context(), "key-a", 3, 60)
	}
	allowed, _ := rl.Allow(t.Context(), "key-b", 3, 60)
	if !allowed { t.Fatal("key-b should not be limited by key-a exhaustion") }
}

func TestMiddleware_SetsHeaders(t *testing.T) {
	rl, _ := newTestRL(t)
	handler := rl.Middleware(10, 60, middleware.IPKeyFn)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)
	req := httptest.NewRequest("GET", "/api/test", nil)
	req.RemoteAddr = "1.2.3.4:9999"
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK { t.Errorf("expected 200, got %d", rr.Code) }
	if rr.Header().Get("X-RateLimit-Limit") == "" { t.Error("missing X-RateLimit-Limit header") }
}

func TestMiddleware_Returns429WhenLimited(t *testing.T) {
	rl, _ := newTestRL(t)
	handler := rl.Middleware(2, 60, middleware.IPKeyFn)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
	)
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/api", nil)
		req.RemoteAddr = "9.9.9.9:1234"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if i >= 2 && rr.Code != http.StatusTooManyRequests {
			t.Errorf("req %d: expected 429, got %d", i, rr.Code)
		}
	}
}

func TestIPKeyFn_ExtractsIP(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:54321"
	got := middleware.IPKeyFn(req)
	if got != "10.0.0.1" { t.Errorf("expected 10.0.0.1, got %q", got) }
}

func TestIPKeyFn_XForwardedFor(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-Forwarded-For", "203.0.113.10")
	got := middleware.IPKeyFn(req)
	if got != "203.0.113.10" { t.Errorf("expected 203.0.113.10, got %q", got) }
}
