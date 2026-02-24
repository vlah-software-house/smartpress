// Package cache provides Valkey (Redis-compatible) client initialization
// and page caching functionality for the SmartPress rendering engine.
package cache

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

// ConnectValkey creates a Valkey client and verifies the connection with a ping.
func ConnectValkey(host, port, password string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("valkey ping: %w", err)
	}

	slog.Info("valkey connected", "addr", fmt.Sprintf("%s:%s", host, port))
	return client, nil
}
