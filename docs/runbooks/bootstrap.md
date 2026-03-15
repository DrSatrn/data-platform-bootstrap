# Compose Bootstrap Runbook

This runbook owns the packaged self-host startup path. It does not describe the
host-run debug flow. Use [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md)
first if you have not already completed one successful smoke run.

## What This Path Is For

Use this runbook when you want:

- the closest thing to the intended self-hosted local deployment
- PostgreSQL enabled as the preferred control-plane backend
- the packaged web service instead of the Vite dev server

## Role Requirements

- Anonymous: `GET /healthz`, `GET /api/v1/session`
- Viewer: browse UI pages and read APIs
- Editor: trigger runs, save dashboards
- Admin: admin terminal, `platformctl remote ...`

## Optional Compose Overrides

The tracked Compose file has safe defaults already. Only create `.env.compose`
if you want to override token or Postgres settings.

```sh
cd /Users/streanor/Documents/Playground/data-platform
cp .env.compose.example .env.compose
```

## Start The Stack

Working directory:

```sh
cd /Users/streanor/Documents/Playground/data-platform
```

Command:

```sh
make bootstrap
```

Expected success result:

- `docker compose ps` shows:
  - healthy `postgres`
  - completed `migrate`
  - healthy `api`
  - running `worker`
  - running `scheduler`
  - healthy `web`
- API is reachable at `http://127.0.0.1:8080/healthz`
- web is reachable at `http://127.0.0.1:3000`

## Verify The Stack

Health:

```sh
curl http://127.0.0.1:8080/healthz
```

Expected result:

- HTTP `200`
- JSON includes the configured environment and data root

Session:

```sh
curl http://127.0.0.1:8080/api/v1/session
```

Expected result:

- HTTP `200`
- anonymous principal if no bearer token is supplied

Catalog:

```sh
curl -H "Authorization: Bearer viewer-token" http://127.0.0.1:8080/api/v1/catalog
```

Expected result:

- HTTP `200`
- JSON `assets` array is present

## Run One Manual Pipeline

Role required: `editor`

```sh
curl -X POST \
  -H "Authorization: Bearer editor-token" \
  -H "Content-Type: application/json" \
  -d '{"pipeline_id":"personal_finance_pipeline"}' \
  http://127.0.0.1:8080/api/v1/pipelines
```

Expected result:

- HTTP `202`
- response includes `"status":"queued"` in the returned run payload

## Verify Downstream Outputs

Role required: `viewer`

```sh
curl -H "Authorization: Bearer viewer-token" http://127.0.0.1:8080/api/v1/pipelines
curl -H "Authorization: Bearer viewer-token" "http://127.0.0.1:8080/api/v1/analytics?dataset=mart_budget_vs_actual"
curl -H "Authorization: Bearer viewer-token" "http://127.0.0.1:8080/api/v1/catalog/profile?asset_id=mart_budget_vs_actual"
```

Expected result:

- the latest run reaches `succeeded`
- analytics returns rows
- dataset profile returns row count and profiled columns

## Preferred Control-Plane Reality

In this mode:

- PostgreSQL is the preferred source of truth for runs, queue state, artifacts,
  dashboards, audit events, and projected metadata
- filesystem stores remain as local-first fallback and recovery material, not
  the preferred live read path
- DuckDB remains the analytical execution store

If PostgreSQL is unavailable, the runtime falls back to filesystem-backed
stores, but that is a fallback mode, not the preferred packaged behavior.

## If This Fails, Check Next

1. `docker compose -f infra/compose/docker-compose.yml ps`
2. `docker compose -f infra/compose/docker-compose.yml logs migrate`
3. `docker compose -f infra/compose/docker-compose.yml logs api`
4. `docker compose -f infra/compose/docker-compose.yml logs worker`
5. `docker compose -f infra/compose/docker-compose.yml logs scheduler`
6. confirm `127.0.0.1:8080` and `127.0.0.1:3000` are not already occupied

## Recovery Follow-Up

After the stack is healthy, capture a recovery point:

```sh
make backup
```

Expected result:

- command exits `0`
- output includes `backup bundle created:`
- output includes `backup bundle verified:` after verification
