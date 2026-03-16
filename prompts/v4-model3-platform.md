# Model 3 v4 — Runtime Integration, Session Transport, And Authoring Release Safety

**Priority:** Critical  
**Owner:** Model 3  
**Merge order:** 3rd  

---

## Mission

Make the new authoring model real at runtime. The scheduler, worker, API, and
deployment workflows must safely consume published authored objects, support
platform terminal sessions, and preserve import/export and recovery discipline.

Use [docs/product/admin-ide-roadmap.md](/Users/streanor/Documents/Playground/coding-tracker/docs/product/admin-ide-roadmap.md)
as the execution contract.

## Constraints

- Published versions only should be runnable.
- Draft edits must never silently affect live scheduled execution.
- Keep import/export and restore stories believable.
- Avoid destabilizing existing runtime behavior while wiring in authored objects.

## Tasks (In Priority Order)

### Task 1: Runtime Consumption Of Published Pipelines

Wire scheduler and worker to consume published pipeline versions from the new
authoring model.

Required capabilities:

- published pipeline resolver
- draft isolation from runtime
- version-aware trigger and run attribution
- rollback-safe published version selection

Completion signal:

- integration test proves a published pipeline runs while a modified draft does
  not affect live execution until published

### Task 2: Manifest Import / Export And Recovery Compatibility

Keep the existing repo and recovery discipline compatible with the new authored
model.

Required capabilities:

- import manifest into draft
- export published version to manifest shape
- include authored objects in backup/restore where appropriate
- document recovery expectations clearly

Completion signal:

- restore/import/export drills show authored pipelines survive recovery and can
  be re-exported

### Task 3: Platform Terminal Session Transport

Provide the backend/session transport needed for a real platform terminal.

Required capabilities:

- create terminal sessions
- persist session metadata and transcript
- stream command output
- support cancellation/timeouts
- allow reconnect to prior sessions

Completion signal:

- integration test proves streamed output is available through the session API

### Task 4: Release And Smoke Coverage For Authoring

Update the platform verification path so the new authoring stack is part of the
release gate.

Required changes:

- extend smoke coverage for connector creation/test and draft/publish flows
- document operator bootstrap for the authoring model
- keep packaged deployment and restore stories aligned

Completion signal:

- smoke or dedicated authoring drill proves the end-to-end path from connector
  creation to published pipeline run

## Escalation Triggers

- If runtime consumption of published pipelines requires invasive control-plane
  surgery, pause and document the narrowest safe integration seam.
- If streaming transport introduces major operational complexity, start with a
  simple durable polling model and document the tradeoff clearly.
