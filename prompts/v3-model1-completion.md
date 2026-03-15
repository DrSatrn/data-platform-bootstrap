# v3 Model 1 Completion Report

## Files Changed

- `backend/go.mod`
- `backend/go.sum`
- `backend/cmd/platformctl/main.go`
- `backend/internal/config/config.go`
- `backend/internal/execution/runner.go`
- `backend/internal/execution/runner_manifest_jobs_test.go`
- `backend/internal/execution/runner_retry_ingest_test.go`
- `backend/internal/execution/external_tool_test.go`
- `backend/internal/execution/external_tool_operator_inspection_test.go`
- `backend/internal/execution/external_tool_failures_test.go`
- `backend/internal/externaltools/types.go`
- `backend/internal/externaltools/runner.go`
- `backend/internal/python/runtime.go`
- `backend/internal/orchestration/models.go`
- `backend/internal/orchestration/validation_jobs.go`
- `backend/internal/orchestration/validation_jobs_test.go`
- `backend/internal/metadata/models.go`
- `backend/internal/ingestion/README.md`
- `backend/internal/ingestion/exporter.go`
- `backend/internal/retention/README.md`
- `backend/internal/retention/service.go`
- `backend/internal/retention/service_test.go`
- `backend/test/retention_purge_smoke.sh`
- `packages/manifests/assets/raw_transactions.yaml`
- `packages/manifests/assets/raw_account_balances.yaml`
- `packages/manifests/assets/raw_budget_rules.yaml`
- `packages/manifests/assets/raw_stock_movements.yaml`
- `packages/manifests/assets/staging_transactions_enriched.yaml`
- `packages/manifests/assets/intermediate_category_monthly_rollup.yaml`
- `packages/manifests/assets/mart_monthly_cashflow.yaml`
- `packages/manifests/assets/mart_category_spend.yaml`
- `packages/manifests/assets/mart_budget_vs_actual.yaml`
- `packages/manifests/assets/mart_inventory_monthly_summary.yaml`

## What Is Now Verifiably True

- The worker now retries failed jobs with deterministic exponential backoff.
- Each job attempt now gets a stable per-attempt idempotency key of the form `run_id:job_id:attempt`.
- Python and external-tool subprocesses now receive that idempotency key.
- Ingest jobs now support native PostgreSQL and MySQL query export into raw-layer CSV files through the existing manifest contract.
- Manifest validation now understands the new database-ingest shape and retention policy fields.
- Asset manifests can now declare retention windows for materializations, run artifacts, and run history.
- `platformctl retention purge` now removes stale materializations, run snapshot files, run artifact directories, and mirrored PostgreSQL run rows when a database executor is available.
- The dedicated purge proof script succeeds and proves filesystem cleanup from manifest-declared retention windows.

## Verification Commands And Results

- `cd backend && go test -count=1 ./internal/execution/... ./internal/orchestration/... ./internal/retention/... ./cmd/platformctl` — PASS
- `cd backend && go test -count=1 ./...` — PASS
- `cd backend && go run ./cmd/platformctl validate-manifests` — PASS
- `sh backend/test/retention_purge_smoke.sh` — PASS
- `git diff --check` — PASS

## What Was Explicitly NOT Changed

- `backend/internal/app/runtime.go`
- `backend/internal/authz/`
- `backend/internal/db/` schemas or migrations
- frontend files under `web/`
- reviewer-controlled deployment chokepoints such as `infra/compose/docker-compose.yml` and `Makefile`

## Escalation Items

- `guide-wire.md` still conflicts with the newer v3 prompt set in a few ownership details; I treated the v3 prompt files as the active contract.
- The new database-ingest path exports CSV only. That satisfies the current completion signal, but parquet export remains future work if the platform wants database ingest directly into columnar files.
