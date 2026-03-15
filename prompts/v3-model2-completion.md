# Model 2 v3 Completion

## Scope Closed

Completed the v3 Model 2 contract from `prompts/v3-model2-frontend.md`:

1. richer semantic rollups
2. CSV export for analytics widgets
3. shareable dashboard links via URL query state

## What Changed

### Backend
- Extended `backend/internal/analytics/service.go` to support safe multi-dimension rollups via curated comma-separated `group_by` fields.
- Added CSV export helpers for curated analytics results in `backend/internal/analytics/service.go`.
- Added `/api/v1/analytics/export` in `backend/internal/analytics/handler.go`.
- Registered the export endpoint in `backend/internal/app/runtime.go`.
- Persisted widget `group_by` configuration in `backend/internal/reporting/store.go`.

### Frontend
- Added URL-backed dashboard state helpers in `web/src/features/dashboard/queryState.ts`.
- Synced selected dashboard, preset, and active view filters into the browser query string in `web/src/features/dashboard/useDashboardData.ts`.
- Hydrated dashboard state from the URL on load in `web/src/features/dashboard/useDashboardData.ts`.
- Added widget CSV export wiring from the dashboard UI through:
  - `web/src/pages/DashboardPage.tsx`
  - `web/src/components/dashboard/DashboardGrid.tsx`
  - `web/src/components/dashboard/WidgetRenderer.tsx`
  - `web/src/lib/api.ts`
- Added `group_by` editing support in `web/src/components/dashboard/WidgetEditor.tsx`.
- Updated filter behavior in `web/src/components/dashboard/FilterPanel.tsx` so non-edit usage controls shareable view state instead of mutating saved defaults.

### Tests
- Added backend export handler coverage in `backend/internal/analytics/handler_test.go`.
- Added multi-dimension grouping coverage in `backend/internal/analytics/service_deeper_test.go`.
- Added dashboard query-state coverage in `web/src/features/dashboard/queryState.test.ts`.
- Extended `web/src/lib/api.test.ts` for the CSV download flow.
- Updated `web/src/pages/DashboardPage.test.tsx` for the expanded dashboard hook contract.
- Added export auth coverage in `backend/internal/app/runtime_auth_test.go`.

## What Is Now Verifiably True

- Curated analytics queries can group by multiple dimensions at once.
- Dashboard widgets can export their current curated dataset view as CSV through `/api/v1/analytics/export`.
- Dashboard filter/preset state can be shared through the URL and restored on reload.
- Widget `group_by` settings survive dashboard persistence instead of being dropped.
- Frontend tests and production build both pass with the new dashboard behavior.
- Analytics and reporting backend package tests pass with the new rollup and export functionality.

## What Remains Intentionally Unfinished

- The export endpoint currently covers curated analytics results as CSV only; PDF/export-sharing workflows are still future work.
- Historical note: the earlier `backend/internal/app` compile blocker caused by the missing `github.com/go-sql-driver/mysql` module has since been resolved in the repo, so it is no longer an active limitation for this completion note.

## Verification Run

- `cd backend && go test ./internal/analytics ./internal/reporting`
- `cd web && npm test`
- `cd web && npm run build`
- `git diff --check`

## Shared Files Touched

- `backend/internal/analytics/handler.go`
- `backend/internal/analytics/service.go`
- `backend/internal/analytics/service_deeper_test.go`
- `backend/internal/app/runtime.go`
- `backend/internal/app/runtime_auth_test.go`
- `backend/internal/reporting/store.go`
- `web/src/components/dashboard/DashboardGrid.tsx`
- `web/src/components/dashboard/FilterPanel.tsx`
- `web/src/components/dashboard/WidgetEditor.tsx`
- `web/src/components/dashboard/WidgetRenderer.tsx`
- `web/src/components/dashboard/widgetUtils.ts`
- `web/src/features/dashboard/useDashboardData.ts`
- `web/src/lib/api.ts`
- `web/src/lib/api.test.ts`
- `web/src/pages/DashboardPage.tsx`
- `web/src/pages/DashboardPage.test.tsx`

## New Files Added

- `backend/internal/analytics/handler_test.go`
- `web/src/features/dashboard/queryState.ts`
- `web/src/features/dashboard/queryState.test.ts`
