# Making Changes Guide

This guide explains how to extend or modify the platform safely. It is meant
for contributors who understand the product goal but do not yet know where each
kind of change belongs.

## Rule One

Put changes in the subsystem that owns the behavior. Do not add code to random
shared locations just because it is convenient.

## Common Change Types

### Add a new API endpoint

Usually change:

- `backend/internal/<domain>/handler.go`
- `backend/internal/app/runtime.go`

Example domains:

- orchestration
- metadata
- analytics
- reporting
- admin

### Add a new pipeline or job

Usually change:

- `packages/manifests/pipelines/`
- `packages/manifests/assets/`
- maybe `packages/manifests/metrics/`
- maybe `packages/sql/`

If the worker needs new execution behavior, then also change:

- `backend/internal/execution/runner.go`

### Add a new SQL transform

Usually change:

- `packages/sql/transforms/`
- pipeline manifest `transform_ref`

If it should surface through analytics:

- update `backend/internal/analytics/service.go`

If it introduces a new persistent layer artifact:

- add the corresponding asset manifest under `packages/manifests/assets/`

### Add a new metric

Usually change:

- `packages/sql/metrics/`
- `packages/manifests/metrics/`
- pipeline publish step if needed
- analytics service if it should be queryable
- metrics browser frontend if it should be discoverable in the UI

### Add a new catalog field or metadata surface

Usually change:

- `backend/internal/metadata/models.go`
- `backend/internal/metadata/catalog.go`
- `backend/internal/metadata/handler.go`
- maybe `backend/internal/db/metadata_store.go`
- relevant frontend data hooks and pages

### Add a new dashboard or widget type

Usually change:

- `packages/dashboards/`
- `backend/internal/reporting/store.go`
- `backend/internal/reporting/handler.go`
- frontend dashboard feature files under `web/src/features/dashboard/`
- `web/src/pages/DashboardPage.tsx`

### Add a new admin terminal command

Usually change:

- `backend/internal/admin/service.go`
- possibly audit behavior if the command is privileged

### Add a Python-powered helper

Use Python only when it clearly improves the data/runtime side:

- connectors
- profiling
- quality
- specialized transform helpers

Do not move:

- API ownership
- queueing
- scheduling
- orchestration state machine

Current implementation reference:

- [runtime.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/python/runtime.go)
- [enrich_transactions.py](/Users/streanor/Documents/Playground/data-platform/packages/python/tasks/enrich_transactions.py)

## How To Change The Frontend Safely

The frontend is feature-oriented. The safest pattern is:

1. add or update API contract in the backend first
2. add or update a feature hook under `web/src/features/...`
3. render through a page or component
4. run `npm run build`

Good starting points:

- datasets: `web/src/features/datasets/`
- dashboards: `web/src/features/dashboard/`
- pipelines: `web/src/features/pipelines/`
- system: `web/src/features/system/`

## How To Change The Backend Safely

The backend is a modular monolith. The safest pattern is:

1. decide the owning domain
2. update domain model or service
3. update handler
4. wire it in `runtime.go`
5. add or update tests
6. run `go test ./...`

## Minimum Validation After A Change

For most meaningful changes:

```sh
cd backend
go test ./...
go run ./cmd/platformctl validate-manifests
```

```sh
cd web
npm run build
```

And then from repo root:

```sh
make smoke
```

For deployment-affecting changes:

```sh
make compose-smoke
make backup
```

## When To Update Docs

Update docs whenever you change:

- service responsibilities
- runtime commands
- setup instructions
- data flow
- control-plane behavior
- backup or operational procedures

If a new contributor would be surprised by the change, the docs are not done.
