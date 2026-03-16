# Model 2 v4 — Frontend Connector Studio, Pipeline Editor, And Platform Terminal UX

**Priority:** Critical  
**Owner:** Model 2  
**Merge order:** 2nd  

---

## Mission

Turn the management area into a real authoring environment. The frontend should
let an operator create data sources, build draft pipelines, validate/publish
them, and use a session-based platform terminal.

Use [docs/product/admin-ide-roadmap.md](/Users/streanor/Documents/Playground/coding-tracker/docs/product/admin-ide-roadmap.md)
as the execution contract.

## Constraints

- Reuse existing management modules where possible.
- Prefer real API-backed flows over new mocks.
- Do not assume manifests are the live editing substrate.
- Keep the terminal platform-oriented, not arbitrary shell access.

## Tasks (In Priority Order)

### Task 1: Connector Studio UI

Build the UI for creating and managing data sources.

Required capabilities:

- list connector types
- choose connector type
- render connector-specific config forms
- create and edit a data source
- test connection
- preview/discover source shape
- generate ingest-job-ready output for pipeline authoring

Required UX targets:

- local CSV / JSON / Parquet
- Postgres
- MySQL
- clear placeholders for staged connector families such as Oracle / cloud object storage

Completion signal:

- browser tests prove a user can fill a connector form, test it, and view a
  preview result

### Task 2: Structured Pipeline Editor

Build a form-based pipeline editor before attempting a visual DAG builder.

Required capabilities:

- list draft and published pipelines
- create a draft pipeline
- add/edit/remove jobs
- edit dependencies
- edit schedule, retries, timeout, idempotency-facing fields
- attach ingest jobs from created data sources
- validate draft
- publish draft
- show diff/summary between draft and published state

Completion signal:

- browser tests prove a user can create a draft pipeline, add jobs, validate,
  and publish it

### Task 3: Platform Terminal UX Upgrade

Replace the current button-style command readout with a real platform terminal
experience.

Required capabilities:

- terminal session list
- command prompt input
- command history
- streaming output view
- guided command help
- runbook/evidence/follow-up side panes
- reconnect to recent sessions

Completion signal:

- browser tests prove a terminal session can be created, a command can be
  executed, and output can be read in-session without page reload

### Task 4: Operator Ergonomics And Navigation

The new authoring experience needs to feel like one coherent product.

Required changes:

- clear navigation between connectors, pipelines, runs, evidence, and terminal
- stateful success/error messaging
- role-aware disable states
- no-crash guarantees on partial/null backend payloads

Completion signal:

- frontend tests cover null-state and error-state rendering for the new surfaces

## Escalation Triggers

- If the backend authoring API shape is still moving, coordinate with Model 1
  and prefer thin adapter hooks over hardcoding unstable payload shapes.
- If a visual graph editor threatens schedule or stability, stop at a strong
  structured editor and document the deferred graph work cleanly.
