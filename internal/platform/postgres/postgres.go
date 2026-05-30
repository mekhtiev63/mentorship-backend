package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/go-mentorship-platform/backend/internal/platform/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Pool wraps pgxpool for dependency injection and health checks.
type Pool struct {
	*pgxpool.Pool
	healthTimeout time.Duration
}

// New creates a PostgreSQL connection pool.
func New(ctx context.Context, cfg config.Database) (*Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create pg pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, cfg.HealthTimeout)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &Pool{
		Pool:          pool,
		healthTimeout: cfg.HealthTimeout,
	}, nil
}

// Ping checks database connectivity.
func (p *Pool) Ping(ctx context.Context) error {
	pingCtx, cancel := context.WithTimeout(ctx, p.healthTimeout)
	defer cancel()
	return p.Pool.Ping(pingCtx)
}

// Close shuts down the pool.
func (p *Pool) Close() {
	p.Pool.Close()
}
