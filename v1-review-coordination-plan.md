# Principal Engineer Review & 3-Model Coordination Plan

**Review date:** March 15, 2026  
**Reviewer posture:** Hyper-critical principal engineer, release manager, architecture reviewer, workstream coordinator.  
**Review basis:** Full repo inspection — code, config, docs, tests, infra. Where docs and code disagree, code is trusted.

---

## 1. Executive Summary

This repo is further along than most self-hosted platform projects ever get. The backend is real, the frontend is real, the orchestration loop works, the analytical layer produces outputs, and the Compose deployment is functional. That is genuinely impressive.

However, this project is **not v1-ready today**. The distance to a credible self-hosted v1 is closer than it looks on the surface, but the gaps are concentrated in areas that directly affect whether a real user can clone, run, use, trust, and maintain the platform. The biggest risks are not missing features — they are **incomplete wiring, untested user journeys, and a frontend that needs consolidation**.

**Estimated completion toward a "real self-hostable v1":** ~70%

The remaining 30% is not 30% of code. It's the hardest 30% — the part where everything has to actually work together, look right, handle errors gracefully, and survive contact with a real user who isn't the developer.

---

## 2. Project Reality Check

### What Is Actually Implemented (Evidence-Based)

| Area | Status | Evidence |
|------|--------|----------|
| Go API server with HTTP routing | ✅ Implemented | `runtime.go` imports 18 internal packages, wires handlers |
| Worker execution loop | ✅ Implemented | `execution/runner.go` (623 lines of real dispatch logic) |
| Scheduler with cron evaluation | ✅ Implemented | `scheduler/service.go` (329 lines, custom cron parser) |
| `platformctl` CLI tool | ✅ Implemented | `cmd/platformctl/main.go` with tests |
| DuckDB SQL transforms | ✅ Implemented | `transforms/engine.go`, 4 transform SQL files |
| Python subprocess execution | ✅ Implemented | `python/` package, runner with structured I/O |
| PostgreSQL migrations (5 files) | ✅ Implemented | `infra/migrations/0001-0005` |
| PostgreSQL repositories (runs, queue, dashboards, audit, metadata, identity) | ✅ Implemented | `db/` has 11 Go files with real SQL |
| Backup create + restore | ✅ Implemented | `backup/service.go` (17KB), `backup/restore.go` (20KB), tests |
| RBAC with static tokens + native identity | ✅ Implemented | `authz/service.go`, `db/identity_store.go` |
| Audit trail | ✅ Implemented | `audit/store.go`, `db/audit_store.go` |
| Observability (health, metrics, logs) | ✅ Implemented | `observability/` (6 files including tests) |
| Admin terminal | ✅ Implemented | `admin/service.go` (15KB), `admin/handler.go` |
| External tool adapters (dbt) | ✅ Implemented | `externaltools/` (5 files + tests) |
| Opsview operator summaries | ✅ Implemented | `opsview/` (handler + 4 test files) |
| Docker Compose stack (6 services) | ✅ Implemented | `docker-compose.yml` |
| Smoke test scripts | ✅ Implemented | `localhost_smoke.sh` (10KB), `compose_smoke.sh` (4KB) |
| Frontend: 6 routed pages | ✅ Implemented | `App.tsx` routes to Dashboard, Management, Metrics, Pipelines, Datasets, System |
| Frontend: Real login/logout | ✅ Implemented | `useAuth.tsx` with username/password + token override |
| Frontend: Dashboard CRUD | ✅ Implemented | `DashboardPage.tsx` (26KB), `useDashboardData.ts` (18KB) |

### What Is Partially Implemented or Fragile

