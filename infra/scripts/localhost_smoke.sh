#!/bin/sh
# This script performs a full localhost smoke test for the platform using
# loopback-only ports and an isolated temporary data root. It is intended to be
# the fastest reproducible proof that API, scheduler, worker, admin tooling,
# artifact APIs, and the CLI can all cooperate end to end on a developer
# machine without mutating the normal repo-local `var/` state.

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)

PORT="${PLATFORM_SMOKE_PORT:-18080}"
ADMIN_TOKEN="${PLATFORM_ADMIN_TOKEN:-local-dev-admin-token}"
KEEP_ROOT="${PLATFORM_SMOKE_KEEP:-1}"
RUNTIME_ROOT="${PLATFORM_SMOKE_ROOT:-$(mktemp -d /tmp/data-platform-smoke-XXXXXX)}"
DATA_ROOT="$RUNTIME_ROOT/data"
ARTIFACT_ROOT="$RUNTIME_ROOT/artifacts"
LOG_ROOT="$RUNTIME_ROOT/logs"
API_URL="http://127.0.0.1:${PORT}"
GOCACHE_DIR="${GOCACHE:-/tmp/data-platform-go-build}"
GOMODCACHE_DIR="${GOMODCACHE:-/tmp/data-platform-go-mod}"

mkdir -p "$LOG_ROOT"

cleanup() {
  set +e
  [ "${API_PID:-}" ] && kill "$API_PID" >/dev/null 2>&1 || true
  [ "${WORKER_PID:-}" ] && kill "$WORKER_PID" >/dev/null 2>&1 || true
  [ "${SCHEDULER_PID:-}" ] && kill "$SCHEDULER_PID" >/dev/null 2>&1 || true
  wait "${API_PID:-}" "${WORKER_PID:-}" "${SCHEDULER_PID:-}" >/dev/null 2>&1 || true
  if [ "$KEEP_ROOT" = "0" ]; then
    rm -rf "$RUNTIME_ROOT"
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
  echo "API did not become healthy within 30 seconds" >&2
  return 1
}

wait_for_run_artifact() {
  run_id="$1"
  expected_artifact="$2"
  attempts=0
  while [ "$attempts" -lt 30 ]; do
    payload=$(curl -fsS "$API_URL/api/v1/artifacts?run_id=$run_id")
    if printf "%s" "$payload" | grep -q "\"${expected_artifact}\""; then
      return 0
    fi
    attempts=$((attempts + 1))
    sleep 1
  done
  echo "Run ${run_id} did not materialize ${expected_artifact} within 30 seconds" >&2
  return 1
}

wait_for_scheduled_run() {
  attempts=0
  while [ "$attempts" -lt 30 ]; do
    payload=$(curl -fsS "$API_URL/api/v1/pipelines")
    if printf "%s" "$payload" | tr -d '\n' | grep -q '"trigger":"scheduled"'; then
      return 0
    fi
    attempts=$((attempts + 1))
    sleep 1
  done
  echo "Scheduler did not enqueue a scheduled run within 30 seconds" >&2
  return 1
}

if curl -fsS "$API_URL/healthz" >/dev/null 2>&1; then
  echo "Smoke port ${PORT} is already serving a health endpoint. Set PLATFORM_SMOKE_PORT to an unused loopback port and rerun." >&2
  exit 1
fi

(
  cd "$ROOT_DIR/backend"
  env \
    PLATFORM_ENV=development \
    PLATFORM_HTTP_ADDR="127.0.0.1:${PORT}" \
    PLATFORM_API_BASE_URL="$API_URL" \
    PLATFORM_LOG_LEVEL=debug \
    PLATFORM_DATA_ROOT="$DATA_ROOT" \
      PLATFORM_ARTIFACT_ROOT="$ARTIFACT_ROOT" \
    PLATFORM_MANIFEST_ROOT="$ROOT_DIR/packages/manifests" \
      PLATFORM_DASHBOARD_ROOT="$ROOT_DIR/packages/dashboards" \
      PLATFORM_SQL_ROOT="$ROOT_DIR/packages/sql" \
      PLATFORM_PYTHON_TASK_ROOT="$ROOT_DIR/packages/python" \
      PLATFORM_PYTHON_BINARY=python3 \
      PLATFORM_SAMPLE_DATA_ROOT="$ROOT_DIR/packages/sample_data" \
    PLATFORM_MIGRATIONS_ROOT="$ROOT_DIR/infra/migrations" \
    PLATFORM_ADMIN_TOKEN="$ADMIN_TOKEN" \
    PLATFORM_SCHEDULER_TICK=1s \
    PLATFORM_WORKER_POLL=1s \
    GOCACHE="$GOCACHE_DIR" \
    GOMODCACHE="$GOMODCACHE_DIR" \
    go run ./cmd/platform-api >"$LOG_ROOT/api.log" 2>&1
) &
API_PID=$!

