#!/bin/sh

set -eu

REPO_ROOT="$(CDPATH= cd -- "$(dirname "$0")/.." && pwd)"
TMP_ROOT="$(mktemp -d "${TMPDIR:-/tmp}/platform-retention-XXXXXX")"
trap 'rm -rf "$TMP_ROOT"' EXIT

MANIFEST_ROOT="$TMP_ROOT/manifests"
DATA_ROOT="$TMP_ROOT/data"
ARTIFACT_ROOT="$TMP_ROOT/artifacts"
RUN_ROOT="$DATA_ROOT/control_plane/runs"

mkdir -p "$MANIFEST_ROOT/assets" "$MANIFEST_ROOT/pipelines" "$DATA_ROOT/raw" "$RUN_ROOT" "$ARTIFACT_ROOT/runs/run_old"

cat >"$MANIFEST_ROOT/assets/raw_transactions.yaml" <<'YAML'
id: raw_transactions
name: Raw Transactions
layer: raw
description: Test raw transactions asset.
owner: platform-team
kind: table
source_refs:
  - sample.transactions_csv
freshness:
  expected_within: 6h
  warn_after: 8h
retention:
  materializations: 48h
  run_artifacts: 168h
  run_history: 168h
columns:
  - name: transaction_id
    type: text
    description: Stable identifier.
    is_pii: false
YAML

cat >"$MANIFEST_ROOT/pipelines/retention_pipeline.yaml" <<'YAML'
id: retention_pipeline
name: Retention Pipeline
description: Minimal pipeline for retention smoke verification.
owner: platform-team
tags: [smoke]
schedule:
  cron: "0 * * * *"
  timezone: UTC
  catchup: false
jobs:
  - id: ingest_transactions
    name: Ingest Transactions
    type: ingest
    ingest:
      source_ref: personal_finance/transactions.csv
      target_path: raw/raw_transactions.csv
      artifact_path: raw/raw_transactions.csv
YAML

cat >"$RUN_ROOT/run_old.json" <<'JSON'
{
  "id": "run_old",
  "pipeline_id": "retention_pipeline",
  "status": "succeeded",
  "trigger": "manual",
  "started_at": "2026-03-01T00:00:00Z",
  "updated_at": "2026-03-01T00:00:00Z",
  "finished_at": "2026-03-01T00:10:00Z",
  "job_runs": [],
  "events": []
}
JSON

printf 'transaction_id,amount\n1,20\n' >"$DATA_ROOT/raw/raw_transactions.csv"
printf '{}' >"$ARTIFACT_ROOT/runs/run_old/artifact.json"
touch -t 202603010000 "$DATA_ROOT/raw/raw_transactions.csv"

cd "$REPO_ROOT"
go run ./cmd/platformctl retention purge \
  --manifest-root "$MANIFEST_ROOT" \
  --data-root "$DATA_ROOT" \
  --artifact-root "$ARTIFACT_ROOT" \
  --skip-postgres \
  --now "2026-03-15T12:00:00Z" >/dev/null

[ ! -f "$DATA_ROOT/raw/raw_transactions.csv" ]
[ ! -f "$RUN_ROOT/run_old.json" ]
[ ! -d "$ARTIFACT_ROOT/runs/run_old" ]

echo "retention purge smoke passed"
