-include .env
export

APP_NAME   := main
DB_SSLMODE ?= disable
DB_URL     ?= postgres://$(BLUEPRINT_DB_USERNAME):$(BLUEPRINT_DB_PASSWORD)@$(BLUEPRINT_DB_HOST):$(BLUEPRINT_DB_PORT)/$(BLUEPRINT_DB_DATABASE)?sslmode=$(DB_SSLMODE)&search_path=$(BLUEPRINT_DB_SCHEMA)

# go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
SQLC  = sqlc
# go install github.com/pressly/goose/v3/cmd/goose@latest
GOOSE = GOOSE_DRIVER=postgres GOOSE_DBSTRING="$(DB_URL)" GOOSE_MIGRATION_DIR=migrations goose
#
LINT  = go run github.com/golangci/golangci-lint/cmd/golangci-lint@v1.62.2

# Build and test
all: build test

build:
	@echo "Building..."
	@CGO_ENABLED=0 go build -ldflags="-s -w" -o $(APP_NAME) cmd/api/main.go

run:
	@go run cmd/api/main.go

# Docker
docker-run:
	@if docker compose up --build 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose up --build; \
	fi

docker-down:
	@if docker compose down 2>/dev/null; then \
		: ; \
	else \
		echo "Falling back to Docker Compose V1"; \
		docker-compose down; \
	fi

docker-build:
	@docker build -t $(APP_NAME):latest .

# Tests
test:
	@echo "Testing..."
	@go test ./... -v

# Integration tests (testcontainers: database + repositories)
itest:
	@echo "Running integration tests..."
	@go test ./internal/database ./internal/article/... -v

lint:
	@$(LINT) run ./...

# sqlc
sqlc:
	@$(SQLC) generate

sqlc-check:
	@$(SQLC) diff

sqlc-vet:
	@$(SQLC) vet

# Migrations (goose)
migrate-up:
	@$(GOOSE) up

migrate-down:
	@$(GOOSE) down

migrate-status:
	@$(GOOSE) status

migrate-reset:
	@$(GOOSE) reset

# usage: make migrate-create name=add_users
migrate-create:
	@$(GOOSE) create $(name) sql -s

# Clean
clean:
	@echo "Cleaning..."
	@rm -f $(APP_NAME)

# Live reload
watch:
	@if command -v air > /dev/null; then \
            air; \
            echo "Watching...";\
        else \
            read -p "Go's 'air' is not installed on your machine. Do you want to install it? [Y/n] " choice; \
            if [ "$$choice" != "n" ] && [ "$$choice" != "N" ]; then \
                go install github.com/air-verse/air@latest; \
                air; \
                echo "Watching...";\
            else \
                echo "You chose not to install air. Exiting..."; \
                exit 1; \
            fi; \
        fi

.PHONY: all build run docker-run docker-down docker-build test itest lint \
	sqlc sqlc-check sqlc-vet \
	migrate-up migrate-down migrate-status migrate-reset migrate-create \
	clean watch
