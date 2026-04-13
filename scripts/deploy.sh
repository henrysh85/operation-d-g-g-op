#!/usr/bin/env bash
#
# DCGG Intelligence Platform — deploy script
#
# Modes:
#   local       Build + run the full stack locally via docker compose (default)
#   images      Build backend+frontend images and push to GHCR
#   remote      Deploy to a remote host over SSH (rsync compose files, pull, up -d, migrate, seed)
#
# Usage:
#   ./scripts/deploy.sh [local|images|remote] [--tag vX.Y.Z] [--host user@server] [--path /srv/dcgg] [--no-seed]
#
# Env / flags:
#   DEPLOY_HOST       SSH target (user@host)  — required for remote
#   DEPLOY_PATH       Remote install dir       — default: /srv/dcgg
#   IMAGE_TAG         Docker image tag         — default: git short sha
#   GHCR_OWNER        GHCR namespace           — default: henrysh85
#   GHCR_REPO         GHCR repo name           — default: operation-d-g-g-op
#   SKIP_SEED=1       Skip importer seed step
#   SKIP_MIGRATE=1    Skip migration step (migrations also auto-run on backend start)
#
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

# ---------- defaults ----------
MODE="${1:-local}"; shift || true
IMAGE_TAG="${IMAGE_TAG:-$(git rev-parse --short HEAD 2>/dev/null || echo latest)}"
GHCR_OWNER="${GHCR_OWNER:-henrysh85}"
GHCR_REPO="${GHCR_REPO:-operation-d-g-g-op}"
DEPLOY_HOST="${DEPLOY_HOST:-}"
DEPLOY_PATH="${DEPLOY_PATH:-/srv/dcgg}"
SKIP_SEED="${SKIP_SEED:-0}"
SKIP_MIGRATE="${SKIP_MIGRATE:-0}"

# ---------- flag parsing ----------
while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag)       IMAGE_TAG="$2"; shift 2 ;;
    --host)      DEPLOY_HOST="$2"; shift 2 ;;
    --path)      DEPLOY_PATH="$2"; shift 2 ;;
    --no-seed)   SKIP_SEED=1; shift ;;
    --no-migrate) SKIP_MIGRATE=1; shift ;;
    -h|--help)
      sed -n '2,20p' "$0"; exit 0 ;;
    *) echo "unknown flag: $1" >&2; exit 2 ;;
  esac
done

BACKEND_IMAGE="ghcr.io/${GHCR_OWNER}/${GHCR_REPO}-backend"
FRONTEND_IMAGE="ghcr.io/${GHCR_OWNER}/${GHCR_REPO}-frontend"

log()  { printf '\033[1;36m[deploy]\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[warn]\033[0m %s\n' "$*" >&2; }
fail() { printf '\033[1;31m[fail]\033[0m %s\n' "$*" >&2; exit 1; }

require() { command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"; }

ensure_env() {
  if [[ ! -f .env ]]; then
    [[ -f .env.example ]] || fail ".env and .env.example both missing"
    warn ".env not found — copying from .env.example (edit secrets before production use)"
    cp .env.example .env
  fi
}

# ---------- modes ----------

deploy_local() {
  require docker
  ensure_env
  log "building images (tag=${IMAGE_TAG})"
  docker compose build
  log "starting stack"
  docker compose up -d
  log "waiting for backend health"
  for i in {1..30}; do
    if curl -fsS http://localhost:8080/healthz >/dev/null 2>&1; then
      log "backend healthy"; break
    fi
    sleep 2
    [[ $i -eq 30 ]] && fail "backend did not become healthy in 60s"
  done

  if [[ "$SKIP_MIGRATE" != "1" ]]; then
    log "running migrations"
    docker compose exec -T backend /app/api migrate up || warn "migrate command failed (may already be applied or auto-run)"
  fi

  if [[ "$SKIP_SEED" != "1" ]]; then
    log "seeding from prototype HTML"
    docker compose exec -T backend /app/importer || warn "seed step failed (importer may be a skeleton)"
  fi

  log "done. frontend: http://localhost  |  api: http://localhost:8080  |  minio console: http://localhost:9001"
}

deploy_images() {
  require docker
  require git
  log "building + pushing images to GHCR (tag=${IMAGE_TAG})"
  : "${GHCR_TOKEN:?set GHCR_TOKEN (PAT with write:packages) or run 'docker login ghcr.io' first}"
  echo "$GHCR_TOKEN" | docker login ghcr.io -u "$GHCR_OWNER" --password-stdin

  docker build -t "${BACKEND_IMAGE}:${IMAGE_TAG}" -t "${BACKEND_IMAGE}:latest" ./backend
  docker build -t "${FRONTEND_IMAGE}:${IMAGE_TAG}" -t "${FRONTEND_IMAGE}:latest" ./frontend

  docker push "${BACKEND_IMAGE}:${IMAGE_TAG}"
  docker push "${BACKEND_IMAGE}:latest"
  docker push "${FRONTEND_IMAGE}:${IMAGE_TAG}"
  docker push "${FRONTEND_IMAGE}:latest"

  log "pushed:"
  log "  ${BACKEND_IMAGE}:${IMAGE_TAG}"
  log "  ${FRONTEND_IMAGE}:${IMAGE_TAG}"
}

deploy_remote() {
  require rsync
  require ssh
  [[ -n "$DEPLOY_HOST" ]] || fail "remote mode requires --host user@server or DEPLOY_HOST env"
  [[ -f .env ]] || fail "local .env is required — it will be rsynced to the remote host"

  log "syncing compose bundle to ${DEPLOY_HOST}:${DEPLOY_PATH}"
  ssh "$DEPLOY_HOST" "mkdir -p '${DEPLOY_PATH}'"
  rsync -az --delete \
    --include='docker-compose.yml' \
    --include='.env' \
    --include='Makefile' \
    --include='prototype/***' \
    --include='scripts/***' \
    --exclude='backend/***' --exclude='frontend/***' --exclude='.git/***' --exclude='node_modules/***' \
    --exclude='*' \
    ./ "${DEPLOY_HOST}:${DEPLOY_PATH}/"

  # compose override pinning to GHCR images
  cat <<EOF | ssh "$DEPLOY_HOST" "cat > ${DEPLOY_PATH}/docker-compose.override.yml"
services:
  backend:
    image: ${BACKEND_IMAGE}:${IMAGE_TAG}
    build: null
  frontend:
    image: ${FRONTEND_IMAGE}:${IMAGE_TAG}
    build: null
EOF

  log "pulling images and restarting services on ${DEPLOY_HOST}"
  ssh "$DEPLOY_HOST" "cd '${DEPLOY_PATH}' && docker compose pull && docker compose up -d --remove-orphans"

  if [[ "$SKIP_MIGRATE" != "1" ]]; then
    log "running migrations on remote"
    ssh "$DEPLOY_HOST" "cd '${DEPLOY_PATH}' && docker compose exec -T backend /app/api migrate up" || warn "migrate failed (may already be applied)"
  fi

  if [[ "$SKIP_SEED" != "1" ]]; then
    log "seeding on remote"
    ssh "$DEPLOY_HOST" "cd '${DEPLOY_PATH}' && docker compose exec -T backend /app/importer" || warn "seed failed"
  fi

  log "remote deploy complete (${DEPLOY_HOST}:${DEPLOY_PATH})"
}

# ---------- dispatch ----------
case "$MODE" in
  local)  deploy_local  ;;
  images) deploy_images ;;
  remote) deploy_remote ;;
  *) fail "unknown mode: $MODE (expected local|images|remote)" ;;
esac
