# Opsview UI Bridge

This note connects the new backend `opsview` read-model package to the staged
management-console frontend modules.

## Why This Exists

Model 3 now provides a backend-only, pure read-model layer under:

- `backend/internal/opsview/`

The frontend staging area needs a clean way to consume that shape later without
re-deriving operator signals from raw events and artifacts inside routed pages.

## Additive Frontend Bridge

New unwired files in this thread:

- `web/src/features/management/opsview/opsviewBridge.ts`
- `web/src/features/management/opsview/opsviewBridge.test.ts`
- `web/src/features/management/opsview/mockOpsview.ts`
- `web/src/features/management/opsview/OpsviewSummaryPanel.tsx`
- `web/src/features/management/opsview/OpsviewSummaryPanel.test.tsx`

## Intent

The backend `opsview` package should stay pure and backend-owned.

The frontend bridge should:

- accept backend-style summary payloads
- derive card-level operator signals
- preserve exact evidence paths for future links
- stay separate from routed page wiring

## Eventual Wiring

When Model 1 does the final integration pass, the recommended path is:

1. expose backend `opsview` summaries through an API or existing handler seam
2. map the payload through the frontend bridge
3. feed the staged console panels rather than re-inventing summary logic in the page layer

This keeps the backend and frontend responsibilities clear while reducing
duplicate grouping logic.
