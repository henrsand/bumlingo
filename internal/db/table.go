package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// TableData holds the column names and all rows of a queried table.
type TableData struct {
	Columns []string
	Rows    [][]any
}

// ReadTable returns every row of the given table in the default schema
// (or the schema given as an optional second argument).
//
// Values are returned as Go types decoded by pgx; PostGIS geometry columns
// come back as their raw WKB byte slices.
func ReadTable(ctx context.Context, q Querier, table string, schema ...string) (*TableData, error) {
	if table == "" {
		return nil, fmt.Errorf("table name must not be empty")
	}

	target := pgx.Identifier{table}.Sanitize()
	if len(schema) > 1 {
		return nil, fmt.Errorf("at most one schema may be given")
	}
	if len(schema) == 1 && schema[0] != "" {
		target = pgx.Identifier{schema[0], table}.Sanitize()
	}

	rows, err := q.Query(ctx, "SELECT * FROM "+target)
	if err != nil {
		return nil, fmt.Errorf("query table %s: %w", target, err)
	}
	if rows == nil {
		return nil, fmt.Errorf("query table %s: no rows returned", target)
	}
	defer rows.Close()

	fields := rows.FieldDescriptions()
	columns := make([]string, len(fields))
	for i, f := range fields {
		columns[i] = f.Name
	}

	data := &TableData{Columns: columns}
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		data.Rows = append(data.Rows, values)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read rows: %w", err)
	}

	return data, nil
}

// Querier is the subset of pgx used by ReadTable, satisfied by both
// *pgxpool.Pool and pgx.Tx.
type Querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}
