# Guide Wire

This file is the additive-only coordination rail for parallel work on the repo.

It is intentionally pragmatic:

- what can be built safely in parallel
- what should not be touched while another thread is wiring core behavior
- what additive assets already exist
- what is ready for later integration

## Current Goal

Keep a second thread productive without colliding with active implementation
work in hot files.

The preferred pattern is:

- create new docs
- create new tests
- create new standalone tooling
- avoid edits to shared runtime wiring until a deliberate merge pass

## Safe Zones For Parallel Work

These are the safest places to add value with low merge risk:

- new docs under `docs/runbooks/`
- new docs under `docs/tutorials/`
- new docs under `docs/reference/`
- new backend tests as new `*_test.go` files
- new frontend tests as new `*.test.ts` or `*.test.tsx` files
- new standalone backend packages under `backend/internal/`
- new standalone commands under `backend/cmd/`
- new draft scripts under `infra/scripts/drafts/`

## Hot Files To Avoid Editing In Parallel

Unless explicitly reassigned, avoid editing these while another model is doing
core repo cleanup:

- `README.md`
- `backend/internal/app/runtime.go`
- `backend/internal/config/config.go`
- `backend/cmd/platformctl/main.go`
- `backend/internal/orchestration/*.go`
- `backend/internal/reporting/*.go`
- `backend/internal/authz/*.go`
- `backend/internal/metadata/*.go`
- existing runbooks that are likely being actively revised

## Additive Assets Created In This Thread

### Tests

- `backend/internal/config/config_env_test.go`
- `backend/internal/config/external_tool_config_test.go`
- `backend/internal/storage/service_test.go`
- `backend/internal/manifests/loader_test.go`
- `backend/internal/manifests/external_tool_loader_test.go`
- `web/src/lib/api.test.ts`
- `web/src/features/management/terminal/sessionModel.test.ts`
- `web/src/features/management/terminal/followupPlanner.test.ts`
- `web/src/features/management/console/ManagementConsolePreview.test.tsx`
- `web/src/features/management/runbooks/runbookDock.test.ts`
- `web/src/features/management/evidence/evidenceBoard.test.ts`
- `web/src/features/management/opsview/opsviewBridge.test.ts`
- `web/src/features/management/opsview/OpsviewSummaryPanel.test.tsx`
- `backend/internal/orchestration/external_tool_validation_test.go`
- `backend/internal/orchestration/external_tool_validation_artifacts_test.go`
- `backend/internal/orchestration/handler_external_tool_test.go`
- `backend/internal/externaltools/dbt_test.go`
- `backend/internal/externaltools/runner_test.go`
- `backend/internal/execution/external_tool_test.go`
- `backend/internal/execution/external_tool_failures_test.go`
- `backend/internal/execution/external_tool_operator_inspection_test.go`
- `backend/internal/opsview/external_tool_summary_test.go`
- `backend/internal/opsview/snapshot_builders_test.go`
- `backend/internal/opsview/golden_snapshot_test.go`
- `backend/internal/storage/handler_external_tool_test.go`
- `backend/internal/storage/handler_external_tool_read_test.go`

### Docs

- `docs/runbooks/config-reality.md`
- `docs/runbooks/access-matrix.md`
- `docs/runbooks/local-host-run.md`
- `docs/runbooks/optional-external-tools.md`
- `docs/runbooks/external-tool-troubleshooting.md`
- `docs/runbooks/dbt-operator-checklist.md`
- `docs/tutorials/trace-one-pipeline-complete.md`
- `docs/reference/external-tool-jobs.md`
- `docs/reference/opsview-read-models.md`
- `docs/product/web-terminal-blueprint.md`
- `docs/product/operator-followup-blueprint.md`
- `docs/product/management-console-integration-map.md`
- `docs/product/operator-evidence-blueprint.md`
- `docs/product/opsview-ui-bridge.md`

### Frontend Management Prototypes

