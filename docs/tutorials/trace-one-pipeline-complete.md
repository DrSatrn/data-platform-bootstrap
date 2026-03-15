# Trace One Pipeline Complete

This tutorial is an additive replacement draft for the current abbreviated
pipeline walkthrough. It is intended to give a future docs pass a concrete,
end-to-end learner path.

## Goal

Trace the `personal_finance_pipeline` from:

- manifest
- scheduler or manual trigger
- queued run
- worker execution
- materialized artifacts
- analytics API
- reporting UI

## What You Will Learn

- where the pipeline definition lives
- how jobs depend on each other
- what the worker actually materializes
- how curated analytics are served
- why the UI reads constrained APIs instead of arbitrary SQL

## Step 1: Find The Pipeline Manifest

Open:

- `packages/manifests/pipelines/personal_finance_pipeline.yaml`

Look for:

- pipeline id
- schedule
- job order
- job dependencies
- job types such as `ingest`, `transform_sql`, `transform_python`,
  `quality_check`, and `publish_metric`

This file defines the control-plane shape of the workflow before any run is
queued.

## Step 2: Understand The Supporting Definitions

Open these directories:

- `packages/manifests/assets/`
- `packages/manifests/metrics/`
- `packages/manifests/quality/`
- `packages/sql/`
- `packages/python/tasks/`

What they contribute:

- asset manifests define the catalog view
- metric manifests define semantic reporting entities
- quality manifests define operator-visible checks
- SQL files define materialization logic
- Python tasks provide bounded data-runtime helpers

## Step 3: Trigger A Real Run

Use the fastest verified path:

```bash
make smoke
```

Or run the stack yourself and trigger:

```bash
cd backend
PLATFORM_API_BASE_URL=http://127.0.0.1:8080 \
PLATFORM_ADMIN_TOKEN=local-dev-admin-token \
go run ./cmd/platformctl remote trigger personal_finance_pipeline
```

## Step 4: Observe Queued And Running State

Inspect:

```bash
curl http://127.0.0.1:8080/api/v1/pipelines
```

Look for:

- a new run id
- `queued`, then `running`, then `succeeded`
- per-job status under `job_runs`
- recent event messages under `events`

This is the control-plane view of the run.

## Step 5: Inspect What The Worker Wrote

Check the repo-local output roots:

- `var/data/`
- `var/artifacts/runs/<run_id>/`

Typical outputs include:

- raw landed files
- staging enriched files
- intermediate rollups
- mart outputs
- quality check JSON
- metrics JSON

This is the easiest place to see the difference between:

- local materialized state under `var/data/`
- run-scoped snapshots under `var/artifacts/runs/<run_id>/`

## Step 6: Follow The Analytics Layer

Query the API:

```bash
curl "http://127.0.0.1:8080/api/v1/analytics?dataset=mart_budget_vs_actual"
curl "http://127.0.0.1:8080/api/v1/analytics?metric=metrics_category_variance"
```

Notice:

- the API serves curated datasets and metrics
- the reporting UI does not need direct SQL access
- the same curated layer backs both tables and charts

## Step 7: Follow The Metadata Layer

Query:

```bash
curl "http://127.0.0.1:8080/api/v1/catalog"
```

Look for:

- asset ids
- source refs
- quality refs
- documentation refs
- freshness status
- lineage edges

This is how the platform turns repo-managed definitions plus materialized state
into an operator-facing catalog.

## Step 8: See The UI Consume The Same Platform Surface

Open the browser:

- `Dashboard` reads reporting definitions plus analytics results
- `Pipelines` reads pipeline definitions plus run history
- `Datasets` reads catalog and freshness output
- `Metrics` reads semantic metric definitions plus preview data
- `System` reads health, logs, audit, and other trust signals

The important lesson is that the UI is not inventing its own shadow model. It
is consuming platform-owned APIs.

## Step 9: Map Code Ownership

When you want to go deeper, these are the highest-value files to read:

- `backend/internal/orchestration/service.go`
- `backend/internal/orchestration/handler.go`
- `backend/internal/execution/runner.go`
- `backend/internal/transforms/engine.go`
- `backend/internal/metadata/handler.go`
- `backend/internal/analytics/service.go`
- `backend/internal/reporting/handler.go`
- `web/src/pages/PipelinesPage.tsx`
- `web/src/pages/DashboardPage.tsx`
- `web/src/pages/DatasetsPage.tsx`
- `web/src/pages/MetricsPage.tsx`

## What This Tutorial Should Eventually Link To

Before wiring this tutorial into the main docs:

- verify the startup path it references is still the canonical one
- verify role requirements still match the UI and backend
- verify the named pages and APIs still exist
