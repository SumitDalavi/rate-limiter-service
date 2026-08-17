package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

var rdb *redis.Client

const luaScript = `
local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])

local current = redis.call("INCR", key)
if current == 1 then
    redis.call("EXPIRE", key, window)
end

if current > limit then
    return 0
else
    return 1
end
`

func checkRateLimit(ctx context.Context, key string, limit int, window int) (bool, error) {
	res, err := rdb.Eval(ctx, luaScript, []string{key}, limit, window).Result()
	if err != nil { return true, err } // Fail open on error
	return res.(int64) == 1, nil
}

func main() {
	rdb = redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	http.HandleFunc("/api/resource", func(w http.ResponseWriter, r *http.Request) {
		clientIP := r.RemoteAddr // simplified
		allowed, err := checkRateLimit(r.Context(), "rate:"+clientIP, 10, 60)
		if err != nil { log.Printf("Redis err: %v", err) }
		if !allowed {
			http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			return
		}
		w.Write([]byte("Success!"))
	})
	log.Println("Rate Limiter Service running on :8080")
	http.ListenAndServe(":8080", nil)
}