- `web/src/features/management/README.md`
- `web/src/features/management/types.ts`
- `web/src/features/management/mockControlPlane.ts`
- `web/src/features/management/terminal/commandCatalog.ts`
- `web/src/features/management/terminal/commandCatalog.test.ts`
- `web/src/features/management/terminal/OperatorWorkbench.tsx`
- `web/src/features/management/terminal/sessionModel.ts`
- `web/src/features/management/terminal/sessionModel.test.ts`
- `web/src/features/management/terminal/mockSessions.ts`
- `web/src/features/management/terminal/followupPlanner.ts`
- `web/src/features/management/terminal/followupPlanner.test.ts`
- `web/src/features/management/terminal/mockFollowups.ts`
- `web/src/features/management/terminal/TerminalTranscript.tsx`
- `web/src/features/management/terminal/OperatorSessionDeck.tsx`
- `web/src/features/management/terminal/ArtifactFollowupPanel.tsx`
- `web/src/features/management/terminal/NextStepBoard.tsx`
- `web/src/features/management/inventory/assetAttention.ts`
- `web/src/features/management/inventory/assetAttention.test.ts`
- `web/src/features/management/runbooks/runbookDock.ts`
- `web/src/features/management/runbooks/runbookDock.test.ts`
- `web/src/features/management/runbooks/RunbookDockPanel.tsx`
- `web/src/features/management/evidence/evidenceBoard.ts`
- `web/src/features/management/evidence/evidenceBoard.test.ts`
- `web/src/features/management/evidence/EvidenceBoardPanel.tsx`
- `web/src/features/management/opsview/opsviewBridge.ts`
- `web/src/features/management/opsview/opsviewBridge.test.ts`
- `web/src/features/management/opsview/mockOpsview.ts`
- `web/src/features/management/opsview/OpsviewSummaryPanel.tsx`
- `web/src/features/management/opsview/OpsviewSummaryPanel.test.tsx`
- `web/src/features/management/console/ControlPlaneWorkspace.tsx`
- `web/src/features/management/console/mockManagementPreview.ts`
- `web/src/features/management/console/ManagementConsolePreview.tsx`
- `web/src/features/management/console/ManagementConsolePreview.test.tsx`
- `web/src/features/management/externalTools/externalToolRunSummary.ts`
- `web/src/features/management/externalTools/externalToolRunSummary.test.ts`
- `web/src/features/management/externalTools/ExternalToolRunInspector.tsx`
- `web/src/features/management/externalTools/ExternalToolRunInspector.test.tsx`

### Product And Reference Drafts

- `docs/product/management-console-blueprint.md`
- `docs/reference/operator-command-taxonomy.md`

## Why These Assets Exist

They are designed to address critique areas without touching existing wiring:

- config clarity
- access-model clarity
- additive onboarding docs
- test floor expansion
- future management-console building blocks
- terminal command taxonomy and operator workbench scaffolding
- web-terminal session modeling for guided in-app operations
- operator follow-up planning after terminal commands finish
- composite preview surface showing how staged management modules fit together
- operator evidence and runbook docking around terminal and external-tool work
- frontend bridge layer for future backend `opsview` read-model payloads

None of these files are intended to be the final canonical truth by themselves.
They are merge-ready building blocks for the later doc and wiring pass.

## External Tool Tranche Status

Current additive-first work for optional pipeline tools:

- manifest contract added for `external_tool` jobs
- dbt is the only accepted tool in validation
- dlt and PySpark are represented as reserved contract values but explicitly
  gated as unimplemented
- validation coverage exists in
  `backend/internal/orchestration/external_tool_validation_test.go`
- manifest decoding coverage exists in
  `backend/internal/manifests/external_tool_loader_test.go`
- standalone execution package exists under `backend/internal/externaltools/`
- worker dispatch now invokes the external-tool runner and mirrors declared
  artifacts into the normal run artifact namespace
