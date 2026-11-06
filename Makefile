SHELL := /bin/sh
BIN_DIR := bin
DIST_DIR := dist
APP_NAME := rock-sql
LOCAL_BIN := $(BIN_DIR)/$(APP_NAME)
OPENBSD_BIN := $(DIST_DIR)/openbsd_amd64/$(APP_NAME)

# Default Go environment
CGO_ENABLED ?= 0
export CGO_ENABLED

.PHONY: all help build build-openbsd run restart dev test vet fmt clean db-up db-down db-logs db-seed

all: build ## Build binary for current operating system (default)

help: ## Show this help message
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build binary for current platform
	@mkdir -p $(BIN_DIR)
	@echo "==> Building $(LOCAL_BIN)..."
	go build -trimpath -ldflags="-s -w" -o $(LOCAL_BIN) .
	@echo "==> Build complete: $(LOCAL_BIN)"

build-openbsd: ## Cross-compile static binary for OpenBSD (amd64)
	@mkdir -p $(DIST_DIR)/openbsd_amd64
	@echo "==> Cross-compiling for OpenBSD amd64..."
	GOOS=openbsd GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o $(OPENBSD_BIN) .
	@echo "==> OpenBSD binary ready: $(OPENBSD_BIN)"

run: build ## Build and run locally
	@echo "==> Starting $(LOCAL_BIN)..."
	./$(LOCAL_BIN)

restart: build ## Rebuild and restart running local rock-sql instance
	@echo "==> Stopping existing $(APP_NAME) processes..."
	-@pkill -f "$(LOCAL_BIN)" 2>/dev/null || true
	@sleep 1
	@echo "==> Starting $(LOCAL_BIN)..."
	./$(LOCAL_BIN)

dev: db-up build ## Start local PostgreSQL container and launch application
	@echo "==> Waiting for PostgreSQL to be healthy..."
	@docker compose ps
	@until [ "$$(docker inspect -f '{{.State.Health.Status}}' rock-sql-postgres 2>/dev/null)" = "healthy" ]; do \
		sleep 1; \
	done
	@echo "==> PostgreSQL is ready! Starting $(LOCAL_BIN)..."
	./$(LOCAL_BIN)

db-up: ## Start PostgreSQL service in background with seeds mounted
	@echo "==> Starting PostgreSQL container..."
	docker compose up -d postgres

db-down: ## Stop PostgreSQL container
	@echo "==> Stopping PostgreSQL container..."
	docker compose down

db-logs: ## Follow PostgreSQL container logs
	docker compose logs -f postgres

db-seed: ## Re-apply SQL seed files to the running PostgreSQL container
	@echo "==> Re-applying seeds..."
	@docker exec -i rock-sql-postgres psql -U explorer -d astronomy -f /seeds/astronomy_seed.sql
	@docker exec -i rock-sql-postgres psql -U explorer -d math -f /seeds/math_seed.sql
	@docker exec -i rock-sql-postgres psql -U explorer -d anatomy -f /seeds/anatomy_seed.sql
	@echo "==> Seeds applied successfully!"

test: ## Run unit tests
	@echo "==> Running tests..."
	go test -v ./...

vet: ## Run go vet
	@echo "==> Running go vet..."
	go vet ./...

fmt: ## Format Go source files
	@echo "==> Formatting Go files..."
	gofmt -w -s .

clean: ## Clean build artifacts
	@echo "==> Cleaning build artifacts..."
	rm -rf $(BIN_DIR) $(DIST_DIR)
	@echo "==> Clean complete."
