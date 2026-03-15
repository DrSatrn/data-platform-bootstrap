# UAT Results — 2026-03-15

This file records one executed UAT pass against the local Compose stack. It is
an evidence artifact, not the reusable checklist itself.

## Summary

- `make bootstrap`: PASS
- Total numbered checklist sections exercised: 12
- Passed: 10
- Failed: 2

Key failures:

- data quality was not fully passing because
  `check_uncategorized_transactions` returned `warning`
- the checklist expectation for admin-terminal `benchmark` did not match the
  current command surface

## Section Results

### 1. Confirm The Stack Is Reachable

Result: ✅ PASS

- `GET http://127.0.0.1:3000/` returned the SPA shell successfully
- `GET http://127.0.0.1:8080/healthz` returned `200` with
  `{"environment":"development","status":"ok"}`

### 2. Create Or Use Test Accounts

Result: ✅ PASS

- created `viewer-demo`, `editor-demo`, and `admin-demo` via
  `POST /api/v1/admin/users`
- also created and used `operator-admin` for live admin-path checks
- user creation calls returned HTTP `201`

### 3. Verify Authentication And Role Boundaries

Result: ✅ PASS

Anonymous:

- `GET /api/v1/session` returned `anonymous`
- anonymous capabilities showed `view_platform: false`

Viewer:

- signed in as `viewer-demo`
- viewer-triggered `POST /api/v1/pipelines` returned `403`
- viewer-triggered `POST /api/v1/admin/terminal/execute` returned `403`

Editor:

- signed in as `editor-demo`
- editor-triggered `POST /api/v1/pipelines` returned `202`
- editor created and deleted a dashboard successfully
- editor updated metadata annotations successfully
- editor-triggered `POST /api/v1/admin/terminal/execute` returned `403`

Admin:

- signed in with an admin-capable session (`operator-admin`)
- admin terminal command `status` succeeded
- admin user-management endpoint remained accessible with admin credentials

### 4. Trigger A Real Pipeline Run

Result: ✅ PASS

- triggered `personal_finance_pipeline` as admin
- run `run_20260315T071313.236188567` progressed `queued` -> `succeeded`
- job-level statuses updated and all jobs completed successfully

### 5. Verify Run Artifacts

Result: ✅ PASS

- `GET /api/v1/artifacts?run_id=run_20260315T071313.236188567` returned raw,
  staging, intermediate, mart, metric, and quality outputs
- artifact reads returned real content for the generated files

### 6. Verify The Dataset Catalog

Result: ✅ PASS

- `GET /api/v1/catalog` returned the asset catalog successfully
- `mart_monthly_cashflow`, `mart_category_spend`, and
  `mart_budget_vs_actual` were present
- `GET /api/v1/catalog/profile?asset_id=mart_category_spend` returned
  `row_count: 5` with column-level samples and ranges
- lineage and freshness data were present in the catalog payload

### 7. Verify The Metrics Browser

Result: ✅ PASS

- `GET /metrics` returned `200`
- `GET /api/v1/metrics` returned `metrics_savings_rate` and
  `metrics_category_variance`
- metric preview data loaded and a narrow empty filter returned an empty series
  instead of an error

### 8. Verify Dashboard And Reporting Workflows

Result: ✅ PASS

- `GET /reports` returned `200`
- the default finance dashboard was present through `GET /api/v1/reports`
- created dashboard `uat_finance_variance`, saved it, and confirmed it
  persisted on a later `GET /api/v1/reports`
- filtered analytics queries for `mart_budget_vs_actual` returned live data for
  `2026-01` through `2026-03`

### 9. Verify The System Page

Result: ✅ PASS

- `GET /system` returned `200`
- `GET /api/v1/system/overview` returned scheduler lag, source-of-truth modes,
  queue summary, and failure watch data
- the overview showed `postgres` as the source of truth for runs and dashboards
- `GET /api/v1/system/audit` included the live pipeline trigger and dashboard
  save events

### 10. Verify The Admin Terminal

Result: ❌ FAIL

- `status` worked and returned structured output
- `backups` worked and returned structured output
- `benchmark` did not work in the admin terminal and returned
  `unknown command "benchmark"`

### 11. Verify Backup And Recovery Visibility

Result: ✅ PASS

- `make backup` exited `0`
- output included both `backup bundle created:` and
  `backup bundle verified:`
- bundle path produced by `make backup`:
  `/Users/streanor/Documents/Playground/data-platform/var/backups/platform-backup-20260315T071852Z.tar.gz`
- separate admin-terminal `backup create` also succeeded, and the reported
  container path existed on disk
- the admin-terminal `backups` inventory did not immediately list the newly
  created terminal bundle even though the file existed

### 12. Optional Negative Tests

Result:

- ✅ PASS: stopped the worker, triggered a new run, confirmed it stayed
  `queued`, then restarted the worker and saw the run complete successfully
- ✅ PASS: queried analytics with
  `from_month=9999-99&to_month=9999-99` and received an empty series instead
  of an error
- ⚪ NOT RUN: the optional dbt-backed external-tool pipeline was not exercised

## Overall Outcome

Result: ✅ PASS WITH CAVEATS

- the core v1 path worked end to end: auth, orchestration, artifacts, catalog,
  metrics, dashboards, system views, and backup creation all worked
- two issues remained visible in this run:
  - quality was not fully passing because one check was in `warning`
  - the checklist expectation for admin-terminal `benchmark` did not match the
    current command surface
