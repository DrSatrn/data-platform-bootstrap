#!/bin/sh
# This script proves the backup restore path end to end. It creates a real
# bundle from an isolated smoke runtime, restores that bundle into a second
# isolated runtime root, and boots the API against the restored state.

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
SMOKE_PORT="${PLATFORM_RESTORE_E2E_SMOKE_PORT:-18091}"
RESTORE_PORT="${PLATFORM_RESTORE_E2E_PORT:-18092}"
ADMIN_TOKEN="${PLATFORM_ADMIN_TOKEN:-local-dev-admin-token}"
KEEP_ROOT="${PLATFORM_RESTORE_E2E_KEEP:-1}"
RESTORE_ROOT="${PLATFORM_RESTORE_E2E_ROOT:-$(mktemp -d /tmp/data-platform-restore-e2e-XXXXXX)}"
LOG_ROOT="$RESTORE_ROOT/logs"
API_URL="http://127.0.0.1:${RESTORE_PORT}"
GOCACHE_DIR="${GOCACHE:-/tmp/data-platform-go-build}"
GOMODCACHE_DIR="${GOMODCACHE:-/tmp/data-platform-go-mod}"

mkdir -p "$LOG_ROOT"

cleanup() {
  set +e
  [ "${RESTORED_API_PID:-}" ] && kill "$RESTORED_API_PID" >/dev/null 2>&1 || true
  wait "${RESTORED_API_PID:-}" >/dev/null 2>&1 || true
  if [ "$KEEP_ROOT" = "0" ]; then
    rm -rf "$RESTORE_ROOT" "${SMOKE_RUNTIME_ROOT:-}"
  fi
}

trap cleanup EXIT INT TERM

wait_for_api() {
  attempts=0
  while [ "$attempts" -lt 30 ]; do
    if curl -fsS "$API_URL/healthz" >/dev/null 2>&1; then
      return 0
    fi
    attempts=$((attempts + 1))
    sleep 1
  done
  echo "Restored API did not become healthy within 30 seconds" >&2
  return 1
}

SMOKE_OUTPUT=$(
  PLATFORM_SMOKE_KEEP=1 \
  PLATFORM_SMOKE_PORT="$SMOKE_PORT" \
  /bin/sh "$ROOT_DIR/infra/scripts/localhost_smoke.sh"
)

SMOKE_RUNTIME_ROOT=$(printf "%s\n" "$SMOKE_OUTPUT" | sed -n 's/^runtime_root=//p' | tail -n 1)
BACKUP_PATH=$(printf "%s\n" "$SMOKE_OUTPUT" | sed -n 's/^backup_path=//p' | tail -n 1)
MANUAL_RUN_ID=$(printf "%s\n" "$SMOKE_OUTPUT" | sed -n 's/^manual_run_id=//p' | tail -n 1)

if [ -z "$SMOKE_RUNTIME_ROOT" ] || [ -z "$BACKUP_PATH" ] || [ -z "$MANUAL_RUN_ID" ]; then
  printf "%s\n" "$SMOKE_OUTPUT"
  echo "Failed to parse localhost smoke output for restore E2E" >&2
  exit 1
fi

(
  cd "$ROOT_DIR/backend"
  env \
    PLATFORM_DATA_ROOT="$RESTORE_ROOT/data" \
    PLATFORM_ARTIFACT_ROOT="$RESTORE_ROOT/artifacts" \
    PLATFORM_DUCKDB_PATH="$RESTORE_ROOT/duckdb/platform.duckdb" \
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
      --file "$BACKUP_PATH" \
      --yes \
      --postgres-mode skip \
      --extract-root "$RESTORE_ROOT/extracted" >"$LOG_ROOT/restore.log"
)

(
  cd "$ROOT_DIR/backend"
  env \
    PLATFORM_ENV=development \
    PLATFORM_HTTP_ADDR="127.0.0.1:${RESTORE_PORT}" \
    PLATFORM_API_BASE_URL="$API_URL" \
    PLATFORM_LOG_LEVEL=debug \
    PLATFORM_DATA_ROOT="$RESTORE_ROOT/data" \
    PLATFORM_ARTIFACT_ROOT="$RESTORE_ROOT/artifacts" \
    PLATFORM_DUCKDB_PATH="$RESTORE_ROOT/duckdb/platform.duckdb" \
    PLATFORM_MANIFEST_ROOT="$ROOT_DIR/packages/manifests" \
    PLATFORM_DASHBOARD_ROOT="$ROOT_DIR/packages/dashboards" \
    PLATFORM_SQL_ROOT="$ROOT_DIR/packages/sql" \
    PLATFORM_PYTHON_TASK_ROOT="$ROOT_DIR/packages/python" \
    PLATFORM_PYTHON_BINARY=python3 \
    PLATFORM_SAMPLE_DATA_ROOT="$ROOT_DIR/packages/sample_data" \
    PLATFORM_MIGRATIONS_ROOT="$ROOT_DIR/infra/migrations" \
    PLATFORM_ADMIN_TOKEN="$ADMIN_TOKEN" \
    GOCACHE="$GOCACHE_DIR" \
    GOMODCACHE="$GOMODCACHE_DIR" \
    go run ./cmd/platform-api >"$LOG_ROOT/restored-api.log" 2>&1
) &
RESTORED_API_PID=$!

wait_for_api

curl -fsS -H "Authorization: Bearer ${ADMIN_TOKEN}" "$API_URL/api/v1/reports" | grep -q '"finance_overview"'
curl -fsS -H "Authorization: Bearer ${ADMIN_TOKEN}" "$API_URL/api/v1/catalog" | grep -q '"mart_budget_vs_actual"'
curl -fsS -H "Authorization: Bearer ${ADMIN_TOKEN}" "$API_URL/api/v1/catalog/profile?asset_id=mart_budget_vs_actual" | grep -q '"row_count"'
curl -fsS -H "Authorization: Bearer ${ADMIN_TOKEN}" "$API_URL/api/v1/analytics?dataset=mart_budget_vs_actual" | grep -q '"variance_amount"'
curl -fsS -H "Authorization: Bearer ${ADMIN_TOKEN}" "$API_URL/api/v1/artifacts?run_id=$MANUAL_RUN_ID" | grep -q '"metrics/metrics_category_variance.json"'

echo "restore e2e passed"
echo "backup_path=$BACKUP_PATH"
echo "smoke_runtime_root=$SMOKE_RUNTIME_ROOT"
echo "restore_root=$RESTORE_ROOT"
echo "restored_api_url=$API_URL"
echo "manual_run_id=$MANUAL_RUN_ID"
