package main

import (
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func TestRunSuccess(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	os.Setenv("REDIS_URL", "redis://"+mr.Addr())
	os.Setenv("PORT", "0") // OS assigned port
	defer os.Unsetenv("REDIS_URL")
	defer os.Unsetenv("PORT")

	// Start in a goroutine so it doesn't block
	done := make(chan int)
	go func() {
		done <- run()
	}()

	// Wait a moment for server to start, then kill it
	time.Sleep(100 * time.Millisecond)

	// Send SIGINT to trigger shutdown
	quit <- os.Interrupt

	select {
	case code := <-done:
		if code != 0 {
			t.Errorf("expected exit code 0, got %d", code)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("timeout waiting for run() to exit")
	}
}

func TestRunBadRedisURL(t *testing.T) {
	os.Setenv("REDIS_URL", "http://invalid-url")
	defer os.Unsetenv("REDIS_URL")

	if code := run(); code != 1 {
		t.Errorf("expected exit code 1 for bad redis url, got %d", code)
	}
}

func TestRunCannotConnectRedis(t *testing.T) {
	// A valid redis url but pointing to a closed port
	os.Setenv("REDIS_URL", "redis://localhost:12345")
	defer os.Unsetenv("REDIS_URL")

	if code := run(); code != 1 {
		t.Errorf("expected exit code 1 for closed redis, got %d", code)
	}
}
