#!/bin/sh
# This script performs a non-destructive restore drill by restoring the latest
# bundle into an isolated temporary runtime root. It exercises the real
# `platformctl backup restore` path without touching the live platform state.

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
KEEP_ROOT="${PLATFORM_RESTORE_DRILL_KEEP:-1}"
BUNDLE_PATH="${PLATFORM_RESTORE_BUNDLE:-}"
RESTORE_ROOT="${PLATFORM_RESTORE_ROOT:-$(mktemp -d /tmp/data-platform-restore-XXXXXX)}"
GOCACHE_DIR="${GOCACHE:-/tmp/data-platform-go-build}"
GOMODCACHE_DIR="${GOMODCACHE:-/tmp/data-platform-go-mod}"

cleanup() {
  set +e
  if [ "$KEEP_ROOT" = "0" ]; then
    rm -rf "$RESTORE_ROOT"
  fi
}

trap cleanup EXIT INT TERM

if [ -z "$BUNDLE_PATH" ]; then
  for candidate in $(find "$ROOT_DIR/var/backups" -maxdepth 1 -name '*.tar.gz' -print | sort -r); do
    if (
      cd "$ROOT_DIR/backend"
      env \
        PLATFORM_MANIFEST_ROOT="$ROOT_DIR/packages/manifests" \
        PLATFORM_DASHBOARD_ROOT="$ROOT_DIR/packages/dashboards" \
        PLATFORM_SQL_ROOT="$ROOT_DIR/packages/sql" \
        PLATFORM_PYTHON_TASK_ROOT="$ROOT_DIR/packages/python" \
        PLATFORM_PYTHON_BINARY=python3 \
        PLATFORM_SAMPLE_DATA_ROOT="$ROOT_DIR/packages/sample_data" \
        PLATFORM_MIGRATIONS_ROOT="$ROOT_DIR/infra/migrations" \
        GOCACHE="$GOCACHE_DIR" \
        GOMODCACHE="$GOMODCACHE_DIR" \
        go run ./cmd/platformctl backup verify --file "$candidate" >/dev/null 2>&1
    ); then
      BUNDLE_PATH="$candidate"
      break
    fi
  done
fi

if [ -z "$BUNDLE_PATH" ] || [ ! -f "$BUNDLE_PATH" ]; then
  echo "No backup bundle found. Set PLATFORM_RESTORE_BUNDLE or run make backup first." >&2
  exit 1
fi

mkdir -p "$RESTORE_ROOT"

(
  cd "$ROOT_DIR/backend"
  env \
    PLATFORM_DATA_ROOT="$RESTORE_ROOT/runtime-data" \
    PLATFORM_ARTIFACT_ROOT="$RESTORE_ROOT/runtime-artifacts" \
    PLATFORM_DUCKDB_PATH="$RESTORE_ROOT/runtime-duckdb/platform.duckdb" \
    PLATFORM_MANIFEST_ROOT="$ROOT_DIR/packages/manifests" \
    PLATFORM_DASHBOARD_ROOT="$ROOT_DIR/packages/dashboards" \
    PLATFORM_SQL_ROOT="$ROOT_DIR/packages/sql" \
    PLATFORM_PYTHON_TASK_ROOT="$ROOT_DIR/packages/python" \
    PLATFORM_PYTHON_BINARY=python3 \
    PLATFORM_SAMPLE_DATA_ROOT="$ROOT_DIR/packages/sample_data" \
    PLATFORM_MIGRATIONS_ROOT="$ROOT_DIR/infra/migrations" \
    GOCACHE="$GOCACHE_DIR" \
    GOMODCACHE="$GOMODCACHE_DIR" \
    go run ./cmd/platformctl backup restore \
      --file "$BUNDLE_PATH" \
      --yes \
      --postgres-mode skip \
      --extract-root "$RESTORE_ROOT/extracted"
)

test -f "$RESTORE_ROOT/extracted/manifest.json"
find "$RESTORE_ROOT/runtime-data/control_plane/runs" -maxdepth 1 -name '*.json' | grep -q '.'
test -d "$RESTORE_ROOT/runtime-data/raw"
test -d "$RESTORE_ROOT/runtime-artifacts"
test -f "$RESTORE_ROOT/runtime-duckdb/platform.duckdb"

echo "restore drill passed"
echo "bundle_path=$BUNDLE_PATH"
echo "restore_root=$RESTORE_ROOT"
echo "restored_data_root=$RESTORE_ROOT/runtime-data"
echo "restored_artifact_root=$RESTORE_ROOT/runtime-artifacts"