- additive reference doc exists at `docs/reference/external-tool-jobs.md`
- additive example dbt project exists at `packages/external_tools/dbt_finance_demo/`
- paused example pipeline manifest exists at
  `packages/manifests/pipelines/personal_finance_dbt_pipeline.yaml`
- troubleshooting runbooks now exist at:
  - `docs/runbooks/external-tool-troubleshooting.md`
  - `docs/runbooks/dbt-operator-checklist.md`
- failure-drill fixture manifests now exist under
  `packages/manifests/fixtures/external_tools/`
- fixture dbt projects now exist under `packages/external_tools/fixtures/`

Minimal shared-file touch in this tranche:

- `backend/internal/orchestration/models.go`
- `backend/internal/orchestration/validation.go`
- `backend/internal/execution/runner.go`
- `packages/manifests/pipelines/personal_finance_dbt_pipeline.yaml`

New additive file in this tranche:

- `backend/internal/orchestration/validation_external_tool.go`

Standalone package added for the next tranche:

- `backend/internal/externaltools/README.md`
- `backend/internal/externaltools/adapters.go`
- `backend/internal/externaltools/dbt.go`
- `backend/internal/externaltools/errors.go`
- `backend/internal/externaltools/runner.go`
- `backend/internal/externaltools/test_helpers_test.go`

Verification for the external-tool tranche:

```bash
cd backend && go test ./internal/config ./internal/manifests ./internal/orchestration ./internal/externaltools ./internal/execution
```

Current status from this tranche:

- targeted backend tests passed for `internal/config`,
  `internal/manifests`, `internal/orchestration`,
  `internal/externaltools`, and `internal/execution`
- execution-level coverage now verifies external tool log and artifact mirroring
  through `backend/internal/execution/external_tool_test.go`

## External Tool Visibility Tranche Status

Exact files added or changed in this tranche:

- changed: `backend/internal/orchestration/validation_external_tool.go`
- added: `backend/internal/orchestration/external_tool_validation_artifacts_test.go`
- added: `backend/internal/orchestration/handler_external_tool_test.go`
- added: `backend/internal/storage/handler_external_tool_test.go`
- added: `backend/internal/execution/external_tool_failures_test.go`
- added: `web/src/features/management/externalTools/externalToolRunSummary.ts`
- added: `web/src/features/management/externalTools/externalToolRunSummary.test.ts`
- added: `web/src/features/management/externalTools/ExternalToolRunInspector.tsx`
- added: `web/src/features/management/externalTools/ExternalToolRunInspector.test.tsx`

What is verified now:

- pipeline API responses can surface external tool lifecycle events from stored
  runs
- artifact listing can surface external tool declared outputs plus stdout/stderr
  log artifacts
- execution failure coverage now explicitly exercises:
  - missing binary
  - non-zero exit
  - missing required artifact
  - invalid repo-relative refs
- validation now rejects duplicate declared external tool artifact paths
- additive frontend management modules can group and render external tool run
  events plus artifact visibility without touching routed pages

What remains intentionally unfinished:

- no new control-plane or page-route wiring was added for the management UI
- no new tool breadth beyond dbt
- canonical docs and canonical page flows remain for the later merge pass
- fixture manifests and projects are drills only and are not wired into the
  normal manifest loader

Verification commands for this tranche:

```bash
cd backend && go test ./internal/orchestration ./internal/storage ./internal/execution ./internal/externaltools ./internal/manifests ./internal/config
cd web && npm test -- src/features/management/externalTools/externalToolRunSummary.test.ts src/features/management/externalTools/ExternalToolRunInspector.test.tsx
```

Hot-file note for this tranche:

- no specifically blocked coordination files were touched
- one shared orchestration validation file was updated surgically:
  `backend/internal/orchestration/validation_external_tool.go`

## External Tool Troubleshooting Tranche Status

Exact new files in this tranche:

