# Localhost End-to-End Runbook

This runbook explains how to bring up the platform locally and verify a real end-to-end pipeline execution on localhost.

## Goal

Prove that the platform can:

- start locally on an Apple Silicon machine
- accept a manual pipeline trigger
- queue the run durably
- execute the run in the worker
- materialize local data artifacts
- expose updated run history and analytics through the API and UI

## Prerequisites

- Go installed locally
- Node and npm installed locally
- Docker or OrbStack available for local PostgreSQL if you want the full stack profile
- a local `.env` file created from `.env.example` with placeholder credentials replaced

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
- go to the Pipelines page
- click `Run now`

Or use the System page admin terminal:

- run `trigger personal_finance_pipeline`

## Expected outputs

After the worker poll interval, the run should progress to `succeeded` and these files should exist under the local data root:

- `control_plane/runs/<run_id>.json`
- `raw/raw_transactions.csv`
- `raw/raw_account_balances.json`
- `mart/mart_monthly_cashflow.json`
- `quality/check_uncategorized_transactions.json`
- `metrics/metrics_savings_rate.json`

If you are using the default repo-local data root, these will appear under `var/`.

## Verification checks

1. Inspect pipeline status:

```bash
curl http://127.0.0.1:8080/api/v1/pipelines
```

2. Inspect analytics output:

```bash
curl http://127.0.0.1:8080/api/v1/analytics
```

3. Inspect system overview:

```bash
curl http://127.0.0.1:8080/api/v1/system/overview
```

4. Inspect recent logs:

```bash
curl http://127.0.0.1:8080/api/v1/system/logs
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
