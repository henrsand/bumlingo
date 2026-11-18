# Auth plan for the bumlingo API

Goal: only the owner's mobile app may call the API. Right-sized solution: a static
bearer token checked by middleware. No login/JWT yet.

## Recommendation

`Authorization: Bearer <API_TOKEN>` on every `/api/*` request; `/healthz` stays open.

- Token lives in the `API_TOKEN` env var; add it to `.env-example` (`.env` is gitignored,
  so the secret never enters git).
- Compare with `crypto/subtle.ConstantTimeCompare` (no timing oracle).
- Threat-model caveat: a token compiled into the app binary can be extracted by a
  determined attacker. Acceptable for a personal API; if per-user identity is ever
  needed, add `POST /auth/login` issuing short-lived HS256 JWTs (stdlib
  `crypto/hmac` suffices, or `golang-jwt/jwt/v5`) — the middleware shape stays the same,
  only token validation changes.

## Implementation

1. `internal/api/auth.go` — `requireAuth(token string, next http.Handler) http.Handler`:
   - missing/invalid token → 401 `{"error":"Unauthorized"}` + `WWW-Authenticate: Bearer realm="api"`
   - valid → next
2. `internal/api/server.go` — `Server` gains a `token` field, `NewServer(r PointsReader, token string)`;
   `Handler()` wires:
   ```go
   mux.HandleFunc("GET /healthz", s.healthz)
   apiMux := http.NewServeMux()
   apiMux.HandleFunc("GET /api/points", s.getPoints)
   mux.Handle("/api/", requireAuth(s.token, apiMux))
   ```
3. `cmd/api/main.go` — read `API_TOKEN` from env; refuse to start if empty (fail closed).
4. `.env-example` — add `API_TOKEN=...`.
5. Tests (`internal/api`): 401 without header, 401 wrong token, 200 correct token,
   `/healthz` 200 unauthenticated.
6. `AGENTS.md` gotcha: API requires `Authorization: Bearer` from `API_TOKEN`; `/healthz`
   is open; serve behind TLS in production.

## Operational musts

- HTTPS only in production (reverse proxy such as Caddy/nginx terminating TLS, or
  `ListenAndServeTLS` with a Let's Encrypt cert). Bearer tokens are sniffable on plain HTTP.
- Rotation = change `API_TOKEN` in `.env`, restart server, rebuild the app. No state, no DB.
