# Opsview Read Models

`backend/internal/opsview/` is the additive backend read-model layer for a
future management console.

Current scope:

- external tool run summaries from `orchestration.PipelineRun` events
- operator evidence summaries from `storage.Artifact` entries
- compact attention summaries across grouped external-tool jobs

Important boundary:

- this package is pure and unwired
- it does not own handlers, runtime wiring, storage, or orchestration
- it exists so future control-plane surfaces can consume stable operator-facing
  summaries without duplicating grouping logic

Primary files:

- [models.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/opsview/models.go)
- [external_tool_summary.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/opsview/external_tool_summary.go)
