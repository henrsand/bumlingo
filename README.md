# bumlingo

Go app exposing a PostGIS database through an API.

## Layout

```
cmd/show_data/     CLI: print all rows of a table
internal/db/       connection pool + generic table reading (shared with the future API)
```

## Configuration

The connection string is read from `DATABASE_URL`:

```sh
export DATABASE_URL='postgres://user:password@localhost:5432/gis?sslmode=disable'
```

## show_data

Build it:

```sh
go build -o show_data ./cmd/show_data
```

Print all rows of a table:

```sh
./show_data pois            # default schema
./show_data pois gis        # explicit schema
./show_data -json pois      # JSON output
```

Notes:

- The table name is quoted as an identifier, so it is safe against SQL injection.
- `NULL` is shown as `NULL` in text output and `null` in JSON.
- PostGIS geometry columns come back as raw WKB bytes, printed as `\x...` hex. Decoding
  them to WKT/GeoJSON is not implemented yet.
- There is no row limit or pagination; it is meant for inspection, not large tables.

## Development

```sh
go test ./...
go vet ./...
```
