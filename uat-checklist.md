# User Acceptance Testing Checklist

Use this checklist after you have the platform running. It is written for a
human tester, not for an AI agent.

UAT annotation for 2026-03-15:

- `make bootstrap`: PASS
- Total numbered sections exercised: 12
- Explicit failures observed: 2
- Key failures:
  - data quality was not fully passing because `check_uncategorized_transactions` returned `warning`
  - `benchmark` is not a supported admin-terminal command in this build

Recommended starting point:

1. complete [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md)
2. choose either:
   - [bootstrap.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/bootstrap.md) for Docker Compose
   - [localhost-e2e.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/localhost-e2e.md) for host-run debugging
3. come back here when the UI is reachable

## Test Configuration

- Web UI: `http://127.0.0.1:3000`
- API health: `http://127.0.0.1:8080/healthz`
- Bootstrap admin token: `local-dev-admin-token`

Important note:

- the preferred day-to-day path is native user login through `/api/v1/session`
- the bootstrap admin token exists for first-run setup and recovery
- if PostgreSQL is unavailable, native users are unavailable too and you may
  need to use the bootstrap token path

## 1. Confirm The Stack Is Reachable

UAT result: ✅ PASS

- `GET http://127.0.0.1:3000/` returned the SPA shell successfully
- `GET http://127.0.0.1:8080/healthz` returned `200` with `{"environment":"development","status":"ok"}`

Open the web UI:

- `http://127.0.0.1:3000`

Check API health:

```sh
curl http://127.0.0.1:8080/healthz
```

What success looks like:

- the browser loads the app shell
- the health endpoint returns HTTP `200`

If something goes wrong:

1. if you used Docker Compose, follow the troubleshooting section in
   [bootstrap.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/bootstrap.md)
2. if you used host-run services, follow the troubleshooting section in
   [localhost-e2e.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/localhost-e2e.md)

## 2. Create Or Use Test Accounts

UAT result: ✅ PASS

- created `viewer-demo`, `editor-demo`, and `admin-demo` via `POST /api/v1/admin/users`
- an equivalent admin account `operator-admin` was also created and used for the live admin-path checks
- user creation calls returned HTTP `201`

If you already have native users, use them. If not, create three test users
with the bootstrap admin token.

Create a viewer:

```sh
curl -X POST \
  -H "Authorization: Bearer local-dev-admin-token" \
  -H "Content-Type: application/json" \
  -d '{"username":"viewer-demo","display_name":"Viewer Demo","role":"viewer","password":"viewer-password"}' \
  http://127.0.0.1:8080/api/v1/admin/users
```

Create an editor:

```sh
curl -X POST \
  -H "Authorization: Bearer local-dev-admin-token" \
  -H "Content-Type: application/json" \
  -d '{"username":"editor-demo","display_name":"Editor Demo","role":"editor","password":"editor-password"}' \
  http://127.0.0.1:8080/api/v1/admin/users
```

Create an admin:

```sh
curl -X POST \
  -H "Authorization: Bearer local-dev-admin-token" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin-demo","display_name":"Admin Demo","role":"admin","password":"admin-password"}' \
  http://127.0.0.1:8080/api/v1/admin/users
```

What success looks like:

- each command returns HTTP `201` or success JSON for the new user

If something goes wrong:

1. if the request fails with an auth error, confirm you used the bootstrap
   admin token
2. if user creation fails because PostgreSQL is unavailable, switch to the
   bootstrap token path for the rest of this checklist

## 3. Verify Authentication And Role Boundaries

### Anonymous

UAT result: ✅ PASS

- `GET /api/v1/session` returned `anonymous`
- anonymous capabilities showed `view_platform: false`

Open the UI in a fresh browser session without signing in.

What success looks like:

- the session display shows `anonymous`
- privileged actions are unavailable
- read-heavy pages should not expose more than the current auth model allows

If something goes wrong:

- compare the observed behavior with
  [access-matrix.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/access-matrix.md)

### Viewer

UAT result: ✅ PASS

- signed in as `viewer-demo`
- viewer session had `view_platform: true` and no privileged capabilities
- viewer-triggered `POST /api/v1/pipelines` returned `403`
- viewer-triggered `POST /api/v1/admin/terminal/execute` returned `403`

Sign in as `viewer-demo`.

What success looks like:

- the UI shows `viewer`
- you can open Dashboard, Pipelines, Datasets, Metrics, and System
- you cannot trigger a pipeline
- you cannot save or delete dashboards
- you cannot use the admin terminal

