# Trace One Pipeline

This tutorial walks one real pipeline run from manifest definition to worker
execution to API/UI consumption. The goal is to help a new engineer connect the
platform layers without reading the whole repo first.

## What You Will Learn

- where the pipeline is defined
- how job dependencies are represented
- how the worker materializes each data layer
- how metadata, analytics, and reporting consume the resulting outputs

## Before You Start

Complete one successful smoke run first:

```sh
cd /Users/streanor/Documents/Playground/data-platform
make smoke
```

Expected result:

- `localhost smoke test passed`
- printed `runtime_root=/tmp/...`

## Step 1: Read The Pipeline Manifest

Open:

- [personal_finance_pipeline.yaml](/Users/streanor/Documents/Playground/data-platform/packages/manifests/pipelines/personal_finance_pipeline.yaml)

What to notice:

- the pipeline ID is `personal_finance_pipeline`
- jobs declare `type`, `depends_on`, `inputs`, and `outputs`
- the current slice uses:
  - ingest jobs
  - a Python transform for staging enrichment
  - SQL transforms for intermediate and mart layers
  - quality checks
  - metric publication

## Step 2: Follow Execution Ownership

Open:

- [runner.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/execution/runner.go)

What to notice:

- Go owns the control-plane job loop
- Python is used only for bounded data helpers
- DuckDB-backed SQL handles curated transforms and metrics

## Step 3: Follow The Python Step

Open:

- [enrich_transactions.py](/Users/streanor/Documents/Playground/data-platform/packages/python/tasks/enrich_transactions.py)
- [runtime.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/python/runtime.go)

What to notice:

- the worker sends a JSON task request
- Python writes a structured result
- outputs are mirrored back into run artifacts by Go

Expected produced asset:

- `staging/staging_transactions_enriched.json`

## Step 4: Follow The SQL Layers

Open:

- [intermediate_category_monthly_rollup.sql](/Users/streanor/Documents/Playground/data-platform/packages/sql/transforms/intermediate_category_monthly_rollup.sql)
- [budget_vs_actual.sql](/Users/streanor/Documents/Playground/data-platform/packages/sql/transforms/budget_vs_actual.sql)
- [metrics_category_variance.sql](/Users/streanor/Documents/Playground/data-platform/packages/sql/metrics/metrics_category_variance.sql)

What to notice:

- the slice progresses through `staging`, `intermediate`, `mart`, and `metrics`
- analytics is served from curated outputs, not raw landed tables

## Step 5: Inspect The Materialized Outputs

After a run, inspect:

- `var/data/staging/staging_transactions_enriched.json`
- `var/data/intermediate/intermediate_category_monthly_rollup.json`
- `var/data/mart/mart_budget_vs_actual.json`
- `var/data/metrics/metrics_category_variance.json`

Expected result:

- each file exists
- the data becomes progressively more curated as you move down the layers

## Step 6: Inspect Metadata And Profile Output

Role required: `viewer`

```sh
curl -H "Authorization: Bearer viewer-token" http://127.0.0.1:8080/api/v1/catalog
curl -H "Authorization: Bearer viewer-token" "http://127.0.0.1:8080/api/v1/catalog/profile?asset_id=mart_budget_vs_actual"
```

Expected result:

- catalog response includes freshness, lineage, docs, and quality references
- profile response includes `row_count`, `file_bytes`, and per-column summaries

## Step 7: Inspect Analytics And Reporting

Role required: `viewer`

```sh
curl -H "Authorization: Bearer viewer-token" "http://127.0.0.1:8080/api/v1/analytics?dataset=mart_budget_vs_actual"
curl -H "Authorization: Bearer viewer-token" http://127.0.0.1:8080/api/v1/reports
```

Expected result:

- analytics returns curated rows for the mart
- reports returns the saved dashboard definitions that the UI hydrates

## Step 8: Connect It To The UI

Open the web app and inspect:

- `Datasets` for catalog metadata and runtime profile details
- `Metrics` for semantic metric browsing
- `Dashboard` for saved reporting widgets
- `Pipelines` for run history and artifact browsing

At this point you have traced one full path:

manifest -> queue/run -> worker -> Python/SQL execution -> materialized assets
-> metadata and profiles -> analytics API -> reporting UI