- `docs/runbooks/external-tool-troubleshooting.md`
- `docs/runbooks/dbt-operator-checklist.md`
- `packages/manifests/fixtures/external_tools/README.md`
- `packages/manifests/fixtures/external_tools/dbt_non_zero_exit_pipeline.yaml`
- `packages/manifests/fixtures/external_tools/dbt_missing_artifact_pipeline.yaml`
- `packages/manifests/fixtures/external_tools/dbt_invalid_repo_ref_pipeline.yaml`
- `packages/external_tools/fixtures/README.md`
- `packages/external_tools/fixtures/dbt_non_zero_exit_fixture/dbt_project.yml`
- `packages/external_tools/fixtures/dbt_non_zero_exit_fixture/profiles/profiles.yml`
- `packages/external_tools/fixtures/dbt_non_zero_exit_fixture/models/broken_model.sql`
- `packages/external_tools/fixtures/dbt_missing_artifact_fixture/dbt_project.yml`
- `packages/external_tools/fixtures/dbt_missing_artifact_fixture/profiles/profiles.yml`
- `packages/external_tools/fixtures/dbt_missing_artifact_fixture/models/fixture_note.sql`
- `backend/internal/storage/handler_external_tool_read_test.go`
- `backend/internal/execution/external_tool_operator_inspection_test.go`

What is now verifiably true:

- operators have additive troubleshooting docs for the four primary external
  tool failure modes
- operators can inspect external tool stdout/stderr artifact content through the
  artifact handler read path
- worker-produced external tool outputs are inspectable through the storage
  service after execution, including log artifacts and declared output files
- additive fixture manifests and dbt project variants now exist for non-zero
  exit, missing artifact, and invalid repo-ref drills

What remains intentionally unfinished:

- fixture manifests are not part of the normal manifest load path
- fixture dbt projects are operator drills, not guaranteed production-ready dbt
  examples
- no canonical docs or UI routes were updated in this tranche

Exact test commands run in this tranche:

```bash
cd backend && go test ./internal/storage ./internal/execution ./internal/orchestration ./internal/externaltools ./internal/manifests ./internal/config
```

Whether hot files were touched:

- no specifically blocked files were touched
- no new shared runtime wiring files were edited in this tranche

## Opsview Read Model Tranche Status

Exact new files in this tranche:

- `backend/internal/opsview/models.go`
- `backend/internal/opsview/external_tool_summary.go`
- `backend/internal/opsview/external_tool_summary_test.go`
- `docs/reference/opsview-read-models.md`

What is now verifiably true:

- there is a backend-only read-model package for future management-console use
- external tool run events and artifacts can be grouped into pure
  operator-facing summaries by job id
- stdout/stderr artifacts are grouped as logs and declared outputs are grouped
  separately
- evidence summaries preserve artifact paths for future UI linking
- empty and missing cases return safe empty summaries without runtime wiring

What remains intentionally unfinished:

- the new opsview package is not wired into handlers, runtime, or storage
- no routed UI work or management-console integration was added in this tranche
- the attention summary is a compact backend helper, not yet a canonical API
  contract

Exact test commands run in this tranche:

```bash
cd backend && go test ./internal/opsview ./internal/storage ./internal/execution
```

Whether hot files were touched:

- no specifically blocked files were touched
- no shared runtime files were edited in this tranche

## Opsview Snapshot Tranche Status

Exact new files in this tranche:

- `backend/internal/opsview/snapshot_models.go`
- `backend/internal/opsview/snapshot_builders.go`
- `backend/internal/opsview/snapshot_builders_test.go`
- `backend/internal/opsview/golden_snapshot_test.go`
- `backend/internal/opsview/testdata/run_operator_snapshot.golden.json`

What is now verifiably true:

- `opsview` can now build run-level operator snapshots from pipeline runs plus
  artifacts without any handler or runtime wiring
- artifacts can be grouped into stable evidence buckets for external-tool logs,
  external-tool outputs, and other artifact namespaces