If something goes wrong:

1. confirm the session actually changed to `viewer`
2. confirm the UI is showing disabled or hidden privileged actions rather than
   crashing

### Editor

UAT result: ✅ PASS

- signed in as `editor-demo`
- editor-triggered `POST /api/v1/pipelines` returned `202`
- editor created and deleted a dashboard successfully
- editor updated metadata annotations successfully
- editor-triggered `POST /api/v1/admin/terminal/execute` returned `403`

Sign out, then sign in as `editor-demo`.

What success looks like:

- the UI shows `editor`
- you can trigger a pipeline
- you can create, edit, and delete dashboards
- you can update metadata annotations where that workflow is available
- you still cannot use the admin terminal

If something goes wrong:

- verify the role in the session panel and repeat the test from a clean page load

### Admin

UAT result: ✅ PASS

- signed in with an admin-capable session (`operator-admin`)
- admin terminal command `status` succeeded
- admin user-management endpoint remained accessible with admin credentials

Sign out, then sign in as `admin-demo`.

What success looks like:

- the UI shows `admin`
- you can use the admin terminal
- you can access user-management actions on the System page

If something goes wrong:

- confirm the session token belongs to an admin user, not the editor account

## 4. Trigger A Real Pipeline Run

UAT result: ✅ PASS

- triggered `personal_finance_pipeline` as admin
- run `run_20260315T071313.236188567` progressed `queued` -> `succeeded`
- job-level statuses updated and all jobs completed successfully

Sign in as `editor-demo` or `admin-demo`.

Go to:

- `Pipelines`

Trigger:

- `personal_finance_pipeline`

What success looks like:

- the run appears quickly in the recent run list
- the status transitions from `queued` to `running` to `succeeded`
- job-level statuses update

If something goes wrong:

1. if the run stays `queued`, check whether the worker is running
2. if the run fails, inspect the latest run events and artifacts from the
   Pipelines page
3. if the button is disabled, confirm you are signed in as `editor` or `admin`

## 5. Verify Run Artifacts

UAT result: ✅ PASS

- `GET /api/v1/artifacts?run_id=run_20260315T071313.236188567` returned raw, staging, intermediate, mart, metric, and quality outputs
- artifact reads returned real content for the generated files

From the `Pipelines` page, open artifacts for the run you just triggered.

What success looks like:

- the artifact list loads
- at least some run-scoped outputs are present
- opening an artifact returns real content instead of a broken link

If something goes wrong:

1. confirm the selected run actually completed
2. confirm the worker created the artifact files
3. if the list is empty, inspect worker logs for artifact-write failures

## 6. Verify The Dataset Catalog

UAT result: ✅ PASS

- `GET /api/v1/catalog` returned the asset catalog successfully
- `mart_monthly_cashflow`, `mart_category_spend`, and `mart_budget_vs_actual` were present
- `GET /api/v1/catalog/profile?asset_id=mart_category_spend` returned `row_count: 5` with column-level samples and ranges
- lineage and freshness data were present in the catalog payload

Go to:

- `Datasets`

Inspect one or more of these assets:

- `mart_monthly_cashflow`
- `mart_category_spend`
- `mart_budget_vs_actual`

What success looks like:

- the catalog list loads
- selecting an asset shows owner, freshness, coverage, and lineage information
- the dataset profile shows row count and column-level details when available

If something goes wrong:

1. if the list is empty, confirm the pipeline run succeeded
2. if the profile fails, confirm the selected asset exists and profiling output
   was generated

## 7. Verify The Metrics Browser

UAT result: ✅ PASS

- `GET /metrics` returned `200`
- `GET /api/v1/metrics` returned `metrics_savings_rate` and `metrics_category_variance`
- metric preview data loaded and a narrow empty filter returned an empty series instead of an error

Go to:

- `Metrics`

Inspect:

- `metrics_savings_rate`
- `metrics_category_variance`

What success looks like:

- the metrics list loads
- selecting a metric shows dimensions, measures, and preview data
- changing filters updates the preview without crashing the page

If something goes wrong:

1. confirm the pipeline run succeeded
2. confirm the analytics API is healthy
3. try a known-good metric before assuming all metrics are broken

## 8. Verify Dashboard And Reporting Workflows

UAT result: ✅ PASS

