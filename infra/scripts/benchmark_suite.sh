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
OUTPUT_ROOT="${PLATFORM_BENCHMARK_OUTPUT_ROOT:-$ROOT_DIR/var/benchmarks}"
STAMP=$(date -u +"%Y%m%dT%H%M%SZ")
OUTPUT_PATH="${PLATFORM_BENCHMARK_OUTPUT:-$OUTPUT_ROOT/benchmark-${STAMP}.json}"

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
    --out "$OUTPUT_PATH"
)

echo "benchmark suite passed"
echo "server_url=$SERVER_URL"
echo "output_path=$OUTPUT_PATH"
