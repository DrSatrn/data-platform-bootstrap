# This Makefile centralizes the intended developer workflow for the platform.
# The recipes currently emphasize clear, teachable entrypoints over clever shell
# tricks so the build and runtime model stays easy to understand.

SHELL := /bin/sh

.PHONY: doctor fmt lint test build backend-build web-build up down smoke

doctor:
	@echo "Review codex.md before first build."
	@echo "Verify Go, Node, Docker/OrbStack, and ARM64-compatible container images."

fmt:
	cd backend && gofmt -w $$(find . -name '*.go' -print)

lint:
	cd backend && go test ./...
	cd backend && go run ./cmd/platformctl validate-manifests
	cd web && npm run test

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
	docker compose -f infra/compose/docker-compose.yml up --build

down:
	docker compose -f infra/compose/docker-compose.yml down

smoke:
	cd backend && go run ./cmd/platformctl validate-manifests
