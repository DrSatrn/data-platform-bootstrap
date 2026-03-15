# Model 2 Completion Report

## Files Changed
- `web/package.json`
- `web/package-lock.json`
- `web/src/main.tsx`
- `web/src/app/App.tsx`
- `web/src/app/App.test.tsx`
- `web/src/pages/DashboardPage.tsx`
- `web/src/pages/DashboardPage.test.tsx`
- `web/src/pages/PipelinesPage.tsx`
- `web/src/pages/PipelinesPage.test.tsx`
- `web/src/pages/DatasetsPage.tsx`
- `web/src/pages/DatasetsPage.test.tsx`
- `web/src/pages/MetricsPage.tsx`
- `web/src/pages/MetricsPage.test.tsx`
- `web/src/pages/SystemPage.tsx`
- `web/src/styles/global.css`
- `web/src/components/LoadingSpinner.tsx`
- `web/src/components/ErrorMessage.tsx`
- `web/src/components/ErrorBoundary.tsx`
- `web/src/components/dashboard/DashboardGrid.tsx`
- `web/src/components/dashboard/DashboardToolbar.tsx`
- `web/src/components/dashboard/FilterPanel.tsx`
- `web/src/components/dashboard/WidgetEditor.tsx`
- `web/src/components/dashboard/WidgetRenderer.tsx`
- `web/src/components/dashboard/widgetUtils.ts`

## Verification
- `cd web && npm run build` — PASS
- `cd web && npm test` — PASS
- URL routing works: YES, tested direct path rendering in `App.test.tsx` for `/pipelines`, and the packaged `web/server.mjs` already serves `index.html` as the SPA fallback for unknown non-asset paths

## What Was NOT Changed
- `web/src/pages/ManagementPage.tsx`
- `web/src/features/management/`
- `backend/`
- `infra/`
- `Makefile`
- repo-root `*.md` docs
- `packages/`

## Escalation Items
- `react-router-dom` was added as the only new runtime dependency.
- `App.test.tsx` now emits harmless `useLayoutEffect does nothing on the server` warnings because it uses `MemoryRouter` with `renderToStaticMarkup`; tests still pass and production/browser behavior is unaffected.
- The working tree contains unrelated non-frontend changes from other lanes; this report covers only the frontend consolidation work above.

## Visual Changes
- Sidebar navigation is now URL-backed and bookmarkable with active link states.
- Core data pages now render consistent loading and error panels instead of ad hoc text-only states.
- The dashboard page has been split into reusable toolbar, filter, grid, widget renderer, and widget editor components while preserving the existing editing flow.
