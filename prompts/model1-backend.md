# Model 1 Prompt — Test Timeout Fix + Backend Test Coverage Blitz + Error Response Audit

You are working on the `data-platform` repository. You are Model 1 with fast mode and full IDE autonomy. Your previous contract is complete. This is your follow-up work block.

## Read First

- `prompts/model1-completion.md` (your previous work and results)
- `prompts/v1-remaining-work.md` (Gap 1 is your first task)
- `codex.md` (project context)

## Mission

Three objectives in order:
1. Fix the pre-existing external-tool test timeouts so `go test ./...` passes globally
2. Expand backend test coverage across undercovered packages
3. Audit API error response consistency and document findings

## Task 1: Fix External-Tool Test Timeouts (5 minutes)

4 tests fail because the job-level `Timeout: "2s"` is too tight on some machines. Change it to `"30s"` in all 3 files — the fake scripts complete in milliseconds so this adds headroom without slowing tests.

### Files to fix:
- `backend/internal/execution/external_tool_test.go` — line 47: `Timeout: "2s"` → `Timeout: "30s"`
- `backend/internal/execution/external_tool_failures_test.go` — line 128 in `externalToolJob` helper: `Timeout: "2s"` → `Timeout: "30s"`
- `backend/internal/execution/external_tool_operator_inspection_test.go` — line 47: `Timeout: "2s"` → `Timeout: "30s"`

**Completion signal:** `cd backend && go test ./internal/execution/...` passes.

## Task 2: Backend Test Coverage Blitz

Write new test files for undercovered packages. Go in this priority order:

### 2a: `admin/service.go` — ZERO test coverage (15KB of untested code)
Create `backend/internal/admin/service_test.go`.

Read `admin/service.go` to understand the command surface. Test at minimum:
- Command dispatch for known commands (e.g. `benchmark`, `backup`, `help`)
- Unknown/invalid command handling
- Output formatting
- Error propagation from subcommands
- Any permission/role checks that exist

### 2b: `analytics/service.go` — Thin coverage (485 lines, 1 test file)
Create `backend/internal/analytics/service_deeper_test.go` (don't overwrite the existing test file).

Test at minimum:
- Constrained query validation (invalid dataset refs, invalid group-by columns)
- Error responses when DuckDB is unavailable or not initialized
- Fallback logic (DuckDB → artifact files → sample data)
- Edge cases: empty datasets, very large row limits, nil filters

### 2c: `metadata/` — Enrich coverage
Create `backend/internal/metadata/enrichment_test.go`.

Test at minimum:
- Freshness derivation from materialization timestamps
- Coverage summary computation (doc coverage, quality coverage)
- Lineage edge extraction from manifests
- Nil/empty manifest inputs
- Assets with missing optional fields

### 2d: `observability/` — Error path coverage
Create `backend/internal/observability/error_paths_test.go`.

Test at minimum:
- Health endpoint when services are degraded
- Metrics handler with zero recorded observations
- Log buffer when empty
- System overview with partial data

### 2e: `config/` — Malformed input coverage
Expand existing tests or create `backend/internal/config/config_edge_test.go`.

Test at minimum:
- Missing required paths
- Invalid timeout values
- Conflicting configuration combinations
- Empty string values for paths that should have defaults

**Completion signal for Task 2:** All new test files compile and pass with `cd backend && go test ./...`.

## Task 3: API Error Response Audit

Read every HTTP handler file in the backend and check whether error responses are consistent. Produce a report at `prompts/api-error-audit.md`.

Handler files to read:
- `backend/internal/admin/handler.go`
- `backend/internal/analytics/handler.go` (if it exists, or check where analytics routes are registered)
- `backend/internal/authz/handler.go`
- `backend/internal/metadata/handler.go`
- `backend/internal/observability/handlers.go`
- `backend/internal/orchestration/handler.go` (if it exists)
- `backend/internal/quality/handler.go`
- `backend/internal/reporting/handler.go` (or wherever dashboard API is served)
- `backend/internal/storage/handler.go` (or wherever artifact API is served)
- `backend/internal/opsview/handler.go`
- `backend/internal/app/runtime.go` (check route registration)

For each handler, document:
| Endpoint | Method | Success Status | Error Status | Error Format | Consistent? |
|----------|--------|---------------|-------------|-------------|-------------|

Flag any handlers that:
- Return plain text errors instead of JSON
- Use inconsistent HTTP status codes (e.g. 500 for a 400-class error)
- Don't set `Content-Type: application/json` on error responses
- Swallow errors silently
- Return raw Go error strings to the client (information leak risk)

Do NOT fix the handlers — just document findings. Fixes should be a separate task.

## Allowed Files

You may create or edit:
- `backend/internal/execution/external_tool_test.go` (fix timeout)
- `backend/internal/execution/external_tool_failures_test.go` (fix timeout)
- `backend/internal/execution/external_tool_operator_inspection_test.go` (fix timeout)
- `backend/internal/admin/service_test.go` (NEW)
- `backend/internal/analytics/service_deeper_test.go` (NEW)
- `backend/internal/metadata/enrichment_test.go` (NEW)
- `backend/internal/observability/error_paths_test.go` (NEW)
- `backend/internal/config/config_edge_test.go` (NEW)
- `prompts/api-error-audit.md` (NEW)
- `prompts/model1-completion.md` (UPDATE with addendum)

## Forbidden Files

Do NOT edit these under any circumstances:
- ❌ Any non-test `.go` file (do not modify production code)
- ❌ `backend/internal/app/runtime.go`
- ❌ `backend/internal/db/` (all files)
- ❌ `backend/internal/authz/` (all files — read handlers for audit only)
- ❌ `backend/internal/backup/` (all files)
- ❌ `backend/internal/reporting/` (all files — read handlers for audit only)
- ❌ `backend/internal/orchestration/` (all files — read for audit only)
- ❌ `web/` (all frontend files)
- ❌ `infra/` (all infra files)
- ❌ `Makefile`
- ❌ `go.mod` / `go.sum`

## Stop Conditions

Stop and escalate if:
- A package requires test fixtures that need production code changes (e.g. unexported functions)
- You discover a bug in production code — document it in your completion report, do not fix it
- DuckDB CGO prevents a test from running — skip that test and document why
- You need to import a new external dependency

## How To Report Completion

Update `prompts/model1-completion.md` with a new section:

```markdown
## Follow-Up: Test Coverage Blitz + Error Audit

### Task 1: Timeout Fix
- Files changed: (list)
- `go test ./internal/execution/...` — PASS/FAIL

### Task 2: New Test Files
- (list each new test file with number of test functions)
- Total new test functions added: X
- `go test ./...` — PASS/FAIL

### Task 3: Error Audit
- `prompts/api-error-audit.md` created: YES/NO
- Handlers audited: X
- Inconsistencies found: X
- Critical issues: (list)

### Bugs Discovered (Do Not Fix)
- (list any production bugs found during testing)
```