| Area | Status | Concern |
|------|--------|---------|
| `backend/internal/ingestion/` | ⚠️ README-only stub | Package exists but contains zero Go source. Ingestion logic is hardcoded directly in `execution/runner.go` via `copySampleFile`. This is fine for v1 but the package boundary is misleading. |
| `backend/pkg/` | ⚠️ Empty directory | Declared in the structure but contains nothing. Not a blocker but signals an incomplete module design intention. |
| `backend/test/` | ⚠️ README placeholder only | No integration tests exist. All tests are unit tests within individual packages. |
| `web/tests/` | ⚠️ README placeholder only | Same as above. No browser-level or integration tests. |
| `web/src/components/` | ⚠️ Only 2 files | `AdminTerminal.tsx` and `StatCard.tsx`. Every other UI element is inlined into page components. This creates massive page files (DashboardPage is 26KB). |
| Frontend test coverage | ⚠️ Heavily skewed | 10 of 15 frontend test files are in `features/management/`. Core pages (Dashboard, Pipelines, Datasets, Metrics) have **zero dedicated tests**. `PageStates.test.tsx` exists but may only test basic rendering. |
| Quality service | ⚠️ Thin | `quality/service.go` is 5.5KB with one test file (1.3KB). The quality check execution is functional but the service layer is minimal. |
| Transforms engine | ⚠️ No tests | `transforms/engine.go` has no `*_test.go` file. This is the DuckDB execution layer — a critical path with zero test coverage. |
| Reporting store tests | ✅ Has tests | `reporting/store_test.go` exists, good. |

### What Is Only Planned/Documented But Not Built

| Area | Evidence |
|------|----------|
| CI/CD pipeline | No `.github/workflows/` CI config files found. `.github/` contains only a `workflows/` directory (empty or unrelated). |
| Broader domain slices beyond personal finance | codex.md explicitly states only one domain slice exists. |
| Report sharing/export | codex.md lists as a remaining gap. |
| Dimensional analytics browsing | codex.md lists as a remaining gap. |
| Team/user administration UI | codex.md lists as a remaining gap. |
| Advanced cron features (ranges, named weekdays) | codex.md acknowledges limitation. |
| Long-running soak/load testing | codex.md acknowledges limitation. |

### What Looks Fragile

1. **`runtime.go` is a god file.** At 333 lines with 18 package imports, it is the single wiring chokepoint for the entire backend. Any backend feature addition must touch this file. This is the #1 merge conflict risk.

2. **`DashboardPage.tsx` is 26KB.** This is a monolith page component containing widget rendering, drag/drop, filter logic, CRUD operations, and chart rendering. It will be the #1 frontend merge conflict risk.

3. **`useDashboardData.ts` is 18KB.** The data layer for dashboards is a single hook file. Similarly conflict-prone.

4. **File-based queue fallback.** The system still falls back to file-backed persistence when Postgres is unavailable. This dual-path creates implicit complexity in every persistence operation. For v1, this is acceptable but it means every data path has two code paths to test.

5. **No URL-based routing.** The frontend uses `useState<Route>` for navigation, not `react-router` or browser history. Users cannot bookmark pages, share URLs, or use browser back/forward. This is a material UX gap for a v1.

### Assumptions That Could Break v1

1. **Assumption: `make smoke` proves the system works.** It proves the happy path works. It does not test error recovery, concurrent users, partial failures, or the browser UI.
2. **Assumption: All 5 workstreams are complete.** The user has checked off all items in `new-thread-eng-feedback.md`, but there is no evidence of verification beyond the checkmarks. I cannot independently confirm workstreams 3-5 are fully implemented without running the system.
3. **Assumption: The frontend is production-ready.** The frontend has 6 pages but only 2 reusable components. The Dashboard page is a 26KB monolith. There are no loading states, error boundaries, or responsive breakpoints verified in tests.

---

## 3. Definition of "100% Functioning Self-Hostable v1"

### REQUIRED FOR V1

