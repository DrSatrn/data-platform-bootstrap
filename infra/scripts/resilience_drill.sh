#!/bin/sh
# This script runs the repo-owned resilience drills so operators can verify the
# platform still handles restart and recovery scenarios cleanly.

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)

echo "running resilience drill: queue reclaim + duckdb corruption tests"
(
  cd "$ROOT_DIR/backend"
  go test ./test -run 'TestWorkerRestartReclaimsActiveQueueRequest|TestCorruptDuckDBReturnsClearError'
)

echo "running resilience drill: restore e2e"
/bin/sh "$ROOT_DIR/infra/scripts/restore_e2e.sh"

echo "resilience drill passed"
