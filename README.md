# Camera API

A Go backend API for managing PTZ cameras and their recordings, with support for **ONVIF**, **Hikvision ISAPI** and **S3-compatible storage**.

## Features

- JWT auth for API (Bearer) and web (HTTP-only cookie) clients, with session revocation and temporary account lock
- Role-based access with per-user camera access grants and an activity log
- Camera management and configuration; PTZ move, presets, home and zoom
- Recording search/download and snapshots, directly or through S3
- Rate limiting, Vault-based configuration, Swagger, Docker

```text
Echo → Middleware (auth + session check, roles, rate limit) → Handler → Usecase
                                        ├── Repository → PostgreSQL (GORM)
                                        ├── Camera adapter → ONVIF / ISAPI
                                        └── Storage adapter → S3 / MinIO
```

**Requirements:** Go 1.25+, PostgreSQL, HashiCorp Vault (KV v2), S3-compatible storage, Docker (optional).

## Getting started

1. **Generate the Swagger docs.** The app imports the generated `docs` package, which is not committed:
   ```bash
   make swagger-install && make swagger-init
   ```
2. **Store the config in Vault** at `<PRODUCT>/data/<ENV>/<FILE>` (e.g. `camera/data/development/api`), based on [`docs/vault.example.json`](docs/vault.example.json).
3. **Create `.env`** with `cp .env.example .env`. `HOST` and `TOKEN` are **encoded** (base64, then braille), not plain text. Encode them with `vault.Lock`:
   ```bash
   mkdir -p .tmp/lock && cat > .tmp/lock/main.go <<'EOF'
   package main

   import ("fmt"; "os"; "camera_api/pkg/adapter/vault")

   func main() { s, _ := vault.Lock(os.Args[1]); fmt.Println(s) }
   EOF
   go run ./.tmp/lock/main.go "http://127.0.0.1:8200"   # HOST
   go run ./.tmp/lock/main.go "<vault token>"           # TOKEN
   rm -rf .tmp
   ```
4. **Run.** `make dev` (uses `ENV=development` and the root `.env`), or with Docker:
   ```bash
   cp docker/example.docker-compose.yml docker/docker-compose.yml
   export HOST=<encoded> TOKEN=<encoded>   # Compose reads these from the shell or docker/.env
   make prod                               # ENV=production; also: make ps, make logs
   ```

## Database

With `database.migrate: true` in the Vault config, the schema and indexes are managed automatically. With `false` (typical in production), apply these **before** deploying:

```sql
ALTER TABLE users ADD COLUMN IF NOT EXISTS locked_until timestamp(0);          -- required by login
CREATE UNIQUE INDEX IF NOT EXISTS idx_camera_access_camera_user
  ON camera_access (camera_id, user_id) WHERE deleted_at IS NULL;
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_sessions_user_last_seen
  ON sessions (user_id, last_seen_at);                                         -- performance
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_activities_actor_created
  ON user_activities (actor_user_id, created_at);                              -- performance
```

## Protocol support

- **ONVIF:** capabilities, configuration (profiles), position, directional and calibration moves.
- **ISAPI:** all of the above plus presets, home, zoom, storage, recording search/download and snapshots.

Moves are limited to 10 seconds.

## Testing

```bash
gofmt -l .           # must print nothing
go vet ./...
go test -race ./...
```

Integration tests run only when configured: `DATABASE_DSN` (PostgreSQL, no migrations) and `HOST`/`TOKEN` (Vault, reads `camera/data/development/api`, never prints secrets).

## License

[MIT](LICENSE)
