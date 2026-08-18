package novaredis

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/luaxlou/nova/starter/novaconfig"
	"github.com/redis/go-redis/v9"
)

var (
	client      *redis.Client
	initialized bool
	mu          sync.RWMutex
	declared    bool
)

// Init declares that the application will use Redis.
// This should be called during application startup, similar to other starters.
func Init() {
	mu.Lock()
	defer mu.Unlock()
	declared = true
}

// Client returns the automatically initialized Redis client.
func Client() (*redis.Client, error) {
	mu.RLock()
	if initialized && client != nil {
		defer mu.RUnlock()
		return client, nil
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()

	if initialized && client != nil {
		return client, nil
	}

	// NOTE: With local-config-only mode (with local config mode), we don't request resources at runtime.


	log.Printf("Lazy initializing Redis Starter...")

	// Read Redis config from local config (provided by local config file)
	addr := novaconfig.GetString("redis.addr")
	if addr == "" {
		return nil, fmt.Errorf("redis_addr not found in config")
	}

	c := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: novaconfig.GetString("redis.username"),
		Password: novaconfig.GetString("redis.password"),
		DB:       novaconfig.GetInt("redis.db"),
	})

	if err := c.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	client = c
	initialized = true
	log.Println("Redis Starter initialized successfully.")

	return client, nil
}

func Reload() {
	mu.Lock()
	defer mu.Unlock()
	if client != nil {
		client.Close()
		client = nil
	}
	initialized = false
}
