# Active Build Plan

## Current Goal

Close the v3 pre-production gaps identified in
`prompts/v3-review-coordinator-report.md`.

## Model 1

- finish retries and idempotency in orchestration
- land native DB ingestion support
- add retention and purge behavior

## Model 2

- deepen semantic rollups
- add CSV export
- add shareable dashboard links

## Model 3

- keep webhook alerting trustworthy
- expose stable Prometheus telemetry
- automate tagged multi-architecture release assets
- keep coordination rails aligned with the live prompt set

## Model 3 Verified This Pass

- backend webhook alerting for failed runs and stale assets
- `/api/v1/system/metrics` with Prometheus text exposition
- tag-triggered GitHub release packaging for `platformctl`
- explicit Go module coverage for the MySQL driver used by ingestion/export

## Immediate Next Actions

1. keep the new ops surfaces green while Model 1 and Model 2 finish their v3 lanes
2. avoid reintroducing stale prompt references in coordination docs
3. treat tagged-release workflow verification as the next release-manager checkpoint
