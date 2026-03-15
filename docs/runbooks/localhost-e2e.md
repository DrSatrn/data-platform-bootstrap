# Localhost End-to-End Runbook

This runbook explains how to bring up the platform locally and verify a real end-to-end pipeline execution on localhost.

## Goal

Prove that the platform can:

- start locally on an Apple Silicon machine
- accept a manual pipeline trigger
- queue the run durably
- execute the run in the worker
- materialize local data artifacts
- persist run snapshots, queue state, and artifact metadata in PostgreSQL when bootstrapped
- expose updated run history and analytics through the API and UI

## Prerequisites

- Go installed locally
- Node and npm installed locally
- host C/C++ build tools available for the DuckDB Go driver
- Docker or OrbStack available for local PostgreSQL if you want the full stack profile
- a local `.env` file created from `.env.example` with placeholder credentials replaced
- at least one configured bearer token for the role you want to test

## Fastest repeatable check

Use the repo-owned smoke script when you want the quickest verified path:

```bash
make smoke
```

That workflow starts API, worker, and scheduler on loopback, triggers a manual
run after confirming the scheduler path, verifies the artifact API, exercises
the admin terminal, creates and verifies a real backup bundle, and proves the
`platformctl remote` CLI.

If the default smoke port is already occupied, choose another loopback port:

```bash
PLATFORM_SMOKE_PORT=18081 make smoke
```

## Packaged deployment check

Use the Compose smoke workflow when you want to validate the production-style
local deployment rather than host-run binaries:

```bash
make compose-smoke
```

That workflow confirms the hosted web UI is reachable, not just the API.

## Benchmark follow-up

After a successful smoke pass, run the benchmark suite to capture response
budgets for the current build:

```bash
make benchmark
```

Capture a recovery point as part of the same validation pass:

```bash
make backup
```

## Recommended local startup

1. Start the API:

```bash
cd backend
PLATFORM_ADMIN_TOKEN=local-dev-admin-token go run ./cmd/platform-api
```

2. Start the worker in a second terminal:

```bash
cd backend
PLATFORM_ADMIN_TOKEN=local-dev-admin-token go run ./cmd/platform-worker
```

3. Start the frontend in a third terminal:

```bash
cd web
npm install
npm run dev
```

4. Optional but recommended: apply migrations if PostgreSQL is running:

```bash
cd backend
go run ./cmd/platformctl migrate
```

For the Compose-backed service image path:

```bash
make bootstrap
```

That path now starts PostgreSQL, runs migrations automatically through the
`migrate` service, waits for API health, and serves the frontend through the
packaged platform web service rather than a Vite dev server.

## Health checks

Confirm the API is reachable:

```bash
curl http://127.0.0.1:8080/healthz
```

Confirm manifest validation still passes:

```bash
cd backend
go run ./cmd/platformctl validate-manifests
```

## Trigger the pipeline

Use the CLI:

```bash
cd backend
PLATFORM_API_BASE_URL=http://127.0.0.1:8080 \
PLATFORM_ADMIN_TOKEN=local-dev-admin-token \
go run ./cmd/platformctl remote trigger personal_finance_pipeline
```

Or use the browser:

- open `http://127.0.0.1:3000`
- paste an `editor` or `admin` token into the sidebar token field
- go to the Pipelines page
- click `Run now`

Or use the System page admin terminal:

- run `trigger personal_finance_pipeline`

## Expected outputs

After the worker poll interval, the run should progress to `succeeded` and these files should exist under the local data root:

- `control_plane/runs/<run_id>.json`
- `raw/raw_transactions.csv`
- `raw/raw_account_balances.json`
- `raw/raw_budget_rules.json`
- `mart/mart_monthly_cashflow.json`
- `mart/mart_category_spend.json`
- `mart/mart_budget_vs_actual.json`
- `quality/check_uncategorized_transactions.json`
- `metrics/metrics_savings_rate.json`
- `metrics/metrics_category_variance.json`

If you are using the default repo-local data root, these will appear under `var/`.

## Verification checks

1. Inspect pipeline status:

```bash
curl http://127.0.0.1:8080/api/v1/pipelines
```

2. Inspect analytics output:

```bash
curl http://127.0.0.1:8080/api/v1/analytics
curl "http://127.0.0.1:8080/api/v1/analytics?dataset=mart_budget_vs_actual"
curl "http://127.0.0.1:8080/api/v1/analytics?metric=metrics_category_variance"
```

3. Inspect system overview:

```bash
curl http://127.0.0.1:8080/api/v1/system/overview
```

4. Inspect recent logs:

```bash
curl http://127.0.0.1:8080/api/v1/system/logs
```

5. Inspect run artifacts for a specific run:

```bash
curl "http://127.0.0.1:8080/api/v1/artifacts?run_id=<run_id>"
```

6. Inspect artifact contents for a specific file:

```bash
curl "http://127.0.0.1:8080/api/v1/artifacts?run_id=<run_id>&path=metrics%2Fmetrics_savings_rate.json"
```

7. Inspect the saved dashboard definitions that drive the reporting UI:

```bash
curl "http://127.0.0.1:8080/api/v1/reports"
```

8. Create, edit, or delete dashboards from the browser:

- open `http://127.0.0.1:3000`
- go to `Dashboard`
- use `New dashboard`, `Duplicate`, `Edit dashboard`, and `Delete`
- save and then confirm the new definition is returned by `/api/v1/reports`

9. Inspect freshness-enriched catalog output:

```bash
curl "http://127.0.0.1:8080/api/v1/catalog"
```

You should now see `freshness_status` on assets, with states such as `fresh`,
`late`, `stale`, or `missing` depending on the local artifact timestamps.

10. Inspect the resolved browser session:

```bash
curl http://127.0.0.1:8080/api/v1/session
curl -H "Authorization: Bearer local-dev-admin-token" http://127.0.0.1:8080/api/v1/session
```

## Failure modes to check first

- API running but no execution:
  The worker is probably not running or not sharing the same data root.
- Run stays queued:
  Confirm `platform-worker` is running and using the same `PLATFORM_DATA_ROOT` as the API.
- UI is stale:
  Refresh the Pipelines or System page after the worker finishes a run.
- Admin CLI cannot connect:
  Confirm `PLATFORM_API_BASE_URL` and `PLATFORM_ADMIN_TOKEN` match the running API configuration.
- Smoke script fails early:
  Inspect the printed `logs_root` path under `/tmp` and review `api.log`,
  `worker.log`, and `scheduler.log` first.
- Compose services take a while to become healthy:
  The first packaged boot may need time to build images before health checks
  succeed.
