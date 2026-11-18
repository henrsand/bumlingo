// Command show_data prints all rows from a database table.
//
// Usage:
//
//	DATABASE_URL=postgres://... ./show_data <table> [schema]
//	DATABASE_URL=postgres://... ./show_data -json <table> [schema]
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"text/tabwriter"

	"bumlingo/internal/config"
	"bumlingo/internal/db"
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

	jsonOut := flag.Bool("json", false, "print rows as JSON instead of a text table")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [-json] <table> [schema]\n\n", os.Args[0])
		fmt.Fprintln(flag.CommandLine.Output(), "Prints all rows from a table. The connection string is read from DATABASE_URL.")
		flag.PrintDefaults()
	}
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 || len(args) > 2 {
		flag.Usage()
		os.Exit(2)
	}
	table := args[0]
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

	data, err := db.ReadTable(ctx, pool, table, schema...)
	if err != nil {
		return err
	}

	if *jsonOut {
		return printJSON(data)
	}
	printTable(data)
	return nil
}

func printTable(data *db.TableData) {
	if len(data.Rows) == 0 {
		fmt.Printf("(0 rows)\n")
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, strings.Join(data.Columns, "\t"))
	for _, row := range data.Rows {
		cells := make([]string, len(row))
		for i, v := range row {
			cells[i] = formatValue(v)
		}
		fmt.Fprintln(w, strings.Join(cells, "\t"))
	}
	w.Flush()
	fmt.Printf("(%d rows)\n", len(data.Rows))
}

// printJSON emits an array of objects keyed by column name.
func printJSON(data *db.TableData) error {
	records := make([]map[string]any, len(data.Rows))
	for i, row := range data.Rows {
		rec := make(map[string]any, len(data.Columns))
		for j, col := range data.Columns {
			rec[col] = jsonValue(row[j])
		}
		records[i] = rec
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}

// jsonValue converts values that do not marshal to something readable
// (e.g. []byte becomes base64) into a sensible representation.
func jsonValue(v any) any {
	switch val := v.(type) {
	case nil:
		return nil
	case []byte:
		return fmt.Sprintf("\\x%x", val)
	default:
		return val
	}
}

func formatValue(v any) string {
	switch val := v.(type) {
	case nil:
		return "NULL"
	case []byte:
		return fmt.Sprintf("\\x%x", val)
	default:
		return fmt.Sprint(val)
	}
}