(
  cd "$ROOT_DIR/backend"
  env \
    PLATFORM_ENV=development \
    PLATFORM_HTTP_ADDR="127.0.0.1:${PORT}" \
    PLATFORM_API_BASE_URL="$API_URL" \
    PLATFORM_LOG_LEVEL=debug \
    PLATFORM_DATA_ROOT="$DATA_ROOT" \
      PLATFORM_ARTIFACT_ROOT="$ARTIFACT_ROOT" \
    PLATFORM_MANIFEST_ROOT="$ROOT_DIR/packages/manifests" \
      PLATFORM_DASHBOARD_ROOT="$ROOT_DIR/packages/dashboards" \
      PLATFORM_SQL_ROOT="$ROOT_DIR/packages/sql" \
      PLATFORM_PYTHON_TASK_ROOT="$ROOT_DIR/packages/python" \
      PLATFORM_PYTHON_BINARY=python3 \
      PLATFORM_SAMPLE_DATA_ROOT="$ROOT_DIR/packages/sample_data" \
    PLATFORM_MIGRATIONS_ROOT="$ROOT_DIR/infra/migrations" \
    PLATFORM_ADMIN_TOKEN="$ADMIN_TOKEN" \
    PLATFORM_SCHEDULER_TICK=1s \
    PLATFORM_WORKER_POLL=1s \
    GOCACHE="$GOCACHE_DIR" \
    GOMODCACHE="$GOMODCACHE_DIR" \
    go run ./cmd/platform-worker >"$LOG_ROOT/worker.log" 2>&1
) &
WORKER_PID=$!

(
  cd "$ROOT_DIR/backend"
  env \
    PLATFORM_ENV=development \
    PLATFORM_HTTP_ADDR="127.0.0.1:${PORT}" \
    PLATFORM_API_BASE_URL="$API_URL" \
    PLATFORM_LOG_LEVEL=debug \
    PLATFORM_DATA_ROOT="$DATA_ROOT" \
      PLATFORM_ARTIFACT_ROOT="$ARTIFACT_ROOT" \
    PLATFORM_MANIFEST_ROOT="$ROOT_DIR/packages/manifests" \
      PLATFORM_DASHBOARD_ROOT="$ROOT_DIR/packages/dashboards" \
      PLATFORM_SQL_ROOT="$ROOT_DIR/packages/sql" \
      PLATFORM_PYTHON_TASK_ROOT="$ROOT_DIR/packages/python" \
      PLATFORM_PYTHON_BINARY=python3 \
      PLATFORM_SAMPLE_DATA_ROOT="$ROOT_DIR/packages/sample_data" \
    PLATFORM_MIGRATIONS_ROOT="$ROOT_DIR/infra/migrations" \
    PLATFORM_ADMIN_TOKEN="$ADMIN_TOKEN" \
    PLATFORM_SCHEDULER_TICK=1s \
    PLATFORM_WORKER_POLL=1s \
    GOCACHE="$GOCACHE_DIR" \
    GOMODCACHE="$GOMODCACHE_DIR" \
    go run ./cmd/platform-scheduler >"$LOG_ROOT/scheduler.log" 2>&1
) &
SCHEDULER_PID=$!

wait_for_api
wait_for_scheduled_run

