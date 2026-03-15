# Model 2 Prompt — Frontend Consolidation & Polish

You are working on the `data-platform` repository. You are Model 2 of 3 parallel implementation models. Your domain is the frontend. You do NOT touch backend Go code.

## Read First

Read these files before doing anything:
- `v1-review-coordination-plan.md` (your task contracts are in Section 8)
- `codex.md` (project context and architectural rules)
- `web/src/app/App.tsx` (current routing — you will propose changes to this)
- `web/src/pages/DashboardPage.tsx` (26KB monolith — your extraction target)
- `web/src/features/dashboard/useDashboardData.ts` (data layer for dashboards)
- `web/src/features/auth/useAuth.tsx` (auth hook)
- `web/src/components/` (currently only 2 files — you will expand this)
- `web/package.json` (current dependencies)

## Your Mission

Make the frontend production-worthy for a v1 release. The frontend currently works but is unpolished: pages are monolithic, there is no URL-based routing, loading/error states are inconsistent, and core pages have zero test coverage. Your job is to fix all of this without breaking existing functionality.

## Exact Tasks (In Order)

### Task 1: Add URL-Based Routing
Replace the `useState<Route>` navigation in `App.tsx` with `react-router-dom`.

Steps:
1. Install `react-router-dom`: `cd web && npm install react-router-dom`
2. Update `web/src/main.tsx` to wrap the app in `BrowserRouter`
3. Update `web/src/app/App.tsx` to use `<Routes>` and `<Route>` components
4. Map all 6 existing routes to URL paths:
   - `/` or `/dashboard` → DashboardPage
   - `/management` → ManagementPage
   - `/metrics` → MetricsPage
   - `/pipelines` → PipelinesPage
   - `/datasets` → DatasetsPage
   - `/system` → SystemPage
5. Update the sidebar nav to use `<NavLink>` instead of `onClick` + `useState`
6. Ensure browser back/forward and direct URL access work

**Completion signal:** `npm run build` passes. Navigating to `http://localhost:3000/pipelines` directly loads the Pipelines page.

### Task 2: Extract Dashboard Widget Components
Break `DashboardPage.tsx` (26KB) into smaller, testable components.

Create these files under `web/src/components/dashboard/`:
- `WidgetRenderer.tsx` — renders a single widget based on its type (KPI, table, line, bar)
- `WidgetEditor.tsx` — the widget configuration/edit form
- `DashboardToolbar.tsx` — the dashboard action bar (create, duplicate, delete, save)
- `FilterPanel.tsx` — dashboard-level and widget-level filter controls
- `DashboardGrid.tsx` — the layout grid that positions and sizes widgets

Target: `DashboardPage.tsx` should drop to under 10KB after extraction.

**Completion signal:** `npm run build` passes. Dashboard page renders and behaves identically to before.

### Task 3: Add Loading and Error States
Ensure all data-fetching pages show proper loading and error states.

1. Create `web/src/components/LoadingSpinner.tsx`
2. Create `web/src/components/ErrorMessage.tsx`
3. Create `web/src/components/ErrorBoundary.tsx` (React error boundary)
4. Wrap each page's data-fetching logic to show `LoadingSpinner` during fetch and `ErrorMessage` on failure
5. Wrap the main content area in `App.tsx` with `ErrorBoundary` so a single page crash doesn't take down the entire app

**Completion signal:** `npm run build` passes. Temporarily breaking the API URL shows an error message instead of a blank page.

### Task 4: Add Core Page Tests
Create test files for the 4 untested core pages:

- `web/src/pages/DashboardPage.test.tsx` — tests render, loading state, error state
- `web/src/pages/PipelinesPage.test.tsx` — tests render, loading state, error state
- `web/src/pages/DatasetsPage.test.tsx` — tests render, loading state, error state
- `web/src/pages/MetricsPage.test.tsx` — tests render, loading state, error state

Each test file should have at least 3 test cases:
1. Renders without crashing
2. Shows loading state initially
3. Shows error message on API failure (mock fetch)

**Completion signal:** `npm test` passes with all new tests green.

## Allowed Files

You may create or edit:
- `web/src/app/App.tsx` (routing changes — this is an exception to reviewer-only, granted for this task)
- `web/src/main.tsx` (BrowserRouter wrapper)
- `web/src/pages/DashboardPage.tsx` (component extraction)
- `web/src/pages/PipelinesPage.tsx` (loading/error states)
- `web/src/pages/DatasetsPage.tsx` (loading/error states)
- `web/src/pages/MetricsPage.tsx` (loading/error states)
- `web/src/pages/SystemPage.tsx` (loading/error states)
- `web/src/pages/*.test.tsx` (NEW test files)
- `web/src/components/` (NEW component files)
- `web/src/features/auth/useAuth.tsx` (loading/error improvements)
- `web/src/features/dashboard/useDashboardData.ts` (only if needed for extraction)
- `web/src/features/pipelines/` (data hooks)
- `web/src/features/datasets/` (data hooks)
- `web/src/features/metrics/` (data hooks)
- `web/src/features/system/` (data hooks)
- `web/src/styles/` (styling)
- `web/src/lib/` (utilities)
- `web/package.json` (add react-router-dom)
- `web/vite.config.ts` (only if needed for routing)
- `web/server.mjs` (only if needed for client-side routing fallback)

## Forbidden Files

Do NOT edit these under any circumstances:
- ❌ `web/src/pages/ManagementPage.tsx` (Model 3's domain)
- ❌ `web/src/features/management/` (entire directory — frozen for v1)
- ❌ `backend/` (all backend files)
- ❌ `infra/` (all infra files)
- ❌ `Makefile`
- ❌ `*.md` files in repo root (documentation is Model 3's domain)
- ❌ `packages/` (all content packages)

## Inputs

- The existing frontend codebase
- The backend API contract (read backend API handler files to understand request/response shapes, but do not edit them)

## Expected Outputs

1. URL-based routing with `react-router-dom` — 6 pages at 6 paths
2. At least 5 extracted components from DashboardPage in `web/src/components/dashboard/`
3. `LoadingSpinner.tsx`, `ErrorMessage.tsx`, `ErrorBoundary.tsx` in `web/src/components/`
4. 4 new page test files with at least 3 tests each
5. A completion note (see below)

## Stop Conditions

Stop and escalate to the reviewer if:
- You find you need to change `ManagementPage.tsx` or anything in `features/management/`
- Adding routing requires changes to `server.mjs` that could break the production Compose build
- You discover a backend API behavior that prevents proper frontend error handling
- You need to add more than 2 new npm dependencies beyond `react-router-dom`
- Extracting DashboardPage requires changing the dashboard data contract (`useDashboardData.ts`) in a way that breaks the existing API integration

## How To Report Completion

When all tasks are done, create a file `prompts/model2-completion.md` containing:

```markdown
# Model 2 Completion Report

## Files Changed
- (list every file created, modified, or deleted)

## Verification
- `cd web && npm run build` — PASS/FAIL
- `cd web && npm test` — PASS/FAIL
- URL routing works: (YES/NO, tested paths)

## What Was NOT Changed
- (list forbidden files you intentionally did not touch)

## Escalation Items
- (any bugs found, constraints discovered, or help needed)

## Visual Changes
- (describe any visible UI differences so reviewer can check)
```