- `GET /reports` returned `200`
- the default finance dashboard was present through `GET /api/v1/reports`
- created dashboard `uat_finance_variance`, saved it, and confirmed it persisted on a later `GET /api/v1/reports`
- filtered analytics queries for `mart_budget_vs_actual` returned live data for `2026-01` through `2026-03`

Sign in as `editor-demo` or `admin-demo`.

Go to:

- `Dashboard`

Test these actions:

1. open the default dashboard
2. create a new dashboard
3. add a widget
4. save the dashboard
5. refresh the page
6. confirm the dashboard still exists

What success looks like:

- existing widgets render
- new dashboards can be created and saved
- after refresh, the saved dashboard is still present

If something goes wrong:

1. if save fails, note the exact error message
2. if changes disappear after refresh, check whether the runtime is using the
   intended source of truth on the System page

## 9. Verify The System Page

UAT result: ✅ PASS

- `GET /system` returned `200`
- `GET /api/v1/system/overview` returned scheduler lag, source-of-truth modes, queue summary, and failure watch data
- the overview showed `postgres` as the source of truth for runs and dashboards
- `GET /api/v1/system/audit` included the live pipeline trigger and dashboard save events

Go to:

- `System`

Check these areas:

- Service Health
- Built-in Metrics
- Queue And Recovery
- Source Of Truth
- Failure Watch
- Audit Trail

What success looks like:

- the page loads without blank sections
- the scheduler lag and refresh time render
- the source-of-truth summary matches the runtime you started
- your earlier pipeline trigger and dashboard actions appear in the audit trail

If something goes wrong:

1. if the page fails to load, check API health first
2. if audit is empty, verify you performed privileged actions during this test

## 10. Verify The Admin Terminal

UAT result: ❌ FAIL

- `status` worked and returned structured output
- `backups` worked and returned structured output
- `benchmark` did not work in the admin terminal and returned `unknown command "benchmark"`

Sign in as `admin-demo` or use the bootstrap admin token override.

Go to:

- `System`

Use the admin terminal to run a safe command such as:

- `status`
- `backups`

What success looks like:

- the terminal accepts the command
- it returns structured output instead of a generic failure

If something goes wrong:

1. confirm you are using an admin-capable session
2. confirm the admin terminal is enabled in the current runtime

## 11. Verify Backup And Recovery Visibility

UAT result: ✅ PASS

- `make backup` exited `0`
- output included both `backup bundle created:` and `backup bundle verified:`
- bundle path produced by `make backup`: `/Users/streanor/Documents/Playground/data-platform/var/backups/platform-backup-20260315T071852Z.tar.gz`
- separate admin-terminal `backup create` also succeeded, and the reported container path existed on disk
- note: the admin-terminal `backups` inventory did not immediately list the newly created terminal bundle even though the file existed

From an admin-capable terminal outside the browser, run:

```sh
cd /Users/streanor/Documents/Playground/data-platform
make backup
```

What success looks like:

- the command exits `0`
- output includes both `backup bundle created:` and `backup bundle verified:`

If something goes wrong:

1. inspect the backup output for the first failing step
2. compare with
   [backups.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/backups.md)

## 12. Optional Negative Tests

UAT result:

- ✅ PASS: stopped the worker, triggered a new run, confirmed it stayed `queued`, then restarted the worker and saw the run complete successfully
- ✅ PASS: queried analytics with `from_month=9999-99&to_month=9999-99` and received an empty series instead of an error
- ⚪ NOT RUN: the optional dbt-backed external-tool pipeline was not exercised in this pass

Run these only if you want extra confidence:

- stop the worker and confirm a new run remains queued until the worker comes back
- use a deliberately narrow or empty filter in Metrics or Dashboard and confirm
  the UI shows an empty-state message rather than crashing
- if external-tool support is enabled in your environment, trigger a dbt-backed
  external-tool pipeline and verify logs and declared artifacts are visible

## Sign-Off

The UAT pass is successful if:

- you can sign in and observe role differences
- you can trigger and complete a real pipeline run
- artifacts, datasets, metrics, dashboards, and system views all load
- audit and admin features work for admin users
- backup creation and verification complete successfully

Observed outcome: ✅ PASS WITH CAVEATS

- the core v1 path worked end to end: auth, orchestration, artifacts, catalog, metrics, dashboards, system views, and backup creation all worked
- two issues remain visible from this UAT:
  - quality is not fully passing because one check is in `warning`
  - the checklist expectation for admin-terminal `benchmark` does not match the current command surface
