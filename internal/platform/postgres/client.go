package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
)

// Client represents a PostgreSQL database client
type Client struct {
	Pool *pgxpool.Pool
}

// NewClient creates a new PostgreSQL client with connection pooling
func NewClient(ctx context.Context, dsn string) (*Client, error) {
	// Parse config with defaults
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse PostgreSQL DSN: %w", err)
	}

	// Set connection pool settings
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = time.Minute

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create PostgreSQL connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	log.Info().Msg("Successfully connected to PostgreSQL!")

	return &Client{Pool: pool}, nil
}

// Close closes the database connection pool
func (c *Client) Close() {
	if c.Pool != nil {
		c.Pool.Close()
		log.Info().Msg("PostgreSQL connection pool closed")
	}
}

// Ping checks if the database connection is alive
func (c *Client) Ping(ctx context.Context) error {
	return c.Pool.Ping(ctx)
}
