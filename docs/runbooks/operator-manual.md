# Operator Manual

This document is the primary practical manual for building, running,
maintaining, and validating the platform. It is written for someone who did not
build the system originally and needs a reliable path to operate it.

## Start Here

If you are new to the repo, read in this order:

1. [README.md](/Users/streanor/Documents/Playground/data-platform/README.md)
2. [runtime-wiring.md](/Users/streanor/Documents/Playground/data-platform/docs/architecture/runtime-wiring.md)
3. [bootstrap.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/bootstrap.md)
4. this document

## Tooling Prerequisites

Host requirements:

- Go
- Node and npm
- Docker or OrbStack
- Apple Silicon compatible toolchain
- host C/C++ build tools because DuckDB uses CGO

You already installed the main host toolchains, which is enough for normal dev
and validation flows.

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

## Main Ways To Run The Platform

### 1. Fast local smoke path

Use this when you want the quickest confidence check:

```sh
make smoke
```

This starts an isolated localhost API, worker, and scheduler, validates a real
run, validates artifacts, validates the admin terminal, and now validates
backup creation and verification too.

### 2. Packaged self-host style stack

Use this when you want the closest thing to the intended local deployment:

```sh
make bootstrap
```

Then open:

- `http://127.0.0.1:3000`
- `http://127.0.0.1:8080/healthz`

### 3. Run services individually

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

Trigger a pipeline:

```sh
cd backend
go run ./cmd/platformctl remote --token <token> trigger personal_finance_pipeline
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

## Normal Maintenance Tasks

### Confirm the stack is healthy

- check `http://127.0.0.1:8080/healthz`
- check the System page
- check `platformctl remote status`

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

RBAC/audit issues:

- [service.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/authz/service.go)
- [store.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/audit/store.go)

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
- auth is lightweight token-based RBAC, not a full identity system
- the platform proves one strong domain slice, not many
- reporting UX is real but not yet fully polished

That means this repo is already a strong self-hosted platform build, but it
still needs more product and operational depth before it should be called
finished.
