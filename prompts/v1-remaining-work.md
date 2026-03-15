# Post-Model v1 Coordination — Remaining Work

**Date:** March 15, 2026  
**Context:** All 3 model contracts complete. This document tracks what remains for v1.

---

## Model Completion Summary

| Model | Contract Status | Verification Status |
|-------|----------------|-------------------|
| Model 1 (Backend) | ✅ All 4 tasks complete | ⚠️ Targeted tests green. `go test ./...` fails on 4 pre-existing external-tool tests. |
| Model 2 (Frontend) | ✅ All 4 tasks complete | ✅ `npm run build` and `npm test` both pass. |
| Model 3 (Platform) | ✅ 5 of 6 tasks complete | ⚠️ UAT deferred (was blocked on Model 1/2 completions). |

## What Models Delivered

### Model 1
- `transforms/engine_test.go` — 5 test functions covering the DuckDB execution layer
- `quality/service_test.go` — expanded with 5+ additional tests
- Deleted `ingestion/` stub (correct v1 decision)
- Created `backend/test/integration_test.go` skeleton

### Model 2
- URL routing via `react-router-dom` — all 6 pages bookmarkable
- Dashboard component extraction — 6 new components in `components/dashboard/`
- `LoadingSpinner`, `ErrorMessage`, `ErrorBoundary` components
- 4 new page test files (Dashboard, Pipelines, Datasets, Metrics)
- Global CSS updates for loading/error states

### Model 3
- Rewrote `guide-wire.md` (89 lines, clean)
- Rewrote `plan.md` (62 lines, clean)
- Created `.github/workflows/ci.yml` (valid, with CGO prerequisites)
- Updated `codex.md` with v1 review handoff

---

## Remaining v1 Gaps (Ordered by Priority)

### 1. Fix Pre-Existing External-Tool Test Failures (BLOCKER)

**Root cause:** 4 tests in `backend/internal/execution/` fail with `context deadline exceeded`. All tests use fake shell scripts that should complete in milliseconds but are constrained by a 2-second job-level timeout (`Timeout: "2s"` in the test job spec).

**Failing tests:**
- `TestRunExternalToolMirrorsLogsAndArtifacts` (external_tool_test.go)
- `TestRunExternalToolFailsForNonZeroExitAndMirrorsLogs` (external_tool_failures_test.go)
- `TestRunExternalToolFailsWhenRequiredArtifactIsMissing` (external_tool_failures_test.go)
- `TestExternalToolOutputsAreInspectableThroughStorageService` (external_tool_operator_inspection_test.go)

**Files to fix:**
- `backend/internal/execution/external_tool_test.go` — line 47: change `Timeout: "2s"` to `Timeout: "30s"`
- `backend/internal/execution/external_tool_failures_test.go` — the `externalToolJob` helper at line 128: change `Timeout: "2s"` to `Timeout: "30s"`
- `backend/internal/execution/external_tool_operator_inspection_test.go` — line 47: change `Timeout: "2s"` to `Timeout: "30s"`

**Verification:** `cd backend && go test ./...` must pass.

### 2. Run Full UAT (HIGH)

Model 3 deferred this. Now that Models 1 and 2 are complete, UAT should be executed:
1. `make bootstrap`
2. Follow `uat-checklist.md` end to end
3. Annotate every item with PASS/FAIL

### 3. Update `README.md` (MEDIUM)

The README should reflect the current state after Model 2's frontend changes:
- URL routing now exists (bookmarkable pages)
- Dashboard page has been decomposed
- Loading/error states exist across all pages
- The management surface is integrated

### 4. Clean Up Repo Root (LOW)

Model 3 proposed these moves (documented in `prompts/model3-completion.md`):
- `temp-model1-frontend-wire-plan.md` → delete (stale)
- `uat-checklist.md` → move to `docs/runbooks/` after UAT complete
- `infra-overview.md` → move to `docs/architecture/`
- `new-thread-eng-feedback.md` → move to `docs/decisions/` (closed contract)

### 5. Update `v1-review-coordination-plan.md` (LOW)

The original review should be annotated with post-model status to remain accurate.

---

## Updated Coordinator Checklist

- [ ] Fix external-tool test timeouts (Gap 1)
- [ ] Verify `go test ./...` passes globally
- [ ] Run `make bootstrap` → full UAT → annotate checklist (Gap 2)
- [ ] Update README.md (Gap 3)
- [ ] Clean up repo root (Gap 4)
- [ ] Update v1-review-coordination-plan.md with post-model addendum (Gap 5)
- [ ] Final `make smoke` verification
- [ ] Tag the v1 release
