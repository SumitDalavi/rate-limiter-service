package metrics

import (
	"expvar"
	"net/http"
	"sync/atomic"
)

var (
	TotalRequests   = expvar.NewInt("rate_limiter.total_requests")
	AllowedRequests = expvar.NewInt("rate_limiter.allowed_requests")
	DeniedRequests  = expvar.NewInt("rate_limiter.denied_requests")
	RedisErrors     = expvar.NewInt("rate_limiter.redis_errors")
)

// Counter is a simple atomic counter for per-route metrics.
type Counter struct{ n int64 }

func (c *Counter) Inc()         { atomic.AddInt64(&c.n, 1) }
func (c *Counter) Value() int64 { return atomic.LoadInt64(&c.n) }

// MetricsHandler exposes /metrics as JSON via expvar.
func MetricsHandler() http.Handler { return expvar.Handler() }
