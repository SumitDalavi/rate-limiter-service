package config

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

var current atomic.Pointer[Config]

func init() {
	c := Load()
	current.Store(c)
}

// Get returns the current live configuration (safe for concurrent reads)
func Get() *Config {
	c := current.Load()
	if c == nil {
		c = Load()
		current.Store(c)
	}
	return c
}

// InitWatcher starts the fsnotify config hot-reloader if CONFIG_MOUNT_PATH is set
func InitWatcher() {
	mount := os.Getenv("CONFIG_MOUNT_PATH")
	if mount == "" {
		log.Println("[config] no CONFIG_MOUNT_PATH — using boot-time config only")
		return
	}

	if err := watchMount(mount); err != nil {
		log.Printf("[config] watcher unavailable (non-fatal): %v", err)
	}
}

func watchMount(mountPath string) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("fsnotify: %w", err)
	}
	if err := w.Add(mountPath); err != nil {
		w.Close()
		return err
	}

	go func() {
		defer w.Close()
		for {
			select {
			case e, ok := <-w.Events:
				if !ok {
					return
				}
				// Kubernetes ConfigMap update causes symlink swap or create
				if e.Op&(fsnotify.Create|fsnotify.Write) != 0 {
					updated := Load()
					current.Store(updated)
					log.Printf("[config] hot-reload: limit=%d window=%ds port=%s",
						updated.DefaultLimit, updated.DefaultWindowS, updated.Port)
				}
			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				log.Printf("[config] watcher error: %v", err)
			}
		}
	}()

	log.Printf("[config] hot-reload active, watching: %s", mountPath)
	return nil
}
