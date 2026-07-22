# Go Clean Architecture Demo - v2 Enterprise Stack
# Build / lint / test / proto / migrate orchestration.

# ---- Variables ----------------------------------------------------------------
BIN_DIR     := bin
APP_NAME    := go-clean-arch-demo
GO          := go
GOFLAGS     := -trimpath
LDFLAGS     := -s -w
PKG         := ./...

# Tooling versions (downloaded by `make install-tools`)
GOLANGCI_VERSION := 1.61.0
SWAG_VERSION     := 1.16.3
MIGRATE_VERSION  := 4.17.1
BUF_VERSION      := 1.34.0

OS := $(shell uname -s | tr A-Z a-z)
ARCH := $(shell uname -m)
ifeq ($(ARCH),x86_64)
	ARCH := amd64
endif
ifeq ($(ARCH),aarch64)
	ARCH := arm64
endif

# ---- Default target -----------------------------------------------------------
.PHONY: help
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

# ---- Build --------------------------------------------------------------------
.PHONY: build
build: build-rest build-grpc ## Build both REST and gRPC binaries

.PHONY: build-rest
build-rest: ## Build the REST server binary
	@ mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/rest ./cmd/rest

.PHONY: build-grpc
build-grpc: ## Build the gRPC server binary
	@ mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BIN_DIR)/grpc ./cmd/grpc

.PHONY: build-race
build-race: ## Build both binaries with -race flag
	$(GO) build $(GOFLAGS) -race -o $(BIN_DIR)/rest ./cmd/rest
	$(GO) build $(GOFLAGS) -race -o $(BIN_DIR)/grpc ./cmd/grpc

# ---- Test ---------------------------------------------------------------------
.PHONY: test
test: ## Run unit tests with race detector + coverage
	$(GO) test -race -coverprofile=coverage.out -timeout 30s $(PKG)

.PHONY: test-cover
test-cover: test ## Run tests and open HTML coverage report
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# ---- Lint ---------------------------------------------------------------------
.PHONY: lint
lint: golangci ## Run golangci-lint with .golangci.yaml
	$(BIN_DIR)/golangci-lint version
	$(BIN_DIR)/golangci-lint run -c .golangci.yaml $(PKG)

# ---- Codegen ------------------------------------------------------------------
.PHONY: proto
proto: buf ## Generate protobuf Go code via buf
	$(BIN_DIR)/buf generate

.PHONY: swagger
swagger: swag ## Generate Swagger OpenAPI docs from annotations
	$(BIN_DIR)/swag init -g cmd/rest/main.go -o docs/swagger --parseDependency --parseInternal

.PHONY: mocks
mocks: ## Generate mocks for testing (mockery)
	$(GO) generate ./...

# ---- Database migrations ------------------------------------------------------
MIGRATIONS_DIR := internal/infrastructure/persistence/migrations
DB_DSN ?= mysql://$(DB_USER):$(DB_PASS)@tcp($(DB_HOST):$(DB_PORT))/$(DB_NAME)

.PHONY: migrate-up
migrate-up: migrate ## Apply all pending migrations
	$(BIN_DIR)/migrate -database $(DB_DSN) -path $(MIGRATIONS_DIR) up

.PHONY: migrate-down
migrate-down: migrate ## Roll back the last migration
	$(BIN_DIR)/migrate -database $(DB_DSN) -path $(MIGRATIONS_DIR) down 1

.PHONY: migrate-create
migrate-create: migrate ## Create a new migration pair (NAME=foo_bar). Falls back to interactive prompt if NAME unset.
	@if [ -z "$$NAME" ]; then \
		read -p "Migration name: " NAME; \
	fi; \
	$(BIN_DIR)/migrate create -ext sql -dir $(MIGRATIONS_DIR) $$NAME

# ---- Dev environment ----------------------------------------------------------
.PHONY: dev-env
dev-env: ## Start MySQL via docker-compose
	docker compose -f deployments/docker-compose.yaml up -d mysql

