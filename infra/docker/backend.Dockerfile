# This Dockerfile packages the Go backend for ARM64-friendly local and
# self-hosted deployment. The image contains precompiled binaries plus the
# manifests, SQL, sample data, and migrations they need at runtime so the
# Compose stack behaves like a real service deployment instead of compiling on
# every container boot.
FROM golang:1.24-bookworm AS build

WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN GOOS=linux GOARCH=arm64 go build -o /out/platform-api ./cmd/platform-api && \
    GOOS=linux GOARCH=arm64 go build -o /out/platform-scheduler ./cmd/platform-scheduler && \
    GOOS=linux GOARCH=arm64 go build -o /out/platform-worker ./cmd/platform-worker && \
    GOOS=linux GOARCH=arm64 go build -o /out/platformctl ./cmd/platformctl

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates libstdc++6 && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
COPY --from=build /out/ /usr/local/bin/
COPY packages/manifests /workspace/packages/manifests
COPY packages/sample_data /workspace/packages/sample_data
COPY packages/sql /workspace/packages/sql
COPY infra/migrations /workspace/infra/migrations

CMD ["/usr/local/bin/platform-api"]
