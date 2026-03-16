# V4 Review Coordinator Report — Admin IDE, Connector Studio, And Platform Authoring

**Priority:** Critical  
**Audience:** Coordinator / reviewer model  
**Program:** Move the management console from operator surface to full authoring environment  

---

## Mission

Coordinate three models to deliver the next major product phase:

- GUI-first connector and data-source creation
- DB-backed pipeline authoring and publish workflows
- a real platform terminal / TUI
- runtime-safe consumption of published authored objects

The shared design source of truth for this program is:

- [docs/product/admin-ide-roadmap.md](/Users/streanor/Documents/Playground/coding-tracker/docs/product/admin-ide-roadmap.md)

The current system is already a strong self-hosted platform with real runtime,
reporting, metadata, auth, backup/restore, and management surfaces. This
program is about turning the admin area into a true IDE and connector studio.

## Global Program Rules

1. DB-backed authoring is the live source of truth for UI-created drafts and
   published objects.
2. Repo manifests remain seed/import/export artifacts, not the primary mutable
   editing substrate.
3. The terminal is platform-scoped, not arbitrary shell access.
4. Secret values must never be stored in plain authored pipeline definitions.
5. Every sensitive authoring action must be permission-gated and audited.
6. Keep docs in sync during each workstream, not only at the end.

## Merge Order

1. `v4-model1-backend.md`
2. `v4-model2-frontend.md`
3. `v4-model3-platform.md`

This order is intentional:

- backend authoring types and APIs must exist before the frontend can wire to
  them safely
- runtime integration should come after the authoring model exists

## Shared Acceptance Bar

Before the program is considered complete, all of these must be true:

- a new data source can be created entirely in the UI
- a data source can be tested and previewed in the UI
- a draft pipeline can be created and edited entirely in the UI
- a draft pipeline can be validated and published in the UI
- only published pipeline versions are consumed by scheduler and worker
- the management terminal supports real session-based platform commands
- audit records exist for create/edit/test/publish/rollback actions
- docs clearly explain how DB authoring relates to manifest import/export

## Model Responsibilities

### Model 1 — Backend Authoring And Validation

Own:

- connector registry
- data-source CRUD backend
- pipeline draft/publish persistence
- validation endpoints
- secret reference model
- audit and role enforcement for authoring actions

### Model 2 — Frontend Authoring UX

Own:

- connector studio UI
- pipeline editor UI
- draft/publish UX
- terminal/TUI UX
- field validation and operator ergonomics

### Model 3 — Runtime And Platform Integration

Own:

- runtime consumption of published authored objects
- import/export compatibility
- release/smoke coverage for the authoring model
- session streaming transport and runtime plumbing

## Review Focus

When reviewing model outputs, prioritize:

1. source-of-truth clarity
2. runtime safety
3. secret handling
4. auditability
5. operator followability
6. testability and rollback safety

## Required Closeout From Each Model

Each model must report:

- what changed
- what is now verifiably true
- what remains intentionally unfinished
- exact files touched
- commands/tests run

## Coordinator Closeout Questions

At the end of the program, answer:

1. Can a new source be created and tested entirely in the UI?
2. Can a new draft pipeline be created, validated, and published entirely in
   the UI?
3. Is the runtime definitely consuming only published versions?
4. Is the terminal a real session-based platform terminal rather than a button
   launcher?
5. Are secrets, permissions, and audits believable for internal production use?
