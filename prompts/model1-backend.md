# Model 1 Prompt — Backend Hardening & Test Coverage

You are working on the `data-platform` repository. You are Model 1 of 3 parallel implementation models. You have fast mode enabled and full IDE autonomy.

## Read First

Read these files before doing anything:
- `v1-review-coordination-plan.md` (your task contracts are in Section 8)
- `codex.md` (project context and architectural rules)
- `backend/internal/transforms/engine.go` (your primary test target)
- `backend/internal/quality/service.go` (your secondary test target)
- `backend/internal/ingestion/README.md` (stub you need to resolve)

## Your Mission

Make the backend bulletproof for a v1 release. Your job is to fill critical test gaps, harden error handling, and clean up structural stubs. You are NOT building new features. You are making existing features trustworthy.

## Exact Tasks (In Order)

### Task 1: Add tests for `transforms/engine.go`
Create `backend/internal/transforms/engine_test.go` covering:
- `MaterializeRawTables` (happy path + missing file error)
- `RunTransform` (valid ref + invalid ref)
- `QueryRows` (valid query + empty result)
- `QueryRowsFromFile` (valid path + missing file)
- `RunMetric` (valid metric + invalid metric)

**Completion signal:** `cd backend && go test ./internal/transforms/...` passes.

### Task 2: Expand quality service tests
Expand `backend/internal/quality/service_test.go` with at least 5 additional test functions covering:
- Empty query results
- Invalid SQL file paths
- Nil/empty inputs
- Multiple quality checks in sequence
- Error message formatting

**Completion signal:** `cd backend && go test ./internal/quality/...` passes.

### Task 3: Resolve the `ingestion/` stub
The `backend/internal/ingestion/` package contains only a README. No Go source files.
Ingestion logic currently lives directly in `execution/runner.go` via `copySampleFile`.

**Decision required:** Either:
- **(A) Delete the stub** — remove `backend/internal/ingestion/` entirely if extracting it would require touching `execution/runner.go` or `runtime.go` (which are forbidden files). Update the README deletion in your completion note.
- **(B) Implement a thin adapter** — if and only if you can create a working `ingestion.Service` that `runner.go` could later call, WITHOUT editing `runner.go` itself. This is a preparation-only move.

Prefer option (A) for v1 simplicity. Option (B) only if you are confident it's clean.

**Completion signal:** Package either has real Go source or is deleted. `go test ./...` passes.

### Task 4 (Stretch): Create integration test skeleton
Create `backend/test/integration_test.go` with:
- A single test function that boots a `config.Settings` from `.env.example` defaults
- Validates that required paths and config values are non-empty
- Documents what a future integration test suite would look like

**Completion signal:** `cd backend && go test ./test/...` passes (or documents why it can't run in CI).

## Allowed Files

You may create or edit:
- `backend/internal/transforms/engine_test.go` (NEW)
- `backend/internal/quality/service_test.go` (EXPAND)
- `backend/internal/ingestion/` (DELETE or IMPLEMENT)
- `backend/internal/execution/*_test.go` (NEW test files only)
- `backend/internal/analytics/*_test.go` (NEW test files only)
- `backend/internal/storage/*_test.go` (NEW test files only)
- `backend/test/` (NEW integration test files)
- `backend/internal/shared/` (utilities if needed for tests)

## Forbidden Files

Do NOT edit these under any circumstances:
- ❌ `backend/internal/app/runtime.go`
- ❌ `backend/internal/execution/runner.go` (read-only, do not refactor)
- ❌ `backend/internal/db/` (all files)
- ❌ `backend/internal/authz/` (all files)
- ❌ `backend/internal/orchestration/` (all files, except new `*_test.go`)
- ❌ `backend/internal/scheduler/` (all files)
- ❌ `backend/internal/reporting/` (all files)
- ❌ `backend/internal/backup/` (all files)
- ❌ `backend/internal/audit/` (all files)
- ❌ `backend/internal/opsview/` (all files)
- ❌ `web/` (all frontend files)
- ❌ `infra/` (all infra files)
- ❌ `Makefile`
- ❌ `*.md` files in repo root (documentation is Model 3's domain)

## Inputs

- The existing backend codebase as-is
- `.env.example` for config defaults

## Expected Outputs

1. `backend/internal/transforms/engine_test.go` — minimum 5 test functions
2. `backend/internal/quality/service_test.go` — expanded with minimum 5 new test functions
3. `backend/internal/ingestion/` directory resolved (deleted or implemented)
4. (Stretch) `backend/test/integration_test.go` — skeleton
5. A completion note (see below)

## Stop Conditions

Stop and escalate to the reviewer if:
- You find you MUST edit `runtime.go` or `runner.go` to accomplish a task
- DuckDB CGO prevents tests from running and you cannot create fixtures
- You discover a bug in existing code (document it, do not fix without approval)
- You need to add a new dependency to `go.mod`

## How To Report Completion

When all tasks are done, create a file `prompts/model1-completion.md` containing:

```markdown
# Model 1 Completion Report

## Files Changed
- (list every file created, modified, or deleted)

## Verification
- `cd backend && go test ./...` — PASS/FAIL
- `cd backend && go run ./cmd/platformctl validate-manifests` — PASS/FAIL

## What Was NOT Changed
- (list forbidden files you intentionally did not touch)

## Escalation Items
- (any bugs found, constraints discovered, or help needed)

## Decision Log
- ingestion/ package: (deleted / implemented / reason)
```
