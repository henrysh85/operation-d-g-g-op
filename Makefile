.DEFAULT_GOAL := help

COMPOSE ?= docker compose

.PHONY: help up down logs backend-dev frontend-dev migrate-up migrate-down seed lint test typecheck build reset

help: ## Show this help
	@echo "DCGG Intelligence Platform - Make targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

up: ## Start all services via docker compose
	$(COMPOSE) up -d

down: ## Stop all services
	$(COMPOSE) down

logs: ## Tail logs from all services
	$(COMPOSE) logs -f --tail=200

backend-dev: ## Run backend with hot reload (air)
	cd backend && air

frontend-dev: ## Run frontend dev server
	cd frontend && npm run dev

migrate-up: ## Apply database migrations
	cd backend && go run ./cmd/migrate up

migrate-down: ## Roll back last database migration
	cd backend && go run ./cmd/migrate down

seed: ## Seed DB from prototype HTML
	cd backend && go run ./cmd/importer --source ../prototype/DCGG_Intelligence_Platform_v7.html

lint: ## Lint backend + frontend
	cd backend && go vet ./...
	cd frontend && npm run lint

test: ## Run backend + frontend tests
	cd backend && go test ./...
	cd frontend && npm test --if-present

typecheck: ## Type-check frontend
	cd frontend && npm run typecheck

build: ## Build all docker images
	$(COMPOSE) build

reset: ## Wipe volumes and restart
	$(COMPOSE) down -v
	$(COMPOSE) up -d
