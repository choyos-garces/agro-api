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
```shell
go mod download
```
Run with default base config resolution:
```shell
go run . serve
```
Run tests:
```shell
go test ./...
```

## Production
Output the right binary for your target platform:
```shell
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o api-bin .
```

Desploys the binary to the server and restarts the service:
```shell
ssh vps "systemctl stop api-bin"
scp api-bin vps:/var/www/agro-api/
ssh vps "systemctl start api-bin"
```

## Config Files

- config.yaml: shared/base config committed to repo
- config.local.yaml: local machine overrides (typically not committed)

## Migrations

Apply migrations:
    ./agro-api migrate up

Rollback migrations:
    ./agro-api migrate down
