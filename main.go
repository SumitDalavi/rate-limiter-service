package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SumitDalavi/rate-limiter-service/config"
	"github.com/SumitDalavi/rate-limiter-service/middleware"
	"github.com/SumitDalavi/rate-limiter-service/routes"
	"github.com/go-redis/redis/v8"
)

var quit = make(chan os.Signal, 1)

func run() int {
	cfg := config.Load()
	config.InitWatcher()

	// Connect to Redis
	opts, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Printf("invalid REDIS_URL: %v", err)
		return 1
	}
	rdb := redis.NewClient(opts)
	ctx := context.Background()
	if _, err := rdb.Ping(ctx).Result(); err != nil {
		log.Printf("cannot connect to Redis: %v", err)
		return 1
	}
	log.Println("[main] connected to Redis:", cfg.RedisURL)

	rl := middleware.New(rdb)

	mux := http.NewServeMux()
	routes.Register(mux, rl, cfg)

	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Printf("[main] rate limiter listening on :%s (default limit: %d req/%ds)",
		cfg.Port, cfg.DefaultLimit, cfg.DefaultWindowS)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("server error: %v", err)
			quit <- syscall.SIGTERM // trigger shutdown
		}
	}()
	<-quit
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shutCtx)
	log.Println("[main] shutdown complete")
	return 0
}

func main() {
	os.Exit(run())
}
