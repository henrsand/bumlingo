// Package db provides the database connection and generic query helpers
// shared by the CLI utilities and the future API server.
package db

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect opens a connection pool to the database described by the
// DATABASE_URL environment variable, e.g.
//
//	postgres://user:pass@localhost:5432/gis?sslmode=disable
//
// The returned pool is safe for concurrent use and must be closed by the
// caller when done.
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return nil, fmt.Errorf("DATABASE_URL is not set")
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse DATABASE_URL: %w", err)
	}

	// Fail fast instead of hanging on an unreachable host.
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return pool, nil
}

func ReadTables(ctx context.Context, pool *pgxpool.Pool, schema ...string) ([]string, error) {
	var rows []string
	var query string
	var args []any
	if len(schema) > 0 {
		query = "SELECT table_name FROM information_schema.tables WHERE table_schema = $1"
		args = []any{schema[0]}
	} else {
		query = "SELECT table_name FROM information_schema.tables"
	}

	rowsResult, err := pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query tables: %w", err)
	}
	defer rowsResult.Close()

	for rowsResult.Next() {
		var tableName string
		if err := rowsResult.Scan(&tableName); err != nil {
			return nil, fmt.Errorf("scan table name: %w", err)
		}
		rows = append(rows, tableName)
	}

	if rowsResult.Err() != nil {
		return nil, fmt.Errorf("iterate rows: %w", rowsResult.Err())
	}

	return rows, nil
}
