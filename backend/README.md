# DCGG Intelligence Platform — Backend

Go 1.22 / Gin / pgx / MinIO backend for the DCGG Intelligence Platform.

## Layout

```
cmd/
  api/         HTTP server entrypoint
  importer/    CLI that seeds Postgres from the v7 HTML prototype
internal/
  config/      env-backed configuration
  db/          pgx pool + golang-migrate runner
  db/migrations/ numbered up/down SQL files
  auth/        JWT issuance + middleware + HR PIN gate
  storage/     MinIO client
  models/      domain structs
  repo/        data access per aggregate
  handlers/    Gin handlers
  router/      HTTP wiring
```

## Environment

Required:

| var | purpose |
|-----|---------|
| `DB_URL` | Postgres DSN (e.g. `postgres://user:pass@localhost:5432/dcgg?sslmode=disable`) |
| `MINIO_ENDPOINT` | e.g. `localhost:9000` |
| `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | credentials |
| `MINIO_BUCKET` | default `dcgg` |
| `MINIO_USE_SSL` | `true`/`false` |
| `JWT_SECRET` | HMAC secret |
| `HR_PIN` | 6-digit HR-gate PIN |
| `PORT` | default `8080` |
| `ENV` | `development` or `production` |

Put them in `.env` (loaded via godotenv) or export directly.

## Running

```bash
# deps (first time)
go mod tidy

# live-reload dev
go install github.com/cosmtrek/air@latest
air

# one-shot
go run ./cmd/api
```

Migrations auto-apply on server start. To run manually:

```bash
migrate -path internal/db/migrations -database "$DB_URL" up
migrate -path internal/db/migrations -database "$DB_URL" down 1
```

## Importer

```bash
go run ./cmd/importer -file ../prototype/DCGG_Intelligence_Platform_v7.html -dry=true
```

Currently a skeleton: it locates the seven JS seed blocks and logs them. See
the header comment in `cmd/importer/main.go` for the extraction strategy
(node one-liner or in-Go tokenizer + key-quoting pass).

## Docker

```bash
docker build -t dcgg-backend .
docker run --rm -p 8080:8080 --env-file .env dcgg-backend
```

## API surface (v1)

- `POST /api/v1/auth/login`
- `GET  /api/v1/auth/me`
- `POST /api/v1/auth/verify-pin`
- `GET  /api/v1/dashboard/summary`
- `/people`, `/activities`, `/clients`, `/regulatory/*`, `/stakeholders/*`,
  `/consultations`, `/templates`, `/publications`, `/files/*`,
  `/memberships/generate` — full CRUD where applicable, filterable by
  `vertical`, `region_id`, `client_id`, `from`/`to`, `q`, etc.

HR-only routes require role `hr` or `admin` and a PIN-gated token issued by
`/auth/verify-pin`.
