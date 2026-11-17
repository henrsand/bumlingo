# AGENTS.md

Go module `henrik_gis` (Go 1.27.4, pgx/v5): CLI utilities over a PostGIS database. No API server yet.

## Commands

- Build one tool (binary lands in repo root): `go build -o <name> ./cmd/<name>`
- Verify everything: `go build ./... && go vet ./... && go test ./...`
- Single test: `go test ./internal/db -run TestReadTableSanitizesIdentifier -v`
- Run a tool — `DATABASE_URL` must be exported first (nothing loads `.env`, it is only gitignored):
  ```sh
  export DATABASE_URL='postgres://user:pass@localhost:5432/gis?sslmode=disable'
  ./show_data pois gis        # table + optional schema
  ./show_data -json pois      # JSON output
  ```

## Layout

- `cmd/<name>/main.go` — one CLI per directory: `show_data`, `show_tables`, `read_points`, `insert_point`
- `internal/db/` — shared helpers: `Connect`, `ReadTable`, `ReadTables`, `InsertPoint`, `ReadPoints`

## Gotchas

- No Makefile, CI, or pre-commit hooks; the commands above are the whole workflow.
- `go test ./...` needs no database — `internal/db` tests use a fake `Querier`.
- The executables in the repo root are build artifacts, not source. Only `/show_data` is
  gitignored; `show_data2`, `insert_point`, `read_points`, `show_tables` are untracked
  leftovers and `show_data2` has no `cmd/` source. Build from `cmd/`, never edit the binaries.
- `./show_tables` with no args exits 2 (the arg check requires 1-2 args despite the usage
  text); the single arg is a schema filter, e.g. `./show_tables public`.
- `./insert_point` inserts a hardcoded row ("New point", 18.0586, 59.3300) into `poi` on
  every run — it mutates the database, don't run it casually.
- `poi` table contract: columns `label` (text) and `geom` (geometry, SRID 4326).
  `ReadPoints` uses `pgx.RowToStructByName`, so query aliases must match `Point` field names.
- PostGIS geometry columns come back as raw WKB and print as `\x...` hex in `show_data`;
  decoding to WKT/GeoJSON is not implemented.
- `show_data` reads entire tables with no limit — inspection of small tables only.
