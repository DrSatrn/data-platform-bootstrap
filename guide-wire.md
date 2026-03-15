# Guide Wire
This file is the live coordination rail for the v3 three-model push.

## Goal
Finish the remaining pre-production gaps without merge collisions:

- Model 1: orchestration retries/idempotency, DB ingestion, retention GC
- Model 2: semantic rollups, CSV export, shareable dashboard links
- Model 3: alerting, Prometheus metrics, tagged release automation

## Current Prompt Rail
Read these first:

- `prompts/v3-review-coordinator-report.md`
- `prompts/v3-model1-backend.md`
- `prompts/v3-model2-frontend.md`
- `prompts/v3-model3-platform.md`

The old `prompts/coordinator.md` and v1/v2 prompt files are no longer live.

## Hot Files
Edit surgically and only when the current lane requires it:

- `backend/internal/app/runtime.go`
- `backend/internal/config/config.go`
- `backend/cmd/platformctl/main.go`
- `web/src/app/App.tsx`
- `infra/compose/docker-compose.yml`
- `Makefile`

## Ownership Snapshot

- Model 1: backend domain and orchestration work
- Model 2: frontend/dashboard/product wiring
- Model 3: backend ops surfaces, CI/CD, coordination docs, release assets

Model 3 should avoid `web/src/**` and routed UI work unless the live prompt
explicitly says otherwise.

## Verified Model 3 State

- webhook alerting now exists for failed runs and stale assets
- `/api/v1/system/metrics` now serves Prometheus-formatted telemetry
- tagged GitHub Actions builds now package `platformctl` plus `docs/runbooks`
  through Makefile-driven release targets
- the backend module now explicitly includes the MySQL driver required by the
  ingestion exporter and execution-path tests

## Current Verification Commands

- `cd backend && go test ./internal/alerting ./internal/config ./internal/metadata ./internal/observability ./internal/scheduler ./internal/execution ./internal/app`
- `ruby -e 'require "yaml"; YAML.load_file(".github/workflows/ci.yml"); puts "yaml ok"'`
- `make -n release-platformctl TARGET_OS=linux TARGET_ARCH=amd64 TARGET_DIR=dist/linux-amd64/bin`

## Merge Protocol

1. Use the prompt files above as the source of truth over older handoff notes.
2. Prefer additive files first; keep hot-file diffs narrow and explicit.
3. Record verification and any touched hot files in the closeout.
