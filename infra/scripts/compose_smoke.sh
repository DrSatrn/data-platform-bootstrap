#!/bin/sh
# This script validates the packaged Docker Compose deployment path instead of
# the host-run binaries. It boots the full stack, waits for migrations and
# health, exercises a real pipeline run, and proves the built web service can
# serve the UI while proxying the API.

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
COMPOSE_FILE="$ROOT_DIR/infra/compose/docker-compose.yml"
API_URL="${PLATFORM_COMPOSE_API_URL:-http://127.0.0.1:8080}"
WEB_URL="${PLATFORM_COMPOSE_WEB_URL:-http://127.0.0.1:3000}"
ADMIN_TOKEN="${PLATFORM_ADMIN_TOKEN:-local-dev-admin-token}"
KEEP_STACK="${PLATFORM_COMPOSE_KEEP:-0}"

cleanup() {
  set +e
  if [ "$KEEP_STACK" = "0" ]; then
    docker compose -f "$COMPOSE_FILE" down -v --remove-orphans >/dev/null 2>&1 || true
  fi
}

trap cleanup EXIT INT TERM

wait_for_url() {
  url="$1"
  attempts=0
  while [ "$attempts" -lt 60 ]; do
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    attempts=$((attempts + 1))
    sleep 2
  done
  echo "Timed out waiting for $url" >&2
  return 1
}

wait_for_run_artifact() {
  run_id="$1"
  expected_artifact="$2"
  attempts=0
  while [ "$attempts" -lt 60 ]; do
    payload=$(curl -fsS "$API_URL/api/v1/artifacts?run_id=$run_id")
    if printf "%s" "$payload" | grep -q "\"${expected_artifact}\""; then
      return 0
    fi
    attempts=$((attempts + 1))
    sleep 2
  done
  echo "Run ${run_id} did not materialize ${expected_artifact}" >&2
  return 1
}

docker compose -f "$COMPOSE_FILE" down -v --remove-orphans >/dev/null 2>&1 || true
docker compose -f "$COMPOSE_FILE" up -d --build

wait_for_url "$API_URL/healthz"
wait_for_url "$WEB_URL/readyz"

manual_response=$(curl -fsS -X POST "$API_URL/api/v1/pipelines" \
  -H "Authorization: Bearer ${ADMIN_TOKEN}" \
  -H 'Content-Type: application/json' \
  -d '{"pipeline_id":"personal_finance_pipeline"}')
manual_run_id=$(printf "%s" "$manual_response" | sed -n 's/.*"run":{"id":"\([^"]*\)".*/\1/p')

if [ -z "$manual_run_id" ]; then
  echo "Failed to parse run ID from compose smoke trigger response" >&2
  exit 1
fi

wait_for_run_artifact "$manual_run_id" "metrics/metrics_category_variance.json"

curl -fsS "$API_URL/api/v1/analytics?dataset=mart_monthly_cashflow" | grep -q '"month"'
curl -fsS "$API_URL/api/v1/analytics?dataset=mart_budget_vs_actual" | grep -q '"variance_amount"'
curl -fsS "$API_URL/api/v1/analytics?metric=metrics_savings_rate" | grep -q '"savings_rate"'
curl -fsS "$API_URL/api/v1/analytics?metric=metrics_category_variance" | grep -q '"variance_amount"'
curl -fsS "$API_URL/api/v1/quality" | grep -q '"checks"'
curl -fsS "$API_URL/api/v1/artifacts?run_id=$manual_run_id" | grep -q '"metrics/metrics_savings_rate.json"'
curl -fsS "$API_URL/api/v1/reports" | grep -q '"finance_overview"'
curl -fsS "$WEB_URL" | grep -q 'Data Platform'

docker compose -f "$COMPOSE_FILE" exec -T api /usr/local/bin/platformctl remote --server http://127.0.0.1:8080 status >/dev/null
docker compose -f "$COMPOSE_FILE" exec -T api /usr/local/bin/platformctl backup create --out /var/lib/platform/data/backups/compose-smoke-backup.tar.gz >/dev/null
docker compose -f "$COMPOSE_FILE" exec -T api /usr/local/bin/platformctl backup verify --file /var/lib/platform/data/backups/compose-smoke-backup.tar.gz >/dev/null
docker compose -f "$COMPOSE_FILE" exec -T api /usr/local/bin/platformctl remote --server http://127.0.0.1:8080 "backup verify compose-smoke-backup.tar.gz" >/dev/null

echo "compose smoke test passed"
echo "api_url=$API_URL"
echo "web_url=$WEB_URL"
echo "manual_run_id=$manual_run_id"
