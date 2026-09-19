package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgresPool(ctx context.Context, connstring string) (*pgxpool.Pool, error){
	config, err := pgxpool.ParseConfig(connstring)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse database config: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("Failed make postgres pool %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5 * time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		return nil, fmt.Errorf("Failed ping database: %w", err)
	}

	return pool, nil
}