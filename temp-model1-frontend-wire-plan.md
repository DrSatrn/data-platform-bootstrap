# Model 1 Frontend Wiring Plan

This is the execution brief for the final frontend wiring and integration pass.

The repo now has enough additive building blocks from Model 2 and Model 3 that
the next step should be deliberate integration, not more scaffolding.

## Read First

- `guide-wire.md`
- `docs/product/management-console-blueprint.md`
- `docs/product/web-terminal-blueprint.md`
- `docs/product/operator-followup-blueprint.md`
- `docs/product/operator-evidence-blueprint.md`
- `docs/product/management-console-integration-map.md`
- `docs/product/opsview-ui-bridge.md`
- `docs/reference/external-tool-jobs.md`
- `docs/reference/opsview-read-models.md`

## Mission

Wire the staged management-console and web-terminal work into the real product
so the frontend is genuinely integrated, not just mock-backed.

This is explicitly the “use the staged assets rather than rebuilding them”
phase.

## Main Assets To Reuse

### Frontend staging area

- `web/src/features/management/terminal/*`
- `web/src/features/management/externalTools/*`
- `web/src/features/management/opsview/*`
- `web/src/features/management/runbooks/*`
- `web/src/features/management/evidence/*`
- `web/src/features/management/console/*`

### Backend additive helpers already available

- `backend/internal/opsview/*`
- external-tool execution, storage, handler, and visibility test coverage

## Required Outcomes

1. Replace mock-backed management-console data flows with real app wiring where
   practical.
2. Expose a real integrated management surface in the web app.
3. Wire external-tool visibility into a real operator-facing page or section.
4. Wire terminal/session/follow-up/evidence/runbook modules into a coherent
   in-app experience.
5. Use backend `opsview` summaries where they help reduce frontend grouping
   logic, or clearly bridge toward them.
6. Preserve and extend test coverage as real wiring replaces mock-only staging.

## Strong Guidance

- Prefer adapting the staged modules over rewriting them.
- Keep the terminal guided and role-aware, not arbitrary shell access.
- Keep exact artifact paths visible so operators can trust what they are
  opening.
- Keep runbook linkage visible at the moment of operator action.
- Keep external-tool runs, logs, outputs, and failure class visible in the UI.

## Likely Wiring Targets

These are the places most likely to need integration work:

- app/page/route wiring
- real data hooks in `web/src/features/**`
- any handler/API additions needed to expose opsview-style summaries cleanly
- auth-aware visibility and operator navigation

## Definition Of Done

- the frontend management experience is real, not just unwired staging
- staged management modules are integrated into the actual app
- external-tool visibility is operator-usable in the frontend
- terminal/session/follow-up/evidence/runbook UX is wired coherently
- tests pass after integration
- docs remain in sync with the integrated behavior

## Verification Bar

At minimum, re-run:

```bash
cd backend && go test ./...
cd web && npm test
cd web && npm run build
go run ./cmd/platformctl validate-manifests
sh infra/scripts/compose_smoke.sh
```

If you change benchmark-sensitive behavior, also run:

```bash
sh infra/scripts/benchmark_suite.sh
```

## Non-Goals For This Pass

- do not create another parallel staging layer
- do not leave the key management surfaces mock-only if real wiring is possible
- do not ignore the additive assets and start over from scratch

Signed,
Model 2
