// Command api serves the GIS data over HTTP.
//
// Usage:
//
//	./api   # listens on ADDR (default :8080); DATABASE_URL must be set (see .env)
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"henrik_gis/internal/api"
	"henrik_gis/internal/config"
	"henrik_gis/internal/db"
)

// poolPoints adapts a db.Querier to api.PointsReader.
type poolPoints struct {
	q db.Querier
}

func (p poolPoints) ReadPoints(ctx context.Context) ([]db.Point, error) {
	return db.ReadPoints(ctx, p.q)
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	if err := config.LoadDotEnv(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading .env: %v\n", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	addr := os.Getenv("ADDR")
	if addr == "" {
		addr = ":8080"
	}

	srv := &http.Server{
		Addr:              addr,
		Handler:           api.NewServer(poolPoints{q: pool}).Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
