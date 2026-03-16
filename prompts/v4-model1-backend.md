# Model 1 v4 — Backend Authoring, Connector Registry, And Publish Safety

**Priority:** Critical  
**Owner:** Model 1  
**Merge order:** 1st  

---

## Mission

Build the backend authoring substrate for the Admin IDE program. The outcome
must be a safe, DB-backed authoring model for data sources and draft/published
pipelines, with permissions, validation, secret references, and audit trails.

Use [docs/product/admin-ide-roadmap.md](/Users/streanor/Documents/Playground/coding-tracker/docs/product/admin-ide-roadmap.md)
as the execution contract.

## Constraints

- Do not make repo YAML manifests the primary mutable editing path.
- Do not store raw secrets in pipeline definitions.
- Keep authoring runtime-safe: only published versions should be runnable.
- Do not widen into frontend implementation beyond minimal API seam changes.

## Tasks (In Priority Order)

### Task 1: Connector Registry And Data Source Persistence

Build a backend connector registry and database-backed data-source model.

Required capabilities:

- list supported connector types
- describe connector capability metadata
- create/update/delete data sources
- store non-secret config fields
- store secret references, not raw values
- test connection
- preview/discover schema

Suggested endpoints:

- `GET /api/v1/connectors`
- `GET /api/v1/connectors/:type`
- `GET /api/v1/data-sources`
- `POST /api/v1/data-sources`
- `PATCH /api/v1/data-sources/:id`
- `POST /api/v1/data-sources/:id/test`
- `POST /api/v1/data-sources/:id/discover`
- `POST /api/v1/data-sources/:id/preview`

Completion signal:

- backend tests prove local-file and Postgres data-source creation, test, and
  preview flows

### Task 2: Draft / Published Pipeline Persistence

Add DB-backed authored pipeline persistence with explicit draft/publish
separation.

Required capabilities:

- create draft pipeline
- edit jobs in draft
- validate draft
- publish draft to immutable version
- keep published versions separate from in-progress drafts
- rollback by promoting an earlier published version

Suggested model:

- `pipeline_drafts`
- `pipeline_draft_jobs`
- `pipeline_versions`
- `pipeline_version_jobs`

Completion signal:

- backend tests prove a draft can be created, validated, published, and later
  rolled back

### Task 3: Validation And Import/Export Bridges

Add validation and manifest-bridge helpers so authored pipelines can be checked
and exported without making manifests the primary mutable store.

Required capabilities:

- validate authored pipeline objects using the same durable rules as manifest
  pipelines
- export a published pipeline version to manifest-shaped YAML
- import a manifest pipeline into a draft record

Completion signal:

- tests prove draft validation catches dependency/config errors and that
  import/export preserves runnable pipeline semantics

### Task 4: Permissions, Audit, And Secret Hygiene

Authoring actions must be properly gated and auditable.

Required changes:

- enforce role checks for connector edit/test, draft edit, publish, and rollback
- add audit events for create/update/delete/test/publish/rollback actions
- ensure API responses never leak raw secret material

Completion signal:

- auth/audit tests prove unauthorized access is rejected and authorized actions
  emit audit events

## Escalation Triggers

- If the current DB schema layout makes clean draft/publish modeling too messy,
  pause and propose the minimum normalized schema change rather than layering on
  ad hoc JSON blobs.
- If secret storage requires a broader security decision, implement secret
  references with clear TODO boundaries and document the chosen interim model.
