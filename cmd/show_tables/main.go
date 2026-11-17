// Command show_tables prints all table names in the database.
//
// Usage:
//
//	DATABASE_URL=postgres://... ./show_tables

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	"henrik_gis/internal/config"
	"henrik_gis/internal/db"
)

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

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s\n\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Prints all table names in the database. The connection string is read from DATABASE_URL.")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 || len(args) > 2 {
		flag.Usage()
		os.Exit(2)
	}
	var schema []string
	if len(args) == 2 {
		schema = args[1:]
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := db.Connect(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	data, err := db.ReadTables(ctx, pool, schema...)
	if err != nil {
		return err
	}

	printTables(data)
	return nil
}

func printTables(tables []string) {
	if len(tables) == 0 {
		fmt.Printf("(0 rows)\n")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, table := range tables {
		fmt.Fprintln(w, table)
	}
	w.Flush()
	fmt.Printf("(%d rows)\n", len(tables))
}
