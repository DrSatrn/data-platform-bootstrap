# This Makefile centralizes the intended developer workflow for the platform.
# The recipes currently emphasize clear, teachable entrypoints over clever shell
# tricks so the build and runtime model stays easy to understand.

SHELL := /bin/sh
COMPOSE_ENV_FILE := $(if $(wildcard .env.compose),--env-file .env.compose,)

.PHONY: doctor fmt lint test build backend-build web-build up down smoke compose-smoke bootstrap benchmark backup restore-drill

doctor:
	@echo "Review codex.md before first build."
	@echo "Verify Go, Node, Python 3, Docker/OrbStack, and ARM64-compatible container images."
	@echo "Verify host C/C++ build tools are installed because the DuckDB Go driver uses CGO."

fmt:
	cd backend && gofmt -w $$(find . -name '*.go' -print)

lint:
	cd backend && go test ./...
	cd backend && go run ./cmd/platformctl validate-manifests
	cd web && npm run build

test:
	cd backend && go test ./...

build: backend-build web-build

backend-build:
	cd backend && GOOS=darwin GOARCH=arm64 go build -o bin/platform-api ./cmd/platform-api
	cd backend && GOOS=darwin GOARCH=arm64 go build -o bin/platform-scheduler ./cmd/platform-scheduler
	cd backend && GOOS=darwin GOARCH=arm64 go build -o bin/platform-worker ./cmd/platform-worker
	cd backend && GOOS=darwin GOARCH=arm64 go build -o bin/platformctl ./cmd/platformctl

web-build:
	cd web && npm run build

up:
	docker compose $(COMPOSE_ENV_FILE) -f infra/compose/docker-compose.yml up --build

down:
	docker compose $(COMPOSE_ENV_FILE) -f infra/compose/docker-compose.yml down

smoke:
	sh infra/scripts/localhost_smoke.sh

compose-smoke:
	sh infra/scripts/compose_smoke.sh

benchmark:
	sh infra/scripts/benchmark_suite.sh

backup:
	sh infra/scripts/backup_snapshot.sh

restore-drill:
	sh infra/scripts/restore_drill.sh

bootstrap:
	docker compose $(COMPOSE_ENV_FILE) -f infra/compose/docker-compose.yml up -d --build
