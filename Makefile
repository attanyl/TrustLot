.PHONY: help dev-up dev-down build build-api build-worker build-synth run-api run-worker web-install web-dev test lint fmt migrate migrate-down seed db-reset clean

DATABASE_URL ?= postgres://trustlot:trustlot@localhost:5432/trustlot?sslmode=disable

GO := go
GOFLAGS :=
BIN := bin

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# --- Infrastructure ---

dev-up: ## Start local infrastructure (postgres, redis)
	docker compose up -d

dev-down: ## Stop local infrastructure
	docker compose down

# --- Go builds ---

build: build-api build-worker build-synth ## Build all Go binaries

build-api: ## Build the API server
	$(GO) build $(GOFLAGS) -o $(BIN)/api ./cmd/api

build-worker: ## Build the background worker
	$(GO) build $(GOFLAGS) -o $(BIN)/worker ./cmd/worker

build-synth: ## Build the synthetic data generator
	$(GO) build $(GOFLAGS) -o $(BIN)/synth ./tools/synth

# --- Run ---

run-api: ## Run the API server
	$(GO) run ./cmd/api

run-worker: ## Run the background worker
	$(GO) run ./cmd/worker

# --- Database ---

migrate: ## Run database migrations up
	migrate -path db/migrations -database "$(DATABASE_URL)" up

migrate-down: ## Roll back one migration
	migrate -path db/migrations -database "$(DATABASE_URL)" down 1

seed: ## Seed the database with synthetic data
	psql "$(DATABASE_URL)" -f db/seed.sql

db-reset: ## Drop all tables and re-migrate + seed
	migrate -path db/migrations -database "$(DATABASE_URL)" drop -f
	$(MAKE) migrate
	$(MAKE) seed

# --- Frontend ---

web-install: ## Install frontend dependencies
	cd apps/web && npm install

web-dev: ## Start frontend dev server
	cd apps/web && npm run dev

# --- Quality ---

test: ## Run all Go tests
	$(GO) test ./...

lint: ## Run Go vet
	$(GO) vet ./...

fmt: ## Format Go code
	gofmt -s -w .

# --- Cleanup ---

clean: ## Remove build artifacts
	rm -rf $(BIN)
