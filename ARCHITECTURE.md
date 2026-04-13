# Architecture

This document describes the internal architecture of the DCGG Intelligence Platform: data flow, authentication, file handling, RBAC, and scaling considerations.

## High-Level Topology

```
  Browser (Vue 3 SPA)
       | HTTPS, JSON, Bearer JWT
       v
  Gin HTTP API (Go 1.22)
       |   \
   pgx |    \ minio-go (S3)
       v     v
  Postgres  MinIO
```

All application state lives in Postgres. All binary artifacts (uploaded files, generated documents, avatars, membership PDFs) live in MinIO. The backend is stateless and horizontally scalable.

## Request Lifecycle

1. Browser calls `VITE_API_BASE` (configured at build time) with a bearer JWT.
2. Gin middleware chain:
   - `RequestID` — correlates logs and audit entries.
   - `Logger` / `Recover` — structured logs, panic safety.
   - `CORS` — restricted to known origins in production.
   - `Auth` — verifies JWT, loads user + role claims.
   - `RBAC` — per-route role check.
   - `HRGate` — additional PIN check on HR Vault routes only.
3. Handler calls a service; service uses a `Store` (pgx) and/or `Blobs` (minio-go).
4. Response is JSON. Errors follow `{ "error": { "code", "message" } }`.

## Data Flow

- **Write path:** handler → validate DTO → service (transactional) → `store.Tx` → commit → optional async notification.
- **Read path:** handler → service → store query → DTO projection → JSON.
- **Search (Intelligence Library, Radar):** Postgres full-text (`tsvector`) with per-module materialized views; future: ship to an external search service if corpus grows past single-node comfort.

## Authentication & Authorization

### JWT

- HS256 signed with `JWT_SECRET`.
- Claims: `sub` (user id), `role`, `exp`, `iat`, `jti`.
- Short-lived access token (15 min); refresh token stored httpOnly.
- Rotation on refresh; `jti` denylist on logout.

### HR PIN Gate

HR Vault endpoints require a **second factor**: the `HR_PIN` (shared by authorized HR staff, rotated out-of-band). The PIN is submitted once per session, exchanged for a short-lived `hr_scope` token, and required on every HR route. This protects personnel data even if a session token is compromised.

### RBAC Roles

| Role       | Capabilities                                                                |
|------------|-----------------------------------------------------------------------------|
| `admin`    | Full access, user management, audit log, system settings.                   |
| `lead`     | Read + write across all non-HR modules, can manage working groups.          |
| `staff`    | Read + write within assigned modules; no admin surfaces.                    |
| `readonly` | Read-only across non-HR modules.                                            |
| `hr`       | Read + write on HR Vault only (after PIN gate); no other modules.           |

Role is stored on the `users` table; per-resource ACLs (e.g., working-group membership) are joined in at query time.

## File Upload Flow (Presigned URLs)

The backend never proxies file bytes. Instead:

1. Client `POST /api/v1/uploads` with `{ bucket, filename, contentType }`.
2. Backend validates role + bucket, generates an `object_key` (`<module>/<yyyy>/<uuid>-<slug>`), issues a presigned `PUT` URL (short TTL, e.g. 15 min).
3. Client `PUT`s the bytes directly to MinIO.
4. Client `POST /api/v1/uploads/:id/commit` to register the object (size, checksum, linked entity).
5. Downloads use presigned `GET` URLs issued on demand.

Buckets:

- `dcgg-outputs` — generated briefings, exports.
- `dcgg-avatars` — user and stakeholder avatars (publicly readable).
- `dcgg-templates` — document templates for Content Studio.
- `dcgg-memberships` — signed membership agreements and onboarding docs.

## Audit

Every state-changing request writes to `audit_log` (actor, action, entity, before/after hash, request id, ip, ua). HR actions are tagged and retained separately per policy.

## Migrations

Versioned SQL migrations in `backend/migrations/` applied by `cmd/migrate`. CI runs them against an ephemeral Postgres. Never edit an applied migration — always add a new one.

## Observability

- Structured JSON logs (zerolog) with `request_id`.
- `/healthz` (liveness) and `/readyz` (DB + MinIO ping).
- Metrics: Prometheus `/metrics` (HTTP latency, DB pool, MinIO errors).

## Scaling Notes

- Backend is stateless — scale horizontally behind a load balancer.
- Postgres: start single primary; add read replicas once read load demands it. Use `pgbouncer` in front for connection pooling.
- MinIO: deploy in distributed mode (>= 4 nodes) for production; keep buckets immutable where possible, enable versioning on `dcgg-memberships`.
- Background jobs (importer, notifications, PDF generation): move from in-process workers to a dedicated worker pool + queue (e.g., Postgres `LISTEN/NOTIFY` or NATS) when throughput requires.
- Frontend: static assets served via nginx / CDN; SPA is fully cacheable by hash.
- Secrets: `.env` locally; SOPS / SSM / Vault in deployed environments. Never commit `.env`.
