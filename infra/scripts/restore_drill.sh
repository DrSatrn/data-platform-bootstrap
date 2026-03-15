#!/bin/sh
# This script performs a safe restore drill by extracting a backup bundle into
# an isolated temporary directory and checking that the expected control-plane
# and data paths are present. It does not overwrite the live runtime.

set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname "$0")" && pwd)
ROOT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)
KEEP_ROOT="${PLATFORM_RESTORE_DRILL_KEEP:-1}"
BUNDLE_PATH="${PLATFORM_RESTORE_BUNDLE:-}"
RESTORE_ROOT="${PLATFORM_RESTORE_ROOT:-$(mktemp -d /tmp/data-platform-restore-XXXXXX)}"

cleanup() {
  set +e
  if [ "$KEEP_ROOT" = "0" ]; then
    rm -rf "$RESTORE_ROOT"
  fi
}

trap cleanup EXIT INT TERM

if [ -z "$BUNDLE_PATH" ]; then
  BUNDLE_PATH=$(find "$ROOT_DIR/var/backups" -maxdepth 1 -name '*.tar.gz' -print | sort | tail -n 1 || true)
fi

if [ -z "$BUNDLE_PATH" ] || [ ! -f "$BUNDLE_PATH" ]; then
  echo "No backup bundle found. Set PLATFORM_RESTORE_BUNDLE or run make backup first." >&2
  exit 1
fi

mkdir -p "$RESTORE_ROOT"
tar -xzf "$BUNDLE_PATH" -C "$RESTORE_ROOT"

test -f "$RESTORE_ROOT/manifest.json"
test -f "$RESTORE_ROOT/exports/pipeline_runs.json"
test -f "$RESTORE_ROOT/exports/queue_requests.json"
test -f "$RESTORE_ROOT/exports/dashboards.json"
test -f "$RESTORE_ROOT/exports/audit_events.json"
test -f "$RESTORE_ROOT/exports/data_assets.json"
test -d "$RESTORE_ROOT/files/data"
test -d "$RESTORE_ROOT/files/artifacts"

echo "restore drill passed"
echo "bundle_path=$BUNDLE_PATH"
echo "restore_root=$RESTORE_ROOT"
echo "next_step_copy_data_from=$RESTORE_ROOT/files/data"
echo "next_step_copy_artifacts_from=$RESTORE_ROOT/files/artifacts"
