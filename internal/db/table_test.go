package db

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

// fakeQuerier records the query it receives and returns canned rows.
type fakeQuerier struct {
	gotSQL string
	rows   pgx.Rows
	err    error
}

func (f *fakeQuerier) Query(_ context.Context, sql string, _ ...any) (pgx.Rows, error) {
	f.gotSQL = sql
	return f.rows, f.err
}

func TestReadTableSanitizesIdentifier(t *testing.T) {
	tests := []struct {
		name   string
		table  string
		schema []string
		want   string
	}{
		{"plain", "pois", nil, `SELECT * FROM "pois"`},
		{"schema", "pois", []string{"gis"}, `SELECT * FROM "gis"."pois"`},
		{"empty schema", "pois", []string{""}, `SELECT * FROM "pois"`},
		{"injection", `pois"; DROP TABLE x; --`, nil, `SELECT * FROM "pois""; DROP TABLE x; --"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeQuerier{}
			// nil rows is fine: we only assert on the generated SQL here.
			if _, err := ReadTable(context.Background(), f, tt.table, tt.schema...); err == nil {
				t.Fatal("expected error from nil rows")
			}
			if f.gotSQL != tt.want {
				t.Errorf("sql = %q, want %q", f.gotSQL, tt.want)
			}
		})
	}
}

func TestReadTableRejectsEmptyName(t *testing.T) {
	if _, err := ReadTable(context.Background(), &fakeQuerier{}, ""); err == nil {
		t.Fatal("expected error for empty table name")
	}
}
