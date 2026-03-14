# This Dockerfile packages the Go backend for ARM64-friendly local and
# self-hosted deployment. It intentionally keeps the image simple so it remains
# easy to inspect and adjust while the platform is still evolving quickly.
FROM golang:1.23-alpine AS build

WORKDIR /workspace/backend
COPY backend/go.mod ./
RUN go mod download

COPY backend/ ./
RUN GOOS=linux GOARCH=arm64 go build -o /out/platform-api ./cmd/platform-api && \
    GOOS=linux GOARCH=arm64 go build -o /out/platform-scheduler ./cmd/platform-scheduler && \
    GOOS=linux GOARCH=arm64 go build -o /out/platform-worker ./cmd/platform-worker

FROM alpine:3.21

WORKDIR /app
COPY --from=build /out /app/bin
CMD ["/app/bin/platform-api"]
