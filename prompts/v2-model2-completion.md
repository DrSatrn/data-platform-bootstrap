# Model 2 Completion — Operational Resilience & Domain Expansion

## 1. Files changed

- `backend/internal/orchestration/models.go`
- `backend/internal/orchestration/validation_jobs.go`
- `backend/internal/orchestration/validation_external_tool.go`
- `backend/internal/orchestration/validation_jobs_test.go`
- `backend/internal/execution/runner.go`
- `backend/internal/execution/runner_manifest_jobs_test.go`
- `backend/internal/analytics/service.go`
- `backend/internal/analytics/service_test.go`
- `backend/test/domain_expansion_test.go`
- `backend/test/resilience_test.go`
- `backend/cmd/platformctl/main.go`
- `backend/cmd/platformctl/main_test.go`
- `packages/manifests/pipelines/personal_finance_pipeline.yaml`
- `packages/manifests/pipelines/inventory_operations_pipeline.yaml`
- `packages/manifests/assets/raw_stock_movements.yaml`
- `packages/manifests/assets/mart_inventory_monthly_summary.yaml`
- `packages/manifests/metrics/inventory_net_change.yaml`
- `packages/manifests/quality/check_negative_inventory_positions.yaml`
- `packages/sample_data/inventory_operations/stock_movements.csv`
- `packages/sql/bootstrap/raw_stock_movements.sql`
- `packages/sql/transforms/inventory_monthly_summary.sql`
- `packages/sql/metrics/metrics_inventory_net_change.sql`
- `packages/sql/quality/check_negative_inventory_positions.sql`
- `infra/scripts/benchmark_suite.sh`
- `infra/scripts/resilience_drill.sh`
- `docs/runbooks/benchmarking.md`
- `docs/runbooks/operator-manual.md`
- `docs/runbooks/resilience-drills.md`
- `docs/runbooks/upgrading.md`

## 2. What is now verifiably true

- Ingestion is no longer hardcoded by job ID. Ingest jobs now declare `source_ref`, `target_path`, and optional `artifact_path` in the manifest.
- SQL transforms can declare manifest-driven DuckDB bootstrap steps instead of relying on the old finance-only bootstrap path.
- Metric publication can follow declared `metric_refs` instead of the old finance-only hardcoded metric list.
- The existing personal-finance pipeline still executes through the generalized path.
- A second runnable domain now exists: `inventory_operations_pipeline`.
- The analytics service now serves a second curated dataset and metric:
  - `mart_inventory_monthly_summary`
  - `metrics_inventory_net_change`
- Both domain pipelines execute and return analytics data in the backend integration test suite.
- Recovery/resilience proof now exists for:
  - worker restart reclaiming active queue requests
  - corrupt DuckDB returning a clear error instead of panicking
  - restore E2E after a real run
- The benchmark suite now includes:
  - concurrent analytics requests
  - back-to-back trigger bursts
  - an optional post-restore rerun
- Live benchmark verification passed against a temporary loopback stack and against the post-restore stack.

## 3. Verification commands and results

- `cd backend && go test ./internal/orchestration ./internal/execution ./internal/analytics ./internal/manifests`
  - passed
- `cd backend && go test ./cmd/platformctl ./internal/analytics ./test`
  - passed
- `cd backend && go run ./cmd/platformctl validate-manifests`
  - passed
- `sh infra/scripts/resilience_drill.sh`
  - passed after rerunning outside the sandbox so `go run` could fetch missing Go modules
  - output included:
    - `restore e2e passed`
    - `resilience drill passed`
- `env PLATFORM_BENCHMARK_URL=http://127.0.0.1:18120 PLATFORM_ADMIN_TOKEN=local-dev-admin-token PLATFORM_BENCHMARK_POST_RESTORE=1 sh infra/scripts/benchmark_suite.sh`
  - passed
  - wrote:
    - `var/benchmarks/benchmark-20260315T092236Z.json`
    - `var/benchmarks/benchmark-restored-20260315T092236Z.json`
  - observed on the live stack:
    - concurrent analytics: `5/5` successes
    - trigger burst: `3/3` accepted
    - queue visibility: `6.75ms`
    - scheduler lag: `0.59s`
  - observed on the restored stack:
    - concurrent analytics: `5/5` successes
    - trigger burst: `3/3` accepted
    - queue visibility: `5.62ms`
    - scheduler lag: `2.69s`
- `git diff --check`
  - passed

## 4. What was explicitly NOT changed

- `backend/internal/app/runtime.go`
- `backend/internal/authz/`
- frontend wiring or routed UI behavior
- canonical product-positioning docs outside the runbooks touched above
- new job types beyond the existing orchestration model

## 5. Escalation items

- No code-level escalation was required.
- One environment-level issue occurred during verification: the first sandboxed `resilience_drill.sh` run failed because `go run` could not reach `proxy.golang.org`. Rerunning the drill with escalated permissions resolved that, and the drill passed end to end.
