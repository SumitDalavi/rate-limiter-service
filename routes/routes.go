package routes

import (
	"encoding/json"
	"net/http"

	"github.com/SumitDalavi/rate-limiter-service/config"
	"github.com/SumitDalavi/rate-limiter-service/metrics"
	"github.com/SumitDalavi/rate-limiter-service/middleware"
)

func Register(mux *http.ServeMux, rl *middleware.RateLimiter, cfg *config.Config) {
	// Build a route-limit lookup map
	routeLimits := make(map[string]config.RouteLimit)
	for _, rl2 := range cfg.Routes {
		routeLimits[rl2.Path] = rl2
	}

	// Health check — no rate limiting
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// Metrics endpoint
	mux.Handle("/metrics", metrics.MetricsHandler())

	// Demo protected routes — each with its own limit
	for path, limit := range routeLimits {
		p, l := path, limit // capture loop vars
		handler := rl.Middleware(l.Limit, l.WindowS, middleware.IPKeyFn)(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				metrics.AllowedRequests.Add(1)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]interface{}{
					"path": p, "limit": l.Limit, "window_s": l.WindowS,
					"message": "request allowed",
				})
			}),
		)
		mux.Handle(p, countingMiddleware(handler))
	}

	// Default catch-all with global limit
	mux.Handle("/api/", rl.Middleware(cfg.DefaultLimit, cfg.DefaultWindowS, middleware.IPKeyFn)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			metrics.AllowedRequests.Add(1)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"path": r.URL.Path, "message": "ok"})
		}),
	))
}

func countingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.TotalRequests.Add(1)
		next.ServeHTTP(w, r)
	})
}
