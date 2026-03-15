# Host-Run End-to-End Runbook

This runbook owns the host-run binary path. Use it when you are debugging the
API, worker, scheduler, or web app individually. Do not use this as the first
setup document; start with
[quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md).

## What This Path Is For

Use this runbook when you want to:

- run `platform-api`, `platform-worker`, and `platform-scheduler` directly
- inspect local logs and behavior without Docker
- debug config or code changes with tighter feedback loops

## Configuration Reality

- Host-run binaries auto-load `.env` and `.env.local`.
- Copy [.env.example](/Users/streanor/Documents/Playground/data-platform/.env.example)
  to `.env` from the repo root.
- Those values are intended for `go run` from `backend/`.
- Compose-only overrides belong in `.env.compose`, not `.env`.

## Role Requirements

- Viewer: read platform pages and APIs
- Editor: trigger manual runs and save dashboards
- Admin: admin terminal and `platformctl remote ...`

## Prepare Local Config

Working directory:

```sh
cd /Users/streanor/Documents/Playground/data-platform
cp .env.example .env
```

Expected result:

- `.env` exists at repo root
- it contains local filesystem paths such as `../var/data`

If this fails, check next:

1. confirm you are in the repo root
2. confirm `.env.example` still exists

## Optional Postgres Setup

If you want the preferred Postgres control-plane mode while still running
processes directly, start only Postgres through Compose:

```sh
docker compose -f infra/compose/docker-compose.yml up -d postgres
cd backend
go run ./cmd/platformctl migrate
```

Expected result:

- migration command exits `0`
- Postgres is healthy in `docker compose ps`

If you skip this, the runtime will use the filesystem fallback stores.

## Start The Services

API:

```sh
cd /Users/streanor/Documents/Playground/data-platform/backend
go run ./cmd/platform-api
```

Worker:

```sh
cd /Users/streanor/Documents/Playground/data-platform/backend
go run ./cmd/platform-worker
```

Scheduler:

```sh
cd /Users/streanor/Documents/Playground/data-platform/backend
go run ./cmd/platform-scheduler
```

Web:

```sh
cd /Users/streanor/Documents/Playground/data-platform/web
npm run dev
```

Expected success result:

- API responds on `http://127.0.0.1:8080/healthz`
- web responds on `http://127.0.0.1:3000`
- worker logs show polling
- scheduler logs show the configured tick interval

## Trigger One Manual Run

Role required: `editor`

CLI path:

```sh
curl -X POST \
  -H "Authorization: Bearer editor-token" \
  -H "Content-Type: application/json" \
  -d '{"pipeline_id":"personal_finance_pipeline"}' \
  http://127.0.0.1:8080/api/v1/pipelines
```

Browser path:

1. open `http://127.0.0.1:3000`
2. paste an `editor` token into the sidebar
3. open `Pipelines`
4. click `Run now`

Expected result:

- the returned run is `queued`
- the worker moves it to `running`
- the final state becomes `succeeded`

## Verify Outputs

Role required: `viewer`

```sh
curl -H "Authorization: Bearer viewer-token" http://127.0.0.1:8080/api/v1/pipelines
curl -H "Authorization: Bearer viewer-token" http://127.0.0.1:8080/api/v1/catalog
curl -H "Authorization: Bearer viewer-token" "http://127.0.0.1:8080/api/v1/catalog/profile?asset_id=mart_budget_vs_actual"
curl -H "Authorization: Bearer viewer-token" "http://127.0.0.1:8080/api/v1/analytics?metric=metrics_category_variance"
```

Expected result:

- pipelines response shows a recent `succeeded` run
- catalog response includes assets with freshness state
- profile response includes `row_count` and `columns`
- analytics response includes metric rows

Expected local files under `var/`:

- `var/data/raw/raw_transactions.csv`
- `var/data/staging/staging_transactions_enriched.json`
- `var/data/intermediate/intermediate_category_monthly_rollup.json`
- `var/data/mart/mart_budget_vs_actual.json`
- `var/data/metrics/metrics_category_variance.json`

## If This Fails, Check Next

1. API healthy but no progress:
   confirm worker is running and shares the same `PLATFORM_DATA_ROOT`
2. run stays queued:
   confirm worker logs show queue polling
3. browser action fails:
   confirm you used an `editor` token rather than `viewer`
4. admin terminal fails:
   confirm you used an `admin` token
5. profile endpoint fails:
   confirm the run has completed and the materialized asset exists
