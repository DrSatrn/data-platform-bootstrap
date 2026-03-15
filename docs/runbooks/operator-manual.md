# Operator Manual

This document is the primary practical manual for building, running,
maintaining, and validating the platform. It is written for someone who did not
build the system originally and needs a reliable path to operate it.

## Start Here

If you are new to the repo, read in this order:

1. [README.md](/Users/streanor/Documents/Playground/data-platform/README.md)
2. [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md)
3. [runtime-wiring.md](/Users/streanor/Documents/Playground/data-platform/docs/architecture/runtime-wiring.md)
4. this document

## Tooling Prerequisites

Host requirements:

- Go
- Node and npm
- Python 3
- Docker or OrbStack
- Apple Silicon compatible toolchain
- host C/C++ build tools because DuckDB uses CGO

Verify them with `make doctor` before assuming the host is ready.

## First Build

From repo root:

```sh
make doctor
make build
```

What that does:

- checks the expected local toolchain assumptions
- builds the Go binaries for `darwin/arm64`
- builds the React app

## Access Model Summary

- Anonymous: `GET /healthz`, `GET|POST|DELETE /api/v1/session`
- Bootstrap admin: `PLATFORM_ADMIN_TOKEN` for first-run admin and recovery
- Viewer: read APIs and read-only product pages
- Editor: manual run triggers, dashboard writes, and metadata annotation edits
- Admin: admin terminal, user management, and `platformctl remote ...`

Normal operation should use native users plus `/api/v1/session` login. If a
command uses `platformctl remote`, it is admin-only because it goes through
the admin terminal endpoint.

## Control-Plane Source Of Truth

Use the `Source Of Truth` card on the System page or `GET /api/v1/system/overview`
when you need to confirm which backend is authoritative for each subsystem.

Current intended behavior:

- runs: PostgreSQL primary when available, filesystem mirror/fallback
- queue: PostgreSQL primary when available, filesystem fallback
- artifacts: filesystem bytes are authoritative, PostgreSQL is metadata/index only
- dashboards: PostgreSQL primary when available, manifest seeding only, filesystem fallback
- audit: PostgreSQL primary when available, filesystem mirror/fallback
- metadata: PostgreSQL primary when available for catalog reads and runtime
  annotations, manifest seeding for structural fields, manifest loader fallback
  only when PostgreSQL is unavailable

If the running stack disagrees with that summary, trust the System page or
`/api/v1/system/overview` over this document.

## Main Ways To Run The Platform

### 1. Canonical first-run path

Use [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md).
That document owns first-run setup. This manual assumes you already completed
one successful smoke or packaged boot.

### 2. Fast local smoke path

Use this when you want the quickest repeatable confidence check:

```sh
make smoke
```

This starts an isolated localhost API, worker, and scheduler, validates a real
run, validates artifacts, validates the admin terminal, and now validates
backup creation and verification too. When PostgreSQL is not enabled in that
host-run flow, the smoke script explicitly skips metadata editing and native
identity checks rather than pretending they are active.

### 3. Packaged self-host style stack

Use this when you want the closest thing to the intended local deployment:

```sh
make bootstrap
```

Then open:

- `http://127.0.0.1:3000`
- `http://127.0.0.1:8080/healthz`

### 4. Run services individually

Use this when you are debugging one process at a time.

API:

```sh
cd backend
go run ./cmd/platform-api
```

Worker:

```sh
cd backend
go run ./cmd/platform-worker
```

Scheduler:

```sh
cd backend
go run ./cmd/platform-scheduler
```

Web dev:

```sh
cd web
npm run dev
```

## Core Operator Commands

Manifest validation:

```sh
cd backend
go run ./cmd/platformctl validate-manifests
```

Apply migrations:

```sh
cd backend
go run ./cmd/platformctl migrate
```

Check remote status:

```sh
cd backend
go run ./cmd/platformctl remote --token <token> status
```

List semantic metrics:

```sh
curl http://127.0.0.1:8080/api/v1/metrics
```

Fetch a runtime dataset profile:

```sh
curl "http://127.0.0.1:8080/api/v1/catalog/profile?asset_id=mart_budget_vs_actual"
```

Update metadata annotations:

```sh
curl -X PATCH \
  -H "Authorization: Bearer <editor-or-admin-token>" \
  -H "Content-Type: application/json" \
  -d '{"asset_id":"mart_budget_vs_actual","owner":"platform-governance","description":"Runtime annotation override","documentation_refs":["docs/runtime-annotation.md"],"quality_check_refs":["quality_runtime_annotation"],"column_descriptions":[{"name":"month","description":"Month grain override"}]}' \
  http://127.0.0.1:8080/api/v1/catalog
```

