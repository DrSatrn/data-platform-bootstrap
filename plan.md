# Active Build Plan

## Current Goal

Ship a credible self-hosted v1 of the Data Platform.

## Active Workstreams

### Model 1

- add missing backend test coverage on critical paths
- harden error handling and integration seams
- resolve the `ingestion/` stub story

### Model 2

- add real URL routing
- break down large frontend page files
- add loading/error states and page-level tests

### Model 3

- keep coordination and operator docs current
- maintain the release-readiness runbooks and handoff docs
- keep CI minimal, credible, and green
- record UAT and release-gate evidence clearly

## Immediate Priorities

1. remove stale coordination history from `guide-wire.md`
2. align `plan.md` and `codex.md` to the current verified release state
3. keep deployment, backup, and release docs synchronized with the codebase
4. preserve a clean reviewer handoff for the release tag

## v1 Release Criteria

Use `v1-review-coordination-plan.md` as the source of truth.

The release is credible when:

- the required v1 items in Section 3 are satisfied
- the highest-priority gaps in Section 4 are either fixed or consciously
  deferred
- the ownership matrix in Section 6 is respected
- the top tasks in Section 8 have either landed or been explicitly deferred
- UAT from `uat-checklist.md` is executed and annotated

## Current Risks

- reviewer-controlled chokepoints (`runtime.go`, `App.tsx`, Compose, Makefile`)
- DuckDB CGO complexity in CI
- residual product depth gaps beyond the first finance slice
- release discipline drifting away from the actual repo state

## Next Release-Manager Actions

1. keep docs synchronized with actual repo state
2. keep CI green and minimal
3. maintain deployment and release runbooks as first-class operator docs
4. leave a clean completion report for the merge pass
