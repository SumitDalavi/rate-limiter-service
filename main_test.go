package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

func TestRateLimiter(t *testing.T) {
	// For testing, ideally mock Redis or use miniredis. 
	// This assumes a local redis is running or we just test the handler logic mock.
	// But let's write a robust table driven test skeleton.
	
	rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	
	// Try ping
	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		t.Skip("Skipping test because redis is not running on localhost:6379")
	}

	// Flush db for clean state
	rdb.FlushDB(context.Background())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientIP := "127.0.0.1"
		allowed, _ := checkRateLimit(r.Context(), "rate:"+clientIP, 3, 10)
		if !allowed {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		w.Write([]byte("Success!"))
	})

	for i := 1; i <= 4; i++ {
		req, _ := http.NewRequest("GET", "/api/resource", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

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
