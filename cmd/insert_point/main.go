package main

import (
	"context"
	"fmt"
	"henrik_gis/internal/config"
	"henrik_gis/internal/db"
	"os"
	"os/signal"
	"syscall"
)

type poi struct {
	label string
	lon   string
	lat   string
	err   error
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func run() error {
	err := config.LoadDotEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading .env: %v\n", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	point := db.Point{Label: "New point", Lon: 18.0586, Lat: 59.3300}
	num, err := db.InsertPoint(ctx, pool, point)
	if err != nil {
		return fmt.Errorf("Something went wrong: %s", err)
	}

	fmt.Println("All ok: %w", num)

	return nil
}
