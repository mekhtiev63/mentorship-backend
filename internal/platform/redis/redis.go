package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-mentorship-platform/backend/internal/platform/config"
	goredis "github.com/redis/go-redis/v9"
)

// Client wraps go-redis for DI and health checks.
type Client struct {
	*goredis.Client
	healthTimeout time.Duration
}

// New creates a Redis client and verifies connectivity.
func New(ctx context.Context, cfg config.Redis) (*Client, error) {
	client := goredis.NewClient(&goredis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, cfg.HealthTimeout)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return &Client{
		Client:        client,
		healthTimeout: cfg.HealthTimeout,
	}, nil
}

// Ping checks Redis connectivity.
func (c *Client) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, c.healthTimeout)
	defer cancel()
	return c.Client.Ping(pingCtx).Err()
}

// Close closes the Redis client.
func (c *Client) Close() error {
	return c.Client.Close()
}
