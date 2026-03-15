# Comprehensive UAT (User Acceptance Testing) Checklist

This document is your step-by-step interactive manual to test the Data Platform locally and verify all its subsystems are working as intended. Ensure the platform is running via `make bootstrap` before starting.

## Test Configuration
- **Base URL:** `http://127.0.0.1:3000`
- **Admin Token:** `local-dev-admin-token`

---

## 1. Authentication & RBAC Verification
**Goal:** Verify that roles successfully restrict capabilities.

- [ ] **Step 1:** Open `http://127.0.0.1:3000` in a private/incognito window.
- [ ] **Step 2:** Observe that the UI is in an unauthenticated or "anonymous" state. Attempt to interact with any page; you should not see pipeline data or run statuses.
- [ ] **Step 3:** Enter `viewer-token` into the authentication prompt. 
- [ ] **Step 4:** Navigate to **Pipelines**. Ensure the "Run Pipeline" button is hidden or disabled (Viewers cannot trigger pipelines).
- [ ] **Step 5:** Log out or enter `local-dev-admin-token` to elevate to Admin for the remainder of the tests.
- [ ] **Step 6:** Navigate to **System** -> **Admin Terminal**. Verify that as an Admin, you can access and run commands.

---

## 2. Pipeline Execution & Worker Verification
**Goal:** Verify the scheduler, worker, and orchestration engine can successfully ingest and transform data.

- [ ] **Step 1:** Navigate to the **Pipelines** page.
- [ ] **Step 2:** Click the "Run Pipeline" (or equivalent manual trigger) button for `personal_finance_pipeline`.
- [ ] **Step 3:** Click into the actively running pipeline. 
- [ ] **Step 4:** Watch the job statuses. You should see it cleanly transition through `pending` -> `running` -> `succeeded`. 
- [ ] **Step 5:** Verify that the run artifacts (logs, inputs/outputs) are generated without error. *Background check: the worker processed raw Python jobs and materialized DuckDB SQL marts.*

---

## 3. Metadata Generation & Data Quality
**Goal:** Verify the platform correctly identifies, catalogs, and profiles data schemas that the worker just generated.

- [ ] **Step 1:** Navigate to the **Datasets** (Catalog) page.
- [ ] **Step 2:** Search for `mart_category_spend`. Click into it.
- [ ] **Step 3:** Inspect the **Schema**. Verify that column names like `month`, `category`, and `actual_spend` are documented.
- [ ] **Step 4:** Look for the **Dataset Profile**. Ensure the table contains rows (Row count > 0) and that there are computed ranges or sample values. *This confirms the Python profile task worked.*
- [ ] **Step 5:** Check the **Lineage** graph on the asset. It should visually identify upstream raw tables or intermediate rollups that fed into `mart_category_spend`.
- [ ] **Step 6:** (If applicable in the UI) Check the Data Quality indicator. It should show a passing status as defined by the quality manifests.

---

## 4. Analytical Reporting & Dashboards
**Goal:** Verify the dynamic reporting engine effectively queries the local DuckDB instance.

- [ ] **Step 1:** Navigate to **Reports** or **Dashboards**.
- [ ] **Step 2:** Open the default **Finance Overview** dashboard. Verify that charts render (e.g., Monthly Cashflow Table, Savings Rate KPI).
- [ ] **Step 3:** Create a **New Dashboard**. Give it a test name and description.
- [ ] **Step 4:** Add a **Widget**. Choose "Bar Chart", and bind it to the Dataset Reference `mart_budget_vs_actual`. 
- [ ] **Step 5:** Set the X-axis to `category` and the Y-axis (Value Field) to `variance_amount`. Save the widget.
- [ ] **Step 6:** Apply a dashboard-level **Filter** (e.g., `From Month` = `2026-01`, `To Month` = `2026-03`). Watch the charts redraw instantly.
- [ ] **Step 7:** **Save** the Dashboard. Reload the page to ensure your newly created configuration persists natively.

---

## 5. First-Party Observability & Audit Trails
**Goal:** Verify that the platform owns its own system health signals and governance.

- [ ] **Step 1:** Navigate to the **System** page.
- [ ] **Step 2:** Review the "Source of Truth" metrics. Confirm "PostgreSQL" is the primary source of truth for Runs and Dashboards.
- [ ] **Step 3:** Review the **Audit Log** feed. You should distinctly see tracked events for:
  - Your manual trigger of the pipeline.
  - The creation and saving of your test dashboard.
- [ ] **Step 4:** Open the **Admin Terminal** inside the System page.
- [ ] **Step 5:** Type `benchmark` and hit Enter. Verify it executes and reports a low-latency JSON response of the system health.
- [ ] **Step 6:** Type `backup create` and hit Enter.
- [ ] **Step 7:** Wait for the confirmation. This proves the system can accurately tarball its Postgres records, Duckdb binaries, and YAML manifests together safely.

---

## Failure Testing (Negative Tests)
Don't just test the happy path! Try breaking things safely:
- [ ] Stop the worker container (`docker stop platform-worker-1`). Trigger a pipeline and verify it gets stuck in `queued`. Start the worker and verify it picks it back up.
- [ ] Insert a bogus filter into your dashboard widget (e.g. asking for month `9999-99`). Ensure the widget gracefully shows "No Data" rather than crashing the UI.