manual_response=$(curl -fsS -X POST "$API_URL/api/v1/pipelines" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{"pipeline_id":"personal_finance_pipeline"}')
manual_run_id=$(printf "%s" "$manual_response" | sed -n 's/.*"run":{"id":"\([^"]*\)".*/\1/p')

if [ -z "$manual_run_id" ]; then
  echo "Failed to parse manual run id from trigger response" >&2
  exit 1
fi

wait_for_run_artifact "$manual_run_id" "metrics/metrics_category_variance.json"
wait_for_run_artifact "$manual_run_id" "staging/staging_transactions_enriched.json"
wait_for_run_artifact "$manual_run_id" "intermediate/intermediate_category_monthly_rollup.json"

artifacts_payload=$(curl -fsS "$API_URL/api/v1/artifacts?run_id=$manual_run_id")
printf "%s" "$artifacts_payload" | grep -q '"metrics/metrics_savings_rate.json"'
printf "%s" "$artifacts_payload" | grep -q '"metrics/metrics_category_variance.json"'
printf "%s" "$artifacts_payload" | grep -q '"staging/staging_transactions_enriched.json"'
printf "%s" "$artifacts_payload" | grep -q '"intermediate/intermediate_category_monthly_rollup.json"'

reports_payload=$(curl -fsS "$API_URL/api/v1/reports")
printf "%s" "$reports_payload" | grep -q '"finance_overview"'

budget_payload=$(curl -fsS "$API_URL/api/v1/analytics?dataset=mart_budget_vs_actual")
printf "%s" "$budget_payload" | grep -q '"variance_amount"'

metrics_payload=$(curl -fsS "$API_URL/api/v1/metrics")
printf "%s" "$metrics_payload" | grep -q '"metrics_savings_rate"'

BACKUP_PATH="$DATA_ROOT/backups/localhost-smoke-backup.tar.gz"

(
  cd "$ROOT_DIR/backend"
  env \
    PLATFORM_ENV=development \
    PLATFORM_HTTP_ADDR="127.0.0.1:${PORT}" \
    PLATFORM_API_BASE_URL="$API_URL" \
    PLATFORM_LOG_LEVEL=debug \
    PLATFORM_DATA_ROOT="$DATA_ROOT" \
    PLATFORM_ARTIFACT_ROOT="$ARTIFACT_ROOT" \
    PLATFORM_MANIFEST_ROOT="$ROOT_DIR/packages/manifests" \
    PLATFORM_DASHBOARD_ROOT="$ROOT_DIR/packages/dashboards" \
    PLATFORM_SQL_ROOT="$ROOT_DIR/packages/sql" \
    PLATFORM_PYTHON_TASK_ROOT="$ROOT_DIR/packages/python" \
    PLATFORM_PYTHON_BINARY=python3 \
    PLATFORM_SAMPLE_DATA_ROOT="$ROOT_DIR/packages/sample_data" \
    PLATFORM_MIGRATIONS_ROOT="$ROOT_DIR/infra/migrations" \
    PLATFORM_ADMIN_TOKEN="$ADMIN_TOKEN" \
    PLATFORM_SCHEDULER_TICK=1s \
    PLATFORM_WORKER_POLL=1s \
    GOCACHE="$GOCACHE_DIR" \
    GOMODCACHE="$GOMODCACHE_DIR" \
    go run ./cmd/platformctl backup create --out "$BACKUP_PATH" >/dev/null
)

(
  cd "$ROOT_DIR/backend"
  env \
    PLATFORM_ENV=development \
    PLATFORM_HTTP_ADDR="127.0.0.1:${PORT}" \
    PLATFORM_API_BASE_URL="$API_URL" \
    PLATFORM_LOG_LEVEL=debug \
    PLATFORM_DATA_ROOT="$DATA_ROOT" \
    PLATFORM_ARTIFACT_ROOT="$ARTIFACT_ROOT" \
    PLATFORM_MANIFEST_ROOT="$ROOT_DIR/packages/manifests" \
    PLATFORM_DASHBOARD_ROOT="$ROOT_DIR/packages/dashboards" \
    PLATFORM_SQL_ROOT="$ROOT_DIR/packages/sql" \
    PLATFORM_PYTHON_TASK_ROOT="$ROOT_DIR/packages/python" \
    PLATFORM_PYTHON_BINARY=python3 \
    PLATFORM_SAMPLE_DATA_ROOT="$ROOT_DIR/packages/sample_data" \
    PLATFORM_MIGRATIONS_ROOT="$ROOT_DIR/infra/migrations" \
    PLATFORM_ADMIN_TOKEN="$ADMIN_TOKEN" \
    PLATFORM_SCHEDULER_TICK=1s \
    PLATFORM_WORKER_POLL=1s \
    GOCACHE="$GOCACHE_DIR" \
    GOMODCACHE="$GOMODCACHE_DIR" \
    go run ./cmd/platformctl backup verify --file "$BACKUP_PATH" >/dev/null
)

test -f "$BACKUP_PATH"

admin_payload=$(curl -fsS -X POST "$API_URL/api/v1/admin/terminal/execute" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -d "{\"command\":\"artifacts ${manual_run_id}\"}")
printf "%s" "$admin_payload" | grep -q '"success":true'

backup_payload=$(curl -fsS -X POST "$API_URL/api/v1/admin/terminal/execute" \
  -H 'Content-Type: application/json' \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -d '{"command":"backups"}')
printf "%s" "$backup_payload" | grep -q 'localhost-smoke-backup.tar.gz'

(
  cd "$ROOT_DIR/backend"
  env \
    PLATFORM_API_BASE_URL="$API_URL" \
    PLATFORM_ADMIN_TOKEN="$ADMIN_TOKEN" \
    PLATFORM_MANIFEST_ROOT="$ROOT_DIR/packages/manifests" \
    PLATFORM_DASHBOARD_ROOT="$ROOT_DIR/packages/dashboards" \
    PLATFORM_SQL_ROOT="$ROOT_DIR/packages/sql" \
    PLATFORM_PYTHON_TASK_ROOT="$ROOT_DIR/packages/python" \
    PLATFORM_PYTHON_BINARY=python3 \
    PLATFORM_SAMPLE_DATA_ROOT="$ROOT_DIR/packages/sample_data" \
    PLATFORM_MIGRATIONS_ROOT="$ROOT_DIR/infra/migrations" \
    GOCACHE="$GOCACHE_DIR" \
    GOMODCACHE="$GOMODCACHE_DIR" \
    go run ./cmd/platformctl remote "backup verify localhost-smoke-backup.tar.gz" >/dev/null
)

echo "localhost smoke test passed"
echo "api_url=$API_URL"
echo "manual_run_id=$manual_run_id"
echo "backup_path=$BACKUP_PATH"
echo "runtime_root=$RUNTIME_ROOT"
echo "logs_root=$LOG_ROOT"
