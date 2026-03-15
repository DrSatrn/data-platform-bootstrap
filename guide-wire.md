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
- `backend/internal/storage/service_test.go`
- `backend/internal/manifests/loader_test.go`
- `backend/internal/manifests/external_tool_loader_test.go`
- `web/src/lib/api.test.ts`
- `backend/internal/orchestration/external_tool_validation_test.go`
- `backend/internal/externaltools/dbt_test.go`
- `backend/internal/externaltools/runner_test.go`

### Docs

- `docs/runbooks/config-reality.md`
- `docs/runbooks/access-matrix.md`
- `docs/runbooks/local-host-run.md`
- `docs/tutorials/trace-one-pipeline-complete.md`
- `docs/reference/external-tool-jobs.md`

### Frontend Management Prototypes

- `web/src/features/management/README.md`
- `web/src/features/management/types.ts`
- `web/src/features/management/mockControlPlane.ts`
- `web/src/features/management/terminal/commandCatalog.ts`
- `web/src/features/management/terminal/commandCatalog.test.ts`
- `web/src/features/management/terminal/OperatorWorkbench.tsx`
- `web/src/features/management/inventory/assetAttention.ts`
- `web/src/features/management/inventory/assetAttention.test.ts`
- `web/src/features/management/console/ControlPlaneWorkspace.tsx`

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
cd backend && go test ./internal/orchestration ./internal/manifests ./internal/externaltools ./internal/execution
```

Current status from this tranche:

- targeted backend tests passed for `internal/orchestration`,
  `internal/manifests`, `internal/externaltools`, and `internal/execution`

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
  - `web/src/features/management/inventory/assetAttention.test.ts`

The doc additions are intentionally unwired and should be treated as draft
integration material until the main docs pass is ready.
