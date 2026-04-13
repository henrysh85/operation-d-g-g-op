# DCGG Intelligence Platform

Regulatory intelligence and operations platform for the **Digital Currency Governance Group (DCGG)** — a coalition whose founding members include **Tether**, **TRON**, **Ledger**, and **Wintermute**. The platform unifies policy tracking, stakeholder engagement, member operations, and internal workflows behind a single secure interface.

Upstream repository: https://github.com/henrysh85/operation-d-g-g-op

---

## Architecture

```
                +-----------------------+
                |     Browser (SPA)     |
                |  Vue 3 / TS / Pinia   |
                +----------+------------+
                           | HTTPS / JSON
                           v
                +-----------------------+
                |      Gin API (Go)     |
                |   JWT auth / RBAC     |
                +---+---------------+---+
                    |               |
             SQL    |               |  S3 API
                    v               v
          +-------------------+  +-------------------+
          |   Postgres 16     |  |      MinIO        |
          |  primary store    |  | objects / uploads |
          +-------------------+  +-------------------+
```

## Stack

- **Frontend:** Vue 3, TypeScript, Pinia, Vue Router, Tailwind CSS, Vite
- **Backend:** Go 1.22, Gin, pgx (Postgres driver), minio-go, golang-jwt
- **Data:** Postgres 16
- **Objects:** MinIO (S3-compatible) — outputs, avatars, templates, memberships
- **Infra:** Docker Compose for local, GitHub Actions for CI, GHCR for images

## Modules

1. **Dashboard** — unified KPIs, priorities, and alerts across all modules.
2. **Regulatory Radar** — global policy/regulation tracker with jurisdictional filtering.
3. **Stakeholders** — contacts, organizations, and relationship graph.
4. **Engagements** — meetings, positions, talking points, follow-ups.
5. **Members** — founding and associate member directory and onboarding.
6. **Working Groups** — cross-member task forces, deliverables, minutes.
7. **Content Studio** — briefings, statements, template-driven outputs.
8. **Events** — roundtables, summits, registrations.
9. **Intelligence Library** — research, filings, memos with tagging and search.
10. **HR Vault** — gated personnel records (HR PIN protected).
11. **Admin & Audit** — RBAC, user management, audit log, system settings.

## Local Setup

Prerequisites: Docker, Docker Compose, `make`. For running outside containers: Go 1.22+, Node 20+.

```bash
cp .env.example .env
make up        # starts postgres, minio, createbuckets, backend, frontend
make seed      # imports data from prototype HTML
```

Services:

| Service         | URL                          |
|-----------------|------------------------------|
| Frontend (dev)  | http://localhost:5173        |
| Frontend (prod) | http://localhost             |
| Backend API     | http://localhost:8080        |
| Postgres        | localhost:5432               |
| MinIO S3 API    | http://localhost:9000        |
| MinIO Console   | http://localhost:9001        |

Common commands:

```bash
make backend-dev    # air hot-reload on the Go server
make frontend-dev   # Vite dev server
make migrate-up     # apply DB migrations
make logs           # tail everything
make reset          # wipe volumes and restart fresh
```

## Repository Layout

```
dcgg-platform/
├── backend/                  # Go / Gin API
│   ├── cmd/                  # entrypoints (server, migrate, importer)
│   ├── internal/             # handlers, services, models, store
│   └── migrations/           # SQL migrations
├── frontend/                 # Vue 3 SPA
│   ├── src/                  # views, components, stores, router
│   └── public/
├── prototype/                # Original static prototype
│   └── DCGG_Intelligence_Platform_v7.html
├── .github/workflows/        # CI + container publish
├── docker-compose.yml
├── .env.example
├── Makefile
├── ARCHITECTURE.md
├── CONTRIBUTING.md
├── LICENSE
└── README.md
```

## Prototype Reference

The initial UX, module layout, and seed content all derive from the single-file prototype at `./prototype/DCGG_Intelligence_Platform_v7.html`. The importer (`make seed`) parses this file to pre-populate the database in development.

## Contribution Workflow

- Branch from `main`: `feat/<scope>`, `fix/<scope>`, `docs/<scope>`, `chore/<scope>`.
- Commits use [Conventional Commits](https://www.conventionalcommits.org/): `feat(radar): add jurisdiction filter`.
- PRs target `main`, require green CI + one reviewer.
- Run `make lint test typecheck` before opening a PR.

See [CONTRIBUTING.md](./CONTRIBUTING.md) and [ARCHITECTURE.md](./ARCHITECTURE.md) for details.

## License

MIT — see [LICENSE](./LICENSE).
