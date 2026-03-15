# Model 2 v3 — Analytics, Reporting & Product UX

**Priority:** HIGH — Closes out the final product UX and analytics value blockers.  
**Owner:** Model 2  
**Merge order:** 2nd  

---

## Mission

Drive Analytics Serving (Category 5) and the Reporting App (Category 6) to 100%. Empower operators to break out of the closed UI by offering data export, richer semantic rollups, and shareable dashboard links.

## Tasks (In Priority Order)

### Task 1: Richer Semantic Rollups
- **Current state:** The analytics service enforces strict queries based on hardcoded structures to prevent chaos.
- **Required change:** Refactor `backend/internal/analytics/service.go` to safely accept dynamic multi-dimension `GROUP BY` rollups instead of single dimensions. Use AST validation if necessary to ensure queries remain within the bounds of the curated mart.
- **Completion signal:** Custom drill-downs on the frontend successfully group by multiple columns simultaneously.

### Task 2: CSV Data Export
- **Current state:** Widget data is trapped in the browser.
- **Required change:** Add a backend handler mapped to `/api/v1/analytics/export` that returns curated dataset queries directly as text/csv format. Add a corresponding "Export CSV" button to frontend chart widgets.
- **Completion signal:** A browser download flow successfully saves a `.csv` file mimicking the widget's current filters.

### Task 3: Shareable Dashboard Links
- **Current state:** Filter state is purely local to the React session.
- **Required change:** Sync dashboard filter parameters (presets, date ranges, categories) into the browser's URL query string. Ensure the page hydrates its initial state from the URL when loading.
- **Completion signal:** Copying and pasting a modified dashboard URL instantly restores the exact same filter views for another authenticated user.

## Escalation Triggers
- If dynamic semantic rollups introduce vulnerability to SQL injection, halt and escalate back to a safer restricted subset. 