| # | Requirement | Current Status |
|---|-------------|---------------|
| 1 | Clone → configure → `make bootstrap` → working platform in under 5 minutes | ✅ Appears to work |
| 2 | `make down` cleanly stops everything | ✅ Works |
| 3 | Web UI accessible on LAN (not just loopback) with documented config change | ✅ Documented in doc.md |
| 4 | At least one complete pipeline that proves the full data lifecycle | ✅ `personal_finance_pipeline` |
| 5 | Real login/logout with persistent sessions | ✅ Native identity store exists |
| 6 | Role-based UI restrictions (viewer/editor/admin differentiation) | ✅ Implemented |
| 7 | Dashboard create/edit/delete/save | ✅ Implemented |
| 8 | Dataset catalog with schema + lineage + freshness | ✅ Implemented |
| 9 | Pipeline list with run history and manual trigger | ✅ Implemented |
| 10 | System page with health + audit log | ✅ Implemented |
| 11 | Backup create + restore via CLI | ✅ Implemented |
| 12 | All backend tests pass (`go test ./...`) | ✅ Claimed, needs verification |
| 13 | Frontend builds without errors (`npm run build`) | ✅ Claimed, needs verification |
| 14 | Smoke test passes (`make smoke`) | ✅ Claimed, needs verification |
| 15 | Error handling in UI (API errors don't crash, show user-friendly messages) | ⚠️ Uncertain |
| 16 | Loading states in UI (spinners/skeletons while data loads) | ⚠️ Uncertain |
| 17 | Docker health checks converge reliably | ✅ Compose file has health checks |
| 18 | Environment configuration is documented and correct | ✅ `.env.example` and `.env.compose.example` exist |
| 19 | First-time user can understand what the platform does within 60 seconds | ⚠️ README exists but could be clearer for a cold user |

### NICE TO HAVE (v1.0 polish)

| # | Requirement | Current Status |
|---|-------------|---------------|
| 20 | URL-based routing (bookmarkable pages, browser back/forward) | ❌ Not implemented |
| 21 | Responsive/mobile-friendly UI | ❌ Unknown, unlikely |
| 22 | Frontend error boundaries (one widget crash doesn't kill the page) | ❌ Unknown |
| 23 | CI/CD pipeline (automated tests on push) | ❌ Not implemented |
| 24 | Component library extraction (reusable UI components) | ❌ Only 2 components exist |
| 25 | User administration UI (add/remove users from browser) | ❌ Not implemented |
| 26 | Integration tests (browser-level or API-level) | ❌ Only unit tests exist |

### DEFER TO v1.1+

| # | Requirement |
|---|-------------|
| 27 | Multiple domain slices beyond personal finance |
| 28 | Report sharing/export |
| 29 | Advanced cron features |
| 30 | External identity provider integration (OIDC/SAML) |
| 31 | S3/cloud artifact storage |
| 32 | Horizontal scaling (multiple workers) |
| 33 | Long-running performance certification |
| 34 | Team/permission group management |

---

## 4. Gap Analysis

| # | Gap | Why It Matters | Severity | Blocks v1? | Owner | Dependencies |
|---|-----|---------------|----------|-----------|-------|-------------|
| G1 | **No URL routing in frontend** | Users can't bookmark, share, or use back button. Feels broken for any web app. | High | No, but severely hurts UX | Frontend | None |
| G2 | **DashboardPage.tsx is 26KB monolith** | Unmaintainable, untestable, merge-conflict magnet. | High | No | Frontend | None |
| G3 | **Only 2 reusable components** | Every page re-implements UI patterns. Inconsistent look and feel. | Medium | No | Frontend | None |
| G4 | **Zero tests for core pages** | Dashboard, Pipelines, Datasets, Metrics pages have no tests. | High | Soft yes — can't verify features work after changes. | Frontend | None |
| G5 | **No integration/e2e tests** | backend/test/ and web/tests/ are empty. Only unit tests exist. | Medium | No, smoke scripts partially compensate. | Cross-cutting | Running stack |
| G6 | **`transforms/engine.go` has no tests** | The DuckDB execution engine is a critical path with zero test coverage. | High | No, but dangerous for maintenance. | Backend | DuckDB binary |
| G7 | **`ingestion/` package is a stub** | Ingestion logic is hardcoded in `execution/runner.go`. Not modular. | Low | No, it works as-is. | Backend | None |
| G8 | **No CI/CD** | No automated verification on push. Relies on humans running `make smoke`. | Medium | No, but risky for multi-model parallel work. | Platform | GitHub |
| G9 | **frontend error handling/loading states unverified** | If API is slow or returns errors, UI behavior is uncertain. | Medium | Soft yes — bad UX on first impression. | Frontend | None |
| G10 | **No user admin UI** | Users can be created via API/CLI but not managed in the browser. | Low | No, API/CLI is sufficient for v1. | Defer | Identity API |
| G11 | **`runtime.go` is a merge chokepoint** | Every backend feature addition requires editing this file. | High | No, but blocks parallel work. | Architecture | Coordination |
| G12 | **`guide-wire.md` is stale** | Lists files from a previous session's coordination. Doesn't reflect current hot-file reality. | Medium | No, but will cause confusion for 3-model work. | Reviewer | None |
| G13 | **`plan.md` is stale** | Lists "latest completed workstep" as restore automation. Multiple sessions have passed since. | Medium | No | Reviewer | None |

---

## 5. Recommended 3-Model Work Partition

### Strategy: Vertical Domain Boundaries

The safest partitioning is by **layer** (backend, frontend, platform/infra) rather than by feature, because the codebase's merge chokepoints (`runtime.go`, `App.tsx`) make feature-based parallelism dangerous.

---

### Model 1: Backend Hardening & API Completeness

**Mission:** Make the backend bulletproof for v1. Fill test gaps, harden error handling, clean up stubs, and ensure every API endpoint returns predictable responses under all conditions.

**Allowed directories/files:**
- `backend/internal/transforms/` (add tests)
- `backend/internal/quality/` (expand tests)
- `backend/internal/ingestion/` (either implement or delete)
- `backend/internal/execution/` (error handling improvements)
- `backend/internal/storage/` (verify artifact API edge cases)
- `backend/internal/analytics/` (verify error responses)
- `backend/test/` (create integration tests)
- `backend/internal/shared/` (utilities)
- New test files (`*_test.go`) in any package

**Forbidden directories/files:**
- ❌ `backend/internal/app/runtime.go` (reviewer-only)
- ❌ `web/` (all frontend)
- ❌ `infra/compose/docker-compose.yml`
- ❌ `backend/internal/authz/` (don't change auth)
- ❌ `backend/internal/db/` (don't change schemas)

**Required inputs from other models:** None. Independent.

**Outputs:**
- Test coverage for `transforms/engine.go`
- Test coverage for `quality/service.go` (expanded)
- Decision on `ingestion/` package (implement or remove stub)
- Integration test skeleton in `backend/test/`
- Error response audit document

**Handoff points:** After completing test additions, merge first (low conflict risk).

**Merge order:** 1st (safest, additive-only)

---

### Model 2: Frontend Consolidation & Polish

**Mission:** Make the frontend production-worthy for v1. Extract components from monolith pages, add URL routing, add loading/error states, add basic page tests, and ensure visual consistency.

**Allowed directories/files:**
- `web/src/pages/*` (all page files)
- `web/src/components/` (add new components)
- `web/src/features/auth/` (loading/error states)
- `web/src/features/dashboard/` (component extraction)
- `web/src/features/pipelines/`
- `web/src/features/datasets/`
- `web/src/features/metrics/`
- `web/src/features/system/`
- `web/src/styles/`
- `web/src/lib/`
- `web/package.json` (only to add `react-router-dom`)

**Forbidden directories/files:**
- ❌ `web/src/features/management/` (Model 3's domain)
- ❌ `web/src/app/App.tsx` (reviewer-only after routing is agreed upon)
- ❌ `backend/` (all backend)
- ❌ `infra/`

**Required inputs from other models:** None initially. Needs reviewer approval for URL routing approach before changing `App.tsx`.

**Outputs:**
- URL-based routing with `react-router-dom`
- At least 5 extracted components from DashboardPage
- Page-level tests for Dashboard, Pipelines, Datasets, Metrics
- Loading and error states for all data-fetching pages
- Consistent visual styling across all pages

**Handoff points:** Must coordinate with reviewer on `App.tsx` changes. Must not touch `ManagementPage.tsx`.

**Merge order:** 2nd (after backend tests, before infra)

---

### Model 3: Platform, Infra, Docs & Release Readiness

**Mission:** Prepare the project for a credible v1 release. Update stale coordination docs, create CI/CD scaffolding, verify the full boot-to-use-to-teardown lifecycle, write the release checklist, and fix any deployment/config issues.

**Allowed directories/files:**
- `infra/scripts/` (add/improve scripts)
- `infra/compose/docker-compose.yml` (only with reviewer approval for bind address changes)
- `.github/workflows/` (create CI config)
- `docs/` (all documentation)
- `guide-wire.md` (update for current state)
- `plan.md` (update for current state)
- `codex.md` (update for current state)
- `README.md`
- `doc.md`
- `contributing.md`
- `infra-overview.md`
- `uat-checklist.md`
- `new-thread-eng-feedback.md`
- `web/src/features/management/` (maintenance only, no structural changes)

**Forbidden directories/files:**
- ❌ `backend/internal/` (all backend packages)
- ❌ `web/src/pages/` (Model 2's domain)
- ❌ `web/src/components/` (Model 2's domain)
- ❌ `web/src/features/auth/`, `web/src/features/dashboard/` (Model 2's domain)

**Required inputs from other models:** Status updates from Model 1 and Model 2 to update docs accurately.

**Outputs:**
- Updated `guide-wire.md` reflecting current hot-file reality
- Updated `plan.md` reflecting current state
- GitHub Actions CI workflow (test + build + smoke)
- Verified UAT walkthrough (actually run it and document results)
- Release checklist with pass/fail evidence

**Handoff points:** Must get status from Model 1 (backend test results) and Model 2 (frontend state) before finalizing docs.

**Merge order:** 3rd (last, after both code changes are stable)

---

## 6. File Ownership Matrix

### Backend

| Directory/File | Owner | Classification |
|---------------|-------|---------------|
| `backend/internal/app/runtime.go` | **Reviewer-only** | 🔒 Integration chokepoint. No model edits without explicit approval. |
| `backend/internal/transforms/` | Model 1 | Open for test additions |
| `backend/internal/quality/` | Model 1 | Open for test additions |
| `backend/internal/ingestion/` | Model 1 | Decide: implement or delete |
| `backend/internal/execution/` | Model 1 | Error handling + tests |
| `backend/internal/storage/` | Model 1 | Edge case tests |
| `backend/internal/analytics/` | Model 1 | Error response tests |
| `backend/test/` | Model 1 | Integration test scaffolding |
| `backend/internal/authz/` | **No-fly zone** | Do not touch for v1 |
| `backend/internal/db/` | **No-fly zone** | Do not touch for v1 |
| `backend/internal/backup/` | **No-fly zone** | Working, don't touch |
| `backend/internal/orchestration/` | **No-fly zone** | Working, don't touch |
| `backend/internal/scheduler/` | **No-fly zone** | Working, don't touch |
| `backend/internal/reporting/` | **No-fly zone** | Working, don't touch |
| `backend/internal/audit/` | **No-fly zone** | Working, don't touch |
| All other backend packages | **No-fly zone** | Working, don't touch unless specifically needed |

### Frontend

| Directory/File | Owner | Classification |
|---------------|-------|---------------|
| `web/src/app/App.tsx` | **Reviewer-only** | 🔒 Sequential merge only. Model 2 proposes, reviewer approves. |
| `web/src/pages/DashboardPage.tsx` | Model 2 | Component extraction target |
| `web/src/pages/PipelinesPage.tsx` | Model 2 | Polish + tests |
| `web/src/pages/DatasetsPage.tsx` | Model 2 | Polish + tests |
| `web/src/pages/MetricsPage.tsx` | Model 2 | Polish + tests |
| `web/src/pages/SystemPage.tsx` | Model 2 | Polish + tests |
| `web/src/pages/ManagementPage.tsx` | Model 3 | Maintenance only |
| `web/src/components/` | Model 2 | Add new components |
| `web/src/features/management/` | Model 3 | Maintenance only |
| `web/src/features/auth/` | Model 2 | Loading/error states |
| `web/src/features/dashboard/` | Model 2 | Component extraction |
| `web/src/styles/` | Model 2 | Styling |
| `web/src/lib/` | Model 2 | Utilities |

### Infra & Docs

| Directory/File | Owner | Classification |
|---------------|-------|---------------|
| `infra/compose/docker-compose.yml` | **Reviewer-only** | 🔒 No unauthorized changes |
| `infra/scripts/` | Model 3 | Open |
| `infra/migrations/` | **No-fly zone** | Working, don't touch |
| `.github/` | Model 3 | CI setup |
| `docs/` | Model 3 | All docs |
| `guide-wire.md` | Model 3 | Update |
| `plan.md` | Model 3 | Update |
| `codex.md` | Model 3 | Update |
| `README.md` | Model 3 | Update |
| `Makefile` | **Reviewer-only** | 🔒 |

---

## 7. Merge and Coordination Protocol

### Task Assignment
- Tasks assigned by reviewer via updated task contracts below.
- Each model works only within its declared boundaries.
- No model starts a task that touches a reviewer-only or no-fly-zone file.

### File Ownership Enforcement
- Before starting any edit, check the ownership matrix.
- If you need to touch a file you don't own, **stop and escalate** to the reviewer.
- "I need to import a new package in `runtime.go`" = escalation trigger.

### Architecture Change Proposals
- Any change that modifies interfaces, adds new packages, or changes data flow must be proposed as a markdown document, not committed directly.
- Reviewer approves before implementation.

### Shared Types/Contracts
- If Model 1 adds a new backend API endpoint that Model 2 needs, Model 1 documents the endpoint contract (URL, method, request/response shapes) in a handoff note.
- Model 2 does **not** read backend code directly — it reads the contract.

### Required Update Format
After completing each task, each model must leave a note containing:
1. Files changed (list)
2. What is now verifiably true
3. Verification commands run and their results
4. What was explicitly NOT changed
5. Any escalation items for the reviewer

### Merge Order
1. **Model 1 merges first** (backend test additions are additive, lowest conflict risk)
2. **Model 2 merges second** (frontend changes are isolated but may need rebasing on App.tsx)
3. **Model 3 merges last** (docs update to reflect the final state after code changes)

### Collision Resolution
If two models discover they need the same file:
1. Stop immediately.
2. Document the conflict in a markdown note.
3. Reviewer decides ownership or splits the edit.
4. Never resolve it by "hoping the merge works."

---

## 8. Top 10 Next Tasks

### Task 1: Add Tests for `transforms/engine.go`

| Field | Value |
|-------|-------|
| **Objective** | Create `transforms/engine_test.go` covering `MaterializeRawTables`, `RunTransform`, `QueryRows`, `QueryRowsFromFile`, and `RunMetric`. |
| **Owner** | Model 1 |
| **Files allowed** | `backend/internal/transforms/engine_test.go` (NEW) |
| **Files forbidden** | `backend/internal/transforms/engine.go` (read-only) |
| **Deliverable** | Test file with at least 5 test functions covering happy path and error cases |
| **Completion signal** | `cd backend && go test ./internal/transforms/...` passes |
| **Escalation trigger** | If DuckDB requires CGO and test can't run in isolation, document the constraint |
| **Dependencies** | None |
| **Risk** | Medium — DuckDB CGO dependency may require test fixtures |
| **Can parallelize** | Yes |

### Task 2: Add URL-Based Routing to Frontend

| Field | Value |
|-------|-------|
| **Objective** | Replace `useState<Route>` navigation with `react-router-dom` so pages have real URLs (`/dashboard`, `/pipelines`, etc.) |
| **Owner** | Model 2 |
| **Files allowed** | `web/package.json`, `web/src/main.tsx`, `web/src/app/App.tsx` (with reviewer approval) |
| **Files forbidden** | `web/src/features/management/` |
| **Deliverable** | All 6 pages accessible via URL paths, browser back/forward works |
| **Completion signal** | `npm run build` passes, URL bar reflects active page |
| **Escalation trigger** | If adding react-router requires server.mjs changes |
| **Dependencies** | Reviewer approval for App.tsx structural change |
| **Risk** | Low — well-understood library |
| **Can parallelize** | Yes (after approval) |

### Task 3: Update `guide-wire.md` and `plan.md` for Current State

| Field | Value |
|-------|-------|
| **Objective** | Rewrite both coordination files to reflect the actual current hot-file ownership, completed work, and remaining gaps. Remove stale tranche status sections from guide-wire.md. |
| **Owner** | Model 3 |
| **Files allowed** | `guide-wire.md`, `plan.md` |
| **Files forbidden** | All code files |
| **Deliverable** | Both files accurately reflect repo state as of today |
| **Completion signal** | Files are internally consistent and match the file ownership matrix in this review |
| **Escalation trigger** | None |
| **Dependencies** | None |
| **Risk** | None |
| **Can parallelize** | Yes |

### Task 4: Extract Dashboard Widget Components

| Field | Value |
|-------|-------|
| **Objective** | Break `DashboardPage.tsx` (26KB) into at least 5 extracted components: `WidgetRenderer`, `WidgetEditor`, `DashboardToolbar`, `FilterPanel`, `DashboardGrid`. |
| **Owner** | Model 2 |
| **Files allowed** | `web/src/pages/DashboardPage.tsx`, `web/src/components/` (new files) |
| **Files forbidden** | `web/src/features/management/`, backend |
| **Deliverable** | DashboardPage.tsx under 8KB, extracted components in `web/src/components/dashboard/` |
| **Completion signal** | `npm run build` passes, Dashboard page renders identically |
| **Escalation trigger** | If extraction requires changing data hooks in `useDashboardData.ts` |
| **Dependencies** | None |
| **Risk** | Medium — refactoring risk |
| **Can parallelize** | Yes, but after Task 2 |

### Task 5: Add Page-Level Tests for Core Pages

| Field | Value |
|-------|-------|
| **Objective** | Create test files for `DashboardPage`, `PipelinesPage`, `DatasetsPage`, `MetricsPage` testing basic render, loading states, and error states. |
| **Owner** | Model 2 |
| **Files allowed** | `web/src/pages/*.test.tsx` (new files) |
| **Files forbidden** | Backend, management features |
| **Deliverable** | 4 new test files, each with at least 3 test cases |
| **Completion signal** | `npm test` passes |
| **Escalation trigger** | None |
| **Dependencies** | After Task 4 (component extraction) |
| **Risk** | Low |
| **Can parallelize** | After Task 4 |

### Task 6: Create GitHub Actions CI

| Field | Value |
|-------|-------|
| **Objective** | Create `.github/workflows/ci.yml` that runs backend tests, frontend build, frontend tests, and manifest validation on push/PR. |
| **Owner** | Model 3 |
| **Files allowed** | `.github/workflows/ci.yml` (NEW) |
| **Files forbidden** | All source code |
| **Deliverable** | Working CI config that can be tested locally with `act` or committed and run on GitHub |
| **Completion signal** | CI workflow file is syntactically valid |
| **Escalation trigger** | If DuckDB CGO requires special CI setup |
| **Dependencies** | None |
| **Risk** | Medium — CGO and DuckDB in CI is non-trivial |
| **Can parallelize** | Yes |

### Task 7: Expand Quality Service Tests

| Field | Value |
|-------|-------|
| **Objective** | Expand `quality/service_test.go` from 1.3KB to cover edge cases: empty results, invalid SQL paths, nil inputs. |
| **Owner** | Model 1 |
| **Files allowed** | `backend/internal/quality/service_test.go` |
| **Files forbidden** | `backend/internal/quality/service.go` (read-only) |
| **Deliverable** | At least 5 additional test functions |
| **Completion signal** | `cd backend && go test ./internal/quality/...` passes |
| **Escalation trigger** | None |
| **Dependencies** | None |
| **Risk** | Low |
| **Can parallelize** | Yes |

### Task 8: Add Frontend Loading and Error States

| Field | Value |
|-------|-------|
| **Objective** | Ensure all data-fetching pages show loading spinners during fetch and readable error messages on failure. |
| **Owner** | Model 2 |
| **Files allowed** | `web/src/pages/*.tsx`, `web/src/components/` (new ErrorBoundary, LoadingSpinner), `web/src/styles/` |
| **Files forbidden** | Backend, management |
| **Deliverable** | Visible loading/error states on all 5 core pages |
| **Completion signal** | `npm run build` passes, visual verification |
| **Escalation trigger** | None |
| **Dependencies** | None |
| **Risk** | Low |
| **Can parallelize** | Yes |

### Task 9: Resolve or Remove `ingestion/` Stub

| Field | Value |
|-------|-------|
| **Objective** | Either implement `ingestion/` as a proper package that `execution/runner.go` delegates to, OR delete the stub and update the README. |
| **Owner** | Model 1 |
| **Files allowed** | `backend/internal/ingestion/` |
| **Files forbidden** | `backend/internal/execution/runner.go` (don't refactor), `runtime.go` |
| **Deliverable** | Package either works or is deleted |
| **Completion signal** | `go test ./...` passes, no empty stub packages remain |
| **Escalation trigger** | If implementation requires changing `runner.go` materially |
| **Dependencies** | None |
| **Risk** | Low |
| **Can parallelize** | Yes |

### Task 10: Run Full UAT and Document Results

| Field | Value |
|-------|-------|
| **Objective** | Actually execute `make bootstrap`, run the full UAT checklist from `uat-checklist.md`, and document every pass/fail with screenshots or terminal output. |
| **Owner** | Model 3 (or reviewer) |
| **Files allowed** | `uat-checklist.md` (annotate results), docs |
| **Files forbidden** | All source code |
| **Deliverable** | Annotated UAT checklist with evidence |
| **Completion signal** | Every checklist item has a documented result |
| **Escalation trigger** | Any test failure indicates a code bug — escalate to Model 1 or Model 2 |
| **Dependencies** | After Model 1 and Model 2 complete their work |
| **Risk** | Medium — may surface real bugs |
| **Can parallelize** | No — this is the final verification step |

---

## 9. Anti-Bloat / Anti-Fantasy Review

### Features That Should Be Cut From v1

1. **External tool breadth (dlt, PySpark).** The dbt adapter exists and works. dlt and PySpark are "reserved contract values but explicitly gated as unimplemented." **Cut them from the v1 narrative entirely.** Don't mention them in v1 docs.

2. **Dimensional analytics browsing.** codex.md mentions "broader dimensional exploration." This is a v2+ feature. The current constrained analytics service is correct for v1.

3. **Report sharing/export.** Not needed for a self-hosted single-user platform at v1.

4. **Team/user administration UI.** API-based user management is sufficient for v1. A browser admin panel for users is v1.1.

### Complexity Not Justified Yet

1. **The `opsview` package.** This is an operator-summary aggregation layer added in the latest session. It's nice, but it adds another backend package and API endpoint that must be maintained. For v1, the raw data in the existing APIs would suffice. Keep it, but don't expand it.

2. **The `management/` frontend feature (7 subdirectories).** This is the largest frontend feature area by far, with modules for terminal sessions, follow-ups, evidence boards, runbooks, opsview bridges, and external tool inspectors. For a v1, this level of operator tooling is premature. The basic System page with the admin terminal covers the v1 need. **Freeze the management feature — no new work for v1.**

3. **Multi-store fallback pattern.** `MultiStore` (prefer Postgres, fall back to files) is conceptually clean but doubles the code paths. For v1, either commit to Postgres-only when Compose is running, or commit to files-only for local dev. Don't test two paths.

### Abstractions That Are Premature

1. **`backend/pkg/`** — empty directory implying public API packages. Delete it. Nothing in the project needs it.
2. **`packages/schemas/`** — exists but its purpose and contents are unclear. If it's not used by the runtime, it's aspirational scaffolding.

### Places Where Docs Pretend More Completeness Than Exists

1. **`new-thread-eng-feedback.md` workstreams 3-5** are checked off as complete, but there is no verification evidence beyond the checkmarks. The user marked these, not a test suite.
2. **`codex.md` says "v2-ready state for the personal-finance slice."** This is optimistic. The frontend has no page tests, a 26KB monolith page, and no URL routing. "v2-ready" implies a level of polish that doesn't exist.
3. **`guide-wire.md`** is entirely stale. It references "current additive work" from sessions ago and lists files that are no longer relevant to current coordination.

---

## 10. Final Executive Verdict

### Current State: ~70% toward a real self-hostable v1

The backend is strong (~85% ready). The frontend is functional but unpolished (~55% ready). The infra/docs are solid but stale (~65% ready).

### Top 5 Blockers

1. **No URL routing** — The frontend feels like a prototype without bookmarkable URLs.
2. **DashboardPage.tsx is a 26KB monolith** — Untestable, unmaintainable, conflict-prone.
3. **Zero tests for core frontend pages** — Can't verify features work after changes.
4. **`transforms/engine.go` has no tests** — The DuckDB layer is a critical path with zero coverage.
5. **Stale coordination docs** — `guide-wire.md` and `plan.md` will mislead the next models.

### Top 5 Biggest Risks

1. **`runtime.go` merge conflicts** — Any two models touching backend features will collide here.
2. **26KB `DashboardPage.tsx`** — Any two frontend changes will collide here.
3. **Unverified workstream completion** — Workstreams 3-5 are checked off without test evidence.
4. **CGO/DuckDB in CI** — Setting up automated testing with DuckDB's CGO requirement is non-trivial.
5. **Frontend visual quality** — No one has done a systematic visual review of the UI. It might look rough.

### Best Parallelization Strategy

**Layer-based partitioning** (backend / frontend / platform-infra) as described in Section 5. This is the only strategy that keeps the merge chokepoints (`runtime.go`, `App.tsx`) under reviewer control while letting three models make real progress without collisions.

### The Single Most Important Thing To Do Next

**Update the stale coordination docs (`guide-wire.md`, `plan.md`) and establish the file ownership matrix** — without this, any parallel work is flying blind. This costs 30 minutes and prevents days of merge conflicts.
