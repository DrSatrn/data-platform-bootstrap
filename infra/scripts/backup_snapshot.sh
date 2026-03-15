#!/bin/sh
# This script creates and verifies a first-party platform recovery bundle using
# the repo-owned CLI. It is intended as the simplest repeatable operator path
# for local backups outside the broader smoke workflows.

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
TIMESTAMP=$(date -u +%Y%m%dT%H%M%SZ)
OUTPUT_PATH="${PLATFORM_BACKUP_OUT:-$ROOT_DIR/var/backups/platform-backup-$TIMESTAMP.tar.gz}"
GOCACHE_DIR="${GOCACHE:-/tmp/data-platform-go-build}"
GOMODCACHE_DIR="${GOMODCACHE:-/tmp/data-platform-go-mod}"

mkdir -p "$(dirname "$OUTPUT_PATH")"

(
  cd "$ROOT_DIR/backend"
  env \
    GOCACHE="$GOCACHE_DIR" \
    GOMODCACHE="$GOMODCACHE_DIR" \
    go run ./cmd/platformctl backup create --out "$OUTPUT_PATH"
)

(
  cd "$ROOT_DIR/backend"
  env \
    GOCACHE="$GOCACHE_DIR" \
    GOMODCACHE="$GOMODCACHE_DIR" \
    go run ./cmd/platformctl backup verify --file "$OUTPUT_PATH"
)

echo "backup snapshot completed"
echo "bundle_path=$OUTPUT_PATH"
