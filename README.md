# Agro API

PocketBase-extended HTTP server with custom session-based authentication.

## Auth Endpoints

Method | Path | Description
--- | --- | ---
POST | /api/auth/login | Authenticate with email/password and set a session cookie
GET | /api/auth/verify | Validate the current session cookie
POST | /api/auth/logout | Destroy the session cookie

## Configuration Loading

The app loads configuration in this order:

1. Base file (default: config.yaml)
2. Optional local override file with the same name pattern:
   config.local.yaml

Example:
- If base is config.yaml, the app also looks for config.local.yaml
- Values in the local file override only the fields they define

This behavior is implemented in config.go.

## Server Bind Address and HTTP Fallback

When running serve, the app injects the HTTP bind address from config server.host + server.port only if you did not provide --http manually.

- No --http provided:
  the app appends --http host:port from config
- --http provided:
  your CLI value wins

## Development

Install dependencies:
    go mod download

Run with default base config resolution:
    go run . serve

Run with explicit config base file:
    go run . serve --config=config.yaml

Run with manual HTTP override (takes precedence):
    go run . serve --http=localhost:8092

Run tests:
    go test ./...

## Production

Build:
    go build -o agro-api .

Run with default config loading:
    ./agro-api serve

Run with explicit base config:
    ./agro-api serve --config=config.yaml

Run with explicit HTTP bind override:
    ./agro-api serve --http=0.0.0.0:8090

## Config Files

- config.yaml: shared/base config committed to repo
- config.local.yaml: local machine overrides (typically not committed)

## Migrations

Apply migrations:
    ./agro-api migrate up

Rollback migrations:
    ./agro-api migrate down
