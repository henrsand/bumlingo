# AGENTS.md

Go module `bumlingo` (Go 1.27.4, pgx/v5 + godotenv): CLI utilities over a PostGIS database.
No API server yet — the plan is a stdlib `net/http` server in `cmd/api` (first endpoint
`GET /api/points`, bare JSON arrays; SQL stays in `internal/db`).

## Commands

- Build one tool (binary lands in repo root): `go build -o <name> ./cmd/<name>`
- Verify everything: `go build ./... && go vet ./... && go test ./...`
- Single test: `go test ./internal/db -run TestReadTableSanitizesIdentifier -v`
- Run a tool — every command first calls `config.LoadDotEnv()`, which loads `.env` from the
  working directory via godotenv: real environment variables win over `.env`, a missing
  `.env` is fine, a malformed one only prints a warning to stderr. Run from the repo root,
  or `.env` is silently skipped. `.env` is gitignored — create it from `.env-example` and
  fill in real credentials:
  ```sh
  cp .env-example .env   # contains DATABASE_URL and ADDR placeholders
  ./show_data pois gis        # table + optional schema
  ./show_data -json pois      # JSON output
  ```
- Run the API: `go run ./cmd/api` (listens on `ADDR`, default `:8080`), then
  `curl localhost:8080/healthz` and `curl localhost:8080/api/points`

## Layout

- `cmd/<name>/main.go` — one command per directory: CLIs (`show_data`, `show_tables`,
  `read_points`, `insert_point`) plus the HTTP server `api`
- `internal/api/` — HTTP handlers (stdlib `net/http`, Go 1.22+ method patterns). Depends on
  small consumer-side interfaces (e.g. `PointsReader`), so tests use fakes and need no database
- `internal/config/` — `LoadDotEnv` (.env loading; every cmd calls it before `db.Connect`)
- `internal/db/` — shared helpers and all SQL: `Connect`, `ReadTable`, `ReadTables`,
  `InsertPoint`, `ReadPoints`

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
