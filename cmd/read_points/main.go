package main

import (
	"context"
	"fmt"
	"henrik_gis/internal/db"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	fmt.Println("Looking")
	points, err := db.ReadPoints(ctx, pool)
	if err != nil {
		fmt.Printf("Error when reading rows: %s", err)
	}
	fmt.Println("Looked")
	fmt.Println(len(points))

	for i, p := range points {
		fmt.Printf("Point %d: %s, %f, %f", i, p.Label, p.Lon, p.Lat)
		fmt.Println()
	}

	return nil
}
