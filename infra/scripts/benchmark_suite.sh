#!/bin/sh
# This script runs the repo-owned benchmark suite against a running platform
# stack and writes a timestamped JSON report. It is intended to become the
# backbone of future latency budgets and E2E regression benchmarking.

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
SERVER_URL="${PLATFORM_BENCHMARK_URL:-http://127.0.0.1:8080}"
ADMIN_TOKEN="${PLATFORM_ADMIN_TOKEN:-local-dev-admin-token}"
ITERATIONS="${PLATFORM_BENCHMARK_ITERATIONS:-7}"
PIPELINE_ID="${PLATFORM_BENCHMARK_PIPELINE_ID:-personal_finance_pipeline}"
LOAD_TRIGGERS="${PLATFORM_BENCHMARK_LOAD_TRIGGERS:-4}"
LOAD_CONCURRENCY="${PLATFORM_BENCHMARK_LOAD_CONCURRENCY:-2}"
ANALYTICS_REQUESTS="${PLATFORM_BENCHMARK_ANALYTICS_REQUESTS:-5}"
ANALYTICS_CONCURRENCY="${PLATFORM_BENCHMARK_ANALYTICS_CONCURRENCY:-5}"
TRIGGER_BURST="${PLATFORM_BENCHMARK_TRIGGER_BURST:-3}"
QUEUE_VISIBLE_THRESHOLD_MS="${PLATFORM_BENCHMARK_QUEUE_VISIBLE_THRESHOLD_MS:-5000}"
SCHEDULER_LAG_THRESHOLD_SECONDS="${PLATFORM_BENCHMARK_SCHEDULER_LAG_THRESHOLD_SECONDS:-90}"
POST_RESTORE="${PLATFORM_BENCHMARK_POST_RESTORE:-1}"
OUTPUT_ROOT="${PLATFORM_BENCHMARK_OUTPUT_ROOT:-$ROOT_DIR/var/benchmarks}"
STAMP=$(date -u +"%Y%m%dT%H%M%SZ")
OUTPUT_PATH="${PLATFORM_BENCHMARK_OUTPUT:-$OUTPUT_ROOT/benchmark-${STAMP}.json}"
RESTORED_OUTPUT_PATH="${PLATFORM_BENCHMARK_RESTORED_OUTPUT:-$OUTPUT_ROOT/benchmark-restored-${STAMP}.json}"

mkdir -p "$OUTPUT_ROOT"

if ! curl -fsS "$SERVER_URL/healthz" >/dev/null 2>&1; then
  echo "Benchmark target $SERVER_URL is not healthy. Start the stack first, then rerun." >&2
  exit 1
fi

(
  cd "$ROOT_DIR/backend"
  go run ./cmd/platformctl benchmark \
    --server "$SERVER_URL" \
    --token "$ADMIN_TOKEN" \
    --iterations "$ITERATIONS" \
    --pipeline "$PIPELINE_ID" \
    --load-triggers "$LOAD_TRIGGERS" \
    --load-concurrency "$LOAD_CONCURRENCY" \
    --analytics-requests "$ANALYTICS_REQUESTS" \
    --analytics-concurrency "$ANALYTICS_CONCURRENCY" \
    --trigger-burst "$TRIGGER_BURST" \
    --queue-visible-threshold-ms "$QUEUE_VISIBLE_THRESHOLD_MS" \
    --scheduler-lag-threshold-seconds "$SCHEDULER_LAG_THRESHOLD_SECONDS" \
    --out "$OUTPUT_PATH"
)

if [ "$POST_RESTORE" = "1" ]; then
  RESTORE_OUTPUT=$(
    PLATFORM_ADMIN_TOKEN="$ADMIN_TOKEN" \
    /bin/sh "$ROOT_DIR/infra/scripts/restore_e2e.sh"
  )
  RESTORED_API_URL=$(printf "%s\n" "$RESTORE_OUTPUT" | sed -n 's/^restored_api_url=//p' | tail -n 1)
  if [ -z "$RESTORED_API_URL" ]; then
    printf "%s\n" "$RESTORE_OUTPUT"
    echo "Failed to parse restore E2E output for post-restore benchmark" >&2
    exit 1
  fi

  (
    cd "$ROOT_DIR/backend"
    go run ./cmd/platformctl benchmark \
      --server "$RESTORED_API_URL" \
      --token "$ADMIN_TOKEN" \
      --iterations "$ITERATIONS" \
      --pipeline "$PIPELINE_ID" \
      --load-triggers "$LOAD_TRIGGERS" \
      --load-concurrency "$LOAD_CONCURRENCY" \
      --analytics-requests "$ANALYTICS_REQUESTS" \
      --analytics-concurrency "$ANALYTICS_CONCURRENCY" \
      --trigger-burst "$TRIGGER_BURST" \
      --queue-visible-threshold-ms "$QUEUE_VISIBLE_THRESHOLD_MS" \
      --scheduler-lag-threshold-seconds "$SCHEDULER_LAG_THRESHOLD_SECONDS" \
      --out "$RESTORED_OUTPUT_PATH"
  )
fi

echo "benchmark suite passed"
echo "server_url=$SERVER_URL"
echo "output_path=$OUTPUT_PATH"
if [ "$POST_RESTORE" = "1" ]; then
  echo "restored_output_path=$RESTORED_OUTPUT_PATH"
fi
