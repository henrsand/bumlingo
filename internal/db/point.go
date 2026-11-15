package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Point struct {
	Label    string
	Lon, Lat float64
}

func InsertPoint(ctx context.Context, q Execer, p Point) (int64, error) {
	if p.Lon < -180 || p.Lon > 180 || p.Lat < -90 || p.Lat > 90 {
		return 0, fmt.Errorf("Longitude or latitude out of bounds")
	}
	query := "INSERT INTO poi (label, geom) VALUES ($1, ST_SetSRID(ST_MakePoint($2,$3), 4326));"
	tag, err := q.Exec(ctx, query, p.Label, p.Lon, p.Lat)
	if err != nil {
		return 0, fmt.Errorf("Something went wrong inserting point: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return 0, fmt.Errorf("Expected 1 affected row, got: %d", tag.RowsAffected())
	}
	return tag.RowsAffected(), nil
}

func ReadPoints(ctx context.Context, q Querier) ([]Point, error) {
	query := "SELECT label, ST_X(geom) as lon, ST_Y(geom) as lat FROM poi;"
	rows, err := q.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("Something went wrong fetching points: %s", err)
	}
	defer rows.Close()

	points, err := pgx.CollectRows(rows, pgx.RowToStructByName[Point])
	if err != nil {
		return nil, fmt.Errorf("Something went wrong mapping points: %s", err)
	}

	return points, nil
}

type Execer interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}
