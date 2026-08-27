package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SumitDalavi/rate-limiter-service/config"
	"github.com/SumitDalavi/rate-limiter-service/middleware"
	"github.com/SumitDalavi/rate-limiter-service/routes"
	"github.com/go-redis/redis/v8"
)

func TestRateLimiter(t *testing.T) {
	// For testing, ideally mock Redis or use miniredis.
	// This assumes a local redis is running or we just test the handler logic mock.
	
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	
	// Try ping
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		t.Skip("Skipping test because redis is not running on localhost:6379")
	}

	// Flush db for clean state
	rdb.FlushDB(context.Background())

	rl := middleware.New(rdb)
	cfg := config.Config{DefaultLimit: 3, DefaultWindowS: 10}

	mux := http.NewServeMux()
	routes.Register(mux, rl, &cfg)

	for i := 1; i <= 4; i++ {
		req, _ := http.NewRequest("GET", "/api/data", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		rr := httptest.NewRecorder()
		mux.ServeHTTP(rr, req)

		if i <= 3 {
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("Request %d: handler returned wrong status code: got %v want %v",
					i, status, http.StatusOK)
			}
		} else {
			if status := rr.Code; status != http.StatusTooManyRequests {
				t.Errorf("Request %d: handler returned wrong status code: got %v want %v",
					i, status, http.StatusTooManyRequests)
			}
		}
	}
}
