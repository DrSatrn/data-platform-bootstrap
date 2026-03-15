# Trace One Pipeline Complete

This is the most detailed end-to-end walkthrough in the repo. Use it when you
want to understand not just how to start the platform, but how one real data
pipeline moves through manifests, queueing, execution, artifacts, APIs, and
the UI.

Recommended prerequisite:

- complete [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md)
  first so you already know the platform can start successfully

## Goal

Trace `personal_finance_pipeline` from:

1. manifest definition
2. manual trigger
3. queued and running state
4. worker execution
5. materialized outputs
6. dataset catalog and profile output
7. analytics and reporting consumption

## Step 1: Find The Pipeline Definition

Open:

- [personal_finance_pipeline.yaml](/Users/streanor/Documents/Playground/data-platform/packages/manifests/pipelines/personal_finance_pipeline.yaml)

Look for:

- pipeline ID
- owner
- schedule
- jobs
- `depends_on` relationships

What success looks like:

- you can point to the pipeline ID
- you can explain which jobs run first and which jobs depend on earlier jobs

If something goes wrong:

- if the manifest feels opaque, compare it with
  [making-changes.md](/Users/streanor/Documents/Playground/data-platform/docs/tutorials/making-changes.md)
  after you finish this walkthrough

## Step 2: Identify The Execution Building Blocks

Open these inputs:

- [packages/sql/README.md](/Users/streanor/Documents/Playground/data-platform/packages/sql/README.md)
- [packages/python/README.md](/Users/streanor/Documents/Playground/data-platform/packages/python/README.md)
- [runtime-wiring.md](/Users/streanor/Documents/Playground/data-platform/docs/architecture/runtime-wiring.md)

What to notice:

- Go owns orchestration and queueing
- SQL owns the transparent analytical transforms
- Python is used only for bounded helper tasks where it is a better fit

What success looks like:

- you can explain why the platform is not "just Python scripts" and not "just
  SQL files"

If something goes wrong:

- if the architecture terms feel abstract, skim
  [infra-overview.md](/Users/streanor/Documents/Playground/data-platform/infra-overview.md)
  before continuing

## Step 3: Start A Known-Good Local Run

From the repo root:

```sh
cd /Users/streanor/Documents/Playground/data-platform
make smoke
```

What success looks like:

- the command exits `0`
- output includes `localhost smoke test passed`
- the output prints the API URL and a manual run ID

If something goes wrong:

1. use the troubleshooting section in
   [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md)
2. do not continue until you have one successful smoke run

## Step 4: Inspect The Control-Plane View Of The Run

Query the pipelines API:

```sh
curl http://127.0.0.1:8080/api/v1/pipelines
```

Look for:

- a recent run for `personal_finance_pipeline`
- overall run status
- per-job status
- recent event messages

What success looks like:

- you can find a run record
- you can see evidence that the run was queued and then executed

If something goes wrong:

1. if the API is unreachable, confirm the smoke stack is still running
2. if the run list is empty, rerun `make smoke`

## Step 5: Inspect The Worker Outputs

Look under the repo-local runtime tree:

- `var/data/`
- `var/artifacts/runs/<run_id>/`

Common examples include:

- `var/data/raw/raw_transactions.csv`
- `var/data/staging/staging_transactions_enriched.json`
- `var/data/intermediate/intermediate_category_monthly_rollup.json`
- `var/data/mart/mart_budget_vs_actual.json`
- `var/data/metrics/metrics_category_variance.json`

What success looks like:

- you can see that raw data becomes more curated as it moves through the layers
- you can see both persistent data outputs and run-scoped artifact snapshots

If something goes wrong:

1. if files are missing, confirm the run actually succeeded
2. if only some files exist, inspect worker logs or run artifacts for the
   failed job

## Step 6: Inspect The Dataset Catalog

Query the catalog:

```sh
curl http://127.0.0.1:8080/api/v1/catalog
```

Query one runtime profile:

```sh
curl "http://127.0.0.1:8080/api/v1/catalog/profile?asset_id=mart_budget_vs_actual"
```

What success looks like:

- the catalog returns assets with freshness, lineage, and ownership fields
- the profile returns row count and per-column information

If something goes wrong:

1. if the catalog is empty, confirm the run succeeded
2. if the profile fails, confirm the selected asset exists in the materialized data

## Step 7: Inspect The Analytics Layer

Query a curated dataset:

```sh
curl "http://127.0.0.1:8080/api/v1/analytics?dataset=mart_budget_vs_actual"
```

Query a metric:

```sh
curl "http://127.0.0.1:8080/api/v1/analytics?metric=metrics_category_variance"
```

What success looks like:

- both endpoints return curated rows
- the responses are shaped for product consumption rather than arbitrary SQL

If something goes wrong:

1. confirm the run completed successfully
2. confirm the selected dataset or metric exists

## Step 8: Inspect The Reporting Layer

Query saved dashboards:

```sh
curl http://127.0.0.1:8080/api/v1/reports
```

Then open the browser:

- `Dashboard`
- `Pipelines`
- `Datasets`
- `Metrics`
- `System`

What success looks like:

- the UI surfaces are reading platform-owned APIs rather than separate hidden state
- you can connect the same run to the dashboards, datasets, metrics, and system views

If something goes wrong:

1. if the UI is broken but the APIs work, inspect frontend logs
2. if both UI and API fail, go back to the startup runbook and confirm the stack

## Step 9: Map The Ownership Back To Code

When you are ready to go deeper, read these files next:

- [handler.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/orchestration/handler.go)
- [runner.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/execution/runner.go)
- [engine.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/transforms/engine.go)
- [handler.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/metadata/handler.go)
- [service.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/analytics/service.go)
- [handler.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/reporting/handler.go)

What success looks like:

- you can now tie one user-visible platform behavior to the packages that own it

## Final Mental Model

By the end of this tutorial, you should be able to explain this path in one
sentence:

manifest -> queue -> worker -> data outputs -> metadata/profile -> analytics
API -> reporting and operator UI
