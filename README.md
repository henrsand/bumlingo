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

## Deploy (VPS)

The API runs as a systemd service named `bumlingoapi` (unit file:
`deploy/bumlingoapi.service`). It is a single static binary installed at
`/opt/bumlingo/api`, configured through `/etc/bumlingo/env` — never through
the repo's `.env`, so secrets stay out of the git checkout.

One-time setup on the VPS:

```sh
sudo useradd -r -s /usr/sbin/nologin bumlingo
sudo mkdir -p /opt/bumlingo /etc/bumlingo
sudo tee /etc/bumlingo/env > /dev/null <<'EOF'
DATABASE_URL=postgres://gis_user:secret@localhost:5432/gis?sslmode=disable
ADDR=:8080
EOF
sudo chmod 600 /etc/bumlingo/env
sudo install -m 0644 deploy/bumlingoapi.service /etc/systemd/system/bumlingoapi.service
sudo systemctl daemon-reload
sudo systemctl enable --now bumlingoapi
```

Deploy or update from the repo on the VPS:

```sh
git pull && ./deploy.sh
```

`deploy.sh` runs `go build ./... && go vet ./... && go test ./...`, builds the
static binary (`CGO_ENABLED=0 go build -trimpath -o api ./cmd/api`), installs it
to `/opt/bumlingo/api`, reinstalls the unit, restarts `bumlingoapi`, then prints
the service status and hits `/healthz`. It creates the `bumlingo` user and
`/opt/bumlingo` if missing, and never touches `/etc/bumlingo/env`.

Check it:

```sh
sudo systemctl status bumlingoapi
journalctl -u bumlingoapi -f
curl -s localhost:8080/healthz
curl -s localhost:8080/api/points
```

Notes:

- Postgres is expected on the same host (`localhost:5432`). If it requires TLS,
  use `?sslmode=require` (or `verify-full`) in `DATABASE_URL`.
- To expose the API publicly, put a reverse proxy (Caddy/nginx) in front on 443
  and keep 8080 local-only.