.PHONY: dev-air
dev-air: air ## Run with hot reload (requires MySQL already up)
	$(BIN_DIR)/air -c .air.toml

.PHONY: run-rest
run-rest: build-rest ## Build and run REST server
	./$(BIN_DIR)/rest

.PHONY: run-grpc
run-grpc: build-grpc ## Build and run gRPC server
	./$(BIN_DIR)/grpc

# ---- Docker -------------------------------------------------------------------
.PHONY: image-build
image-build: ## Build Docker image
	docker build -f deployments/Dockerfile -t $(APP_NAME):latest .

.PHONY: image-run
image-run: image-build ## Build and run full stack via docker-compose
	docker compose -f deployments/docker-compose.yaml up --build

# ---- Clean --------------------------------------------------------------------
.PHONY: clean
clean: clean-artifacts clean-bin ## Remove build artifacts and bin/

.PHONY: clean-artifacts
clean-artifacts: ## Remove coverage / output files
	rm -f coverage.out coverage.html

.PHONY: clean-bin
clean-bin: ## Remove bin/ directory
	rm -rf $(BIN_DIR)

# ---- Tool installation (local) ------------------------------------------------
.PHONY: install-tools
install-tools: golangci swag migrate buf air ## Install all dev tools to bin/

.PHONY: golangci
golangci: $(BIN_DIR)/golangci-lint
$(BIN_DIR)/golangci-lint:
	@ mkdir -p $(BIN_DIR)
	curl -sSL https://github.com/golangci/golangci-lint/releases/download/v$(GOLANGCI_VERSION)/golangci-lint-$(GOLANGCI_VERSION)-$(OS)-$(ARCH).tar.gz | tar -zOxf - golangci-lint-$(GOLANGCI_VERSION)-$(OS)-$(ARCH)/golangci-lint > $@ && chmod +x $@

.PHONY: swag
swag: $(BIN_DIR)/swag
$(BIN_DIR)/swag:
	@ mkdir -p $(BIN_DIR)
	GOBIN=$(PWD)/$(BIN_DIR) $(GO) install github.com/swaggo/swag/cmd/swag@v$(SWAG_VERSION)

.PHONY: migrate
migrate: $(BIN_DIR)/migrate
$(BIN_DIR)/migrate:
	@ mkdir -p $(BIN_DIR)
	curl -sSL https://github.com/golang-migrate/migrate/releases/download/v$(MIGRATE_VERSION)/migrate.$(OS)-$(ARCH).tar.gz | tar -zOxf - migrate > $@ && chmod +x $@

.PHONY: buf
buf: $(BIN_DIR)/buf
$(BIN_DIR)/buf:
	@ mkdir -p $(BIN_DIR)
	curl -sSL https://github.com/bufbuild/buf/releases/download/v$(BUF_VERSION)/buf-$(OS)-$(ARCH) > $@ && chmod +x $@

.PHONY: air
air: $(BIN_DIR)/air
$(BIN_DIR)/air:
	@ mkdir -p $(BIN_DIR)
	GOBIN=$(PWD)/$(BIN_DIR) $(GO) install github.com/air-verse/air@v$(AIR_VERSION)

# DB connection for migrate-* targets.
# These read from APP_DATABASE_* env vars (set by Viper / CI workflow) so
# the Makefile targets work in lockstep with the application's config.
# Override on command line: make DB_HOST=other migrate-up
DB_HOST  ?= $(or $(APP_DATABASE_HOST),127.0.0.1)
DB_PORT  ?= $(or $(APP_DATABASE_PORT),3306)
DB_USER  ?= $(or $(APP_DATABASE_USER),app)
DB_PASS  ?= $(or $(APP_DATABASE_PASSWORD),app)
DB_NAME  ?= $(or $(APP_DATABASE_NAME),article)
AIR_VERSION := 1.52.0