Trigger a pipeline:

```sh
cd backend
go run ./cmd/platformctl remote --token <token> trigger personal_finance_pipeline
```

Create a native user:

```sh
curl -X POST \
  -H "Authorization: Bearer local-dev-admin-token" \
  -H "Content-Type: application/json" \
  -d '{"username":"operator","display_name":"Operator","role":"editor","password":"operator-password"}' \
  http://127.0.0.1:8080/api/v1/admin/users
```

Sign in and get a session token:

```sh
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"username":"operator","password":"operator-password"}' \
  http://127.0.0.1:8080/api/v1/session
```

List artifacts for a run:

```sh
cd backend
go run ./cmd/platformctl remote --token <token> "artifacts <run_id>"
```

Run benchmark suite:

```sh
make benchmark
```

Create and verify a backup:

```sh
make backup
```

Run a safe restore drill:

```sh
make restore-drill
```

Run the full restore proof:

```sh
make restore-e2e
```

Run a direct restore into the configured runtime roots:

```sh
cd backend
go run ./cmd/platformctl backup restore --file ../var/backups/<bundle-name>.tar.gz --yes
```

## Normal Maintenance Tasks

### Confirm the stack is healthy

- check `http://127.0.0.1:8080/healthz`
- check the System page
- check `platformctl remote status`
- confirm queue, run, and backup summary cards look sane in the System page

### Confirm manifests are still valid

```sh
cd backend
go run ./cmd/platformctl validate-manifests
```

### Confirm the packaged deployment still boots

```sh
make compose-smoke
```

### Capture a benchmark baseline

```sh
make benchmark
```

### Capture a recovery point

```sh
make backup
```

### Prove recovery still works

```sh
make restore-drill
make restore-e2e
```

## Where To Look When Something Breaks

API behavior issues:

- [runtime.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/app/runtime.go)
- [handler.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/orchestration/handler.go)

Worker execution issues:

- [worker.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/execution/worker.go)
- [runner.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/execution/runner.go)
- [engine.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/transforms/engine.go)

Metadata/catalog issues:

- [handler.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/metadata/handler.go)
- [catalog.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/metadata/catalog.go)

Reporting/dashboard issues:

- [handler.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/reporting/handler.go)
- [store.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/reporting/store.go)
- [DashboardPage.tsx](/Users/streanor/Documents/Playground/data-platform/web/src/pages/DashboardPage.tsx)
- [MetricsPage.tsx](/Users/streanor/Documents/Playground/data-platform/web/src/pages/MetricsPage.tsx)

Current reporting runtime behavior:

- dashboards carry owner, tags, and shared-role intent metadata
- dashboard-wide default filters and saved presets now shape widget hydration
- widget-level filters still apply, but they layer on top of dashboard context

RBAC/audit issues:

- [service.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/authz/service.go)
- [store.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/audit/store.go)

Python runtime issues:

- [runtime.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/python/runtime.go)
- [enrich_transactions.py](/Users/streanor/Documents/Playground/data-platform/packages/python/tasks/enrich_transactions.py)
- [profile_asset.py](/Users/streanor/Documents/Playground/data-platform/packages/python/tasks/profile_asset.py)

Dataset profiling issues:

- [profile.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/metadata/profile.go)
- [profile_handler.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/metadata/profile_handler.go)
- [DatasetsPage.tsx](/Users/streanor/Documents/Playground/data-platform/web/src/pages/DatasetsPage.tsx)

## Safe Change Rules

- keep Go as the control-plane language
- add Python only for bounded data/runtime helpers
- do not introduce Redis unless the current queue model becomes a real
  bottleneck
- keep tracked env/config files free of secrets
- keep API and web loopback-first by default
- update docs when behavior changes

## What Is Still Not Finished

Be explicit about current limits:

- restore automation is not implemented yet
- scheduler semantics are not a full cron engine
- native identity now lives in PostgreSQL-backed users and sessions, but the
  bootstrap admin token is still the recovery path when PostgreSQL is absent
- the platform proves one strong domain slice, not many
- reporting UX is real but not yet fully polished
- restore automation is still not built on top of the backup bundle format

That means this repo is already a strong self-hosted platform build, but it
still needs more product and operational depth before it should be called
finished.