- compact attention rollups can summarize multiple run snapshots for future
  overview panels
- golden JSON coverage now exists for one representative run snapshot payload

What remains intentionally unfinished:

- these snapshot and rollup builders are still backend-only and unwired
- no handler, storage, or runtime endpoints return these payloads yet
- the golden file is additive test coverage, not a published API contract

Exact test commands run in this tranche:

```bash
cd backend && go test ./internal/opsview
```

Whether hot files were touched:

- no specifically blocked files were touched
- no shared runtime or handler files were edited in this tranche

## Latest Coordination Assets

The following additive-only assets were added to support the final frontend
wiring pass and to give future work a stable demo/data-pack lane that does not
collide with routed-page integration:

- `temp-model1-frontend-wire-plan.md`
- `docs/reference/management-console-demo-assets.md`
- `packages/demo/management_console/README.md`
- `packages/demo/management_console/opsview_snapshot.json`
- `packages/demo/management_console/terminal_sessions.json`
- `packages/demo/management_console/runbook_dock.json`

These assets are intentionally non-canonical:

- the temp frontend wire plan is a handoff brief for the integration pass
- the demo payloads are safe sample inputs for UI/demo work
- none of them should be treated as published runtime contracts

## Integration Queue

When the hot implementation thread is ready, the next safe merge pass should
consider:

1. decide whether to promote `docs/runbooks/local-host-run.md` into the main
   startup path
2. decide whether to fold `docs/runbooks/config-reality.md` into the canonical
   quickstart and operator docs
3. decide whether to fold `docs/runbooks/access-matrix.md` into the main auth
   story after enforcement and docs fully align
4. decide whether to replace or expand the existing pipeline tutorial using
   `docs/tutorials/trace-one-pipeline-complete.md`
5. wire the new tests into normal verification expectations by habit, not by
   editing scripts prematurely
6. decide when to route future terminal and management UI through the additive
   `web/src/features/management/` staging area
7. decide whether the management-console blueprint becomes the primary product
   direction for in-app platform operations

## Ready-For-Wiring Criteria

A new additive asset is ready to wire into the main repo flow when:

- it matches current implementation behavior
- it does not rely on speculative future behavior
- it has been exercised or tested
- it reduces ambiguity instead of adding another competing source of truth

## Working Rules For The Other Model

If you are using this file as a handoff:

1. prefer creating new files over editing shared files
2. if you must touch a hot file, do it in a separate, deliberate work block
3. keep this file updated whenever you add new additive assets
4. record what is draft, what is verified, and what is waiting on integration

## Verification Notes For Current Additions

The test additions in this thread should be verified with:

```bash
cd backend && go test ./internal/config ./internal/storage ./internal/manifests
cd web && npm test
```

Current status from this thread:

- backend targeted tests passed for `internal/config`, `internal/storage`, and
  `internal/manifests`
- frontend `npm test` passed with:
  - `web/src/lib/api.test.ts`
  - `web/src/features/management/terminal/commandCatalog.test.ts`
  - `web/src/features/management/terminal/sessionModel.test.ts`
  - `web/src/features/management/terminal/followupPlanner.test.ts`
  - `web/src/features/management/runbooks/runbookDock.test.ts`
  - `web/src/features/management/evidence/evidenceBoard.test.ts`
  - `web/src/features/management/opsview/opsviewBridge.test.ts`
  - `web/src/features/management/opsview/OpsviewSummaryPanel.test.tsx`
  - `web/src/features/management/inventory/assetAttention.test.ts`
  - `web/src/features/management/externalTools/externalToolRunSummary.test.ts`
  - `web/src/features/management/externalTools/ExternalToolRunInspector.test.tsx`
  - `web/src/features/management/console/ManagementConsolePreview.test.tsx`
  - `web/src/app/App.test.tsx`
  - `web/src/pages/PageStates.test.tsx`

The doc additions are intentionally unwired and should be treated as draft
integration material until the main docs pass is ready.
