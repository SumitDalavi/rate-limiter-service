package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
)

// tokenBucketLua is an atomic Redis Lua script implementing a fixed-window counter.
// Returns 1 if allowed, 0 if rate limited.
const tokenBucketLua = `
local key     = KEYS[1]
local limit   = tonumber(ARGV[1])
local window  = tonumber(ARGV[2])
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

type RateLimiter struct {
	rdb    *redis.Client
	script *redis.Script
}

func New(rdb *redis.Client) *RateLimiter {
	return &RateLimiter{rdb: rdb, script: redis.NewScript(tokenBucketLua)}
}

// Allow checks whether the given key is within limit/window.
func (rl *RateLimiter) Allow(ctx context.Context, key string, limit, windowS int) (bool, error) {
	result, err := rl.script.Run(ctx, rl.rdb,
		[]string{key},
		limit, windowS,
	).Int()
	if err != nil {
		return true, fmt.Errorf("rate limiter redis error: %w", err) // fail open
	}
	return result == 1, nil
}

// RemainingTTL returns how many seconds until the current window expires.
func (rl *RateLimiter) RemainingTTL(ctx context.Context, key string) time.Duration {
	ttl, err := rl.rdb.TTL(ctx, key).Result()
	if err != nil || ttl < 0 {
		return 0
	}
	return ttl
}

// Middleware returns an http.Handler middleware that enforces rate limiting.
// keyFn extracts the rate-limit key from the request (typically IP or API key).
func (rl *RateLimiter) Middleware(limit, windowS int, keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := "rl:" + keyFn(r)
			allowed, err := rl.Allow(r.Context(), key, limit, windowS)
			if err != nil {
				log.Printf("[ratelimit] redis error: %v — failing open", err)
				next.ServeHTTP(w, r)
				return
			}

			ttl := rl.RemainingTTL(r.Context(), key)
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Window", fmt.Sprintf("%ds", windowS))
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", int(ttl.Seconds())))

			if !allowed {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
				http.Error(w, `{"error":"rate limit exceeded","retry_after":`+fmt.Sprintf("%d", int(ttl.Seconds()))+`}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// IPKeyFn extracts the client IP as the rate-limit key.
func IPKeyFn(r *http.Request) string {
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.RemoteAddr
	}
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// APIKeyFn extracts the X-API-Key header as the rate-limit key.
func APIKeyFn(r *http.Request) string {
	key := r.Header.Get("X-API-Key")
	if key == "" {
		return IPKeyFn(r)
	}
	return key
}
