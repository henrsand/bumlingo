#!/bin/sh
# Deploy bumlingoapi on this host: verify, build the static binary, install it,
# and (re)start the systemd service `bumlingoapi`.
#
# Run from the repo on the VPS:
#
#	git pull && ./deploy.sh
#
# One-time setup is documented in README.md ("Deploy"): /etc/bumlingo/env with
# DATABASE_URL and ADDR. This script creates the bumlingo user and /opt/bumlingo
# if missing, and never touches /etc/bumlingo/env.
set -eu

cd "$(dirname "$0")"

SUDO=""
if [ "$(id -u)" -ne 0 ]; then
	SUDO=sudo
fi

echo "==> go build ./... && go vet ./... && go test ./..."
go build ./... && go vet ./... && go test ./...

echo "==> building static binary (CGO_ENABLED=0)"
CGO_ENABLED=0 go build -trimpath -o api ./cmd/api

if ! id bumlingo >/dev/null 2>&1; then
	echo "==> creating bumlingo user"
	$SUDO useradd -r -s /usr/sbin/nologin bumlingo
fi
$SUDO mkdir -p /opt/bumlingo /etc/bumlingo

if [ ! -f /etc/bumlingo/env ]; then
	echo "WARNING: /etc/bumlingo/env is missing - create it (see README.md) or the service will not start" >&2
fi

echo "==> installing binary and systemd unit"
$SUDO install -m 0755 api /opt/bumlingo/api
$SUDO install -m 0644 deploy/bumlingoapi.service /etc/systemd/system/bumlingoapi.service
$SUDO systemctl daemon-reload
$SUDO systemctl enable bumlingoapi
$SUDO systemctl restart bumlingoapi

echo "==> service status"
$SUDO systemctl status bumlingoapi --no-pager || true

ADDR=$(sed -n 's/^ADDR=//p' /etc/bumlingo/env 2>/dev/null || true)
echo "==> health check: http://localhost${ADDR:-:8080}/healthz"
curl -fsS "http://localhost${ADDR:-:8080}/healthz"
echo
