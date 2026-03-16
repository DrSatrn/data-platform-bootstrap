# Admin IDE And Connector Studio Roadmap

This document is the coordination plan for moving the current operator console
into a true authoring environment. It is intended for multiple agents working
in parallel, so it is explicit about boundaries, sequencing, and acceptance
criteria.

## Current State

The product already has:

- a working operator console
- a guided admin terminal
- manifest-backed pipelines
- bounded job types for ingest, SQL, Python, quality, metrics, and external tools
- real scheduler, worker, API, auth, audit, and reporting surfaces

The product does not yet have:

- a full pipeline IDE
- GUI-first connector and data-source creation
- a real terminal-style TUI experience
- DB-backed draft/publish authoring workflows

That means the current management surface is operational, but it is not yet a
true build-and-configure console.

## Product Goal

Build an in-app authoring experience that allows an operator to:

- create and edit pipelines without hand-authoring YAML
- define jobs, dependencies, retries, schedules, and outputs in the UI
- create and test data sources and ingestion connectors in the UI
- discover schemas and generate ingest jobs from those connections
- publish pipeline changes safely
- operate the platform through a session-based platform terminal

The goal is a platform-oriented IDE, not arbitrary shell access.

## Non-Negotiable Design Decisions

These decisions should be treated as defaults unless a later architecture review
explicitly changes them.

### 1. Source of truth for authoring

Pipeline and connector authoring should become database-backed for live edits.

Why:

- UI editing needs draft state, partial saves, and validation without mutating
  repo files directly.
- Multiple operators need a shared live authoring model.
- Publish, diff, rollback, and audit are much cleaner with explicit DB objects.

Repo-managed manifests should remain supported as:

- seed input
- export target
- reviewable snapshot
- disaster recovery compatibility path

Recommended model:

- DB-backed drafts and published records are the runtime authoring source
- manifests are import/export and seed artifacts

### 2. Terminal model

The admin terminal should become a platform terminal, not a host shell.

It should support:

- structured commands
- streamed output
- command history
- guided workflows
- long-running tasks
- session transcript persistence

It should not support:

- arbitrary OS shell access
- unrestricted filesystem browsing
- unrestricted host process execution

### 3. Connector model

Connectors should be described by a typed backend registry with per-connector
config schemas and capabilities.

Each connector should declare:

- connector type
- capability flags
- config schema
- secret requirements
- test-connection support
- discovery/schema introspection support
- extract behavior
- preview behavior

## Delivery Sequence

Build in this order.

### Phase 1. Connector Studio

This is the highest-value first step because it gives the team a GUI path to
add real sources.

#### Deliverables

- data source CRUD in backend and UI
- connector registry in backend
- schema-driven connector forms in React
- secret-aware connection model
- test connection endpoint
- schema discovery endpoint
- preview/sample endpoint
- ingestion job generator
- asset registration defaults

#### Connector types for first pass

- local CSV
- local JSON
- local Parquet
- PostgreSQL
- MySQL
- generic JDBC-style placeholder for future DB classes
- Oracle / Oracle Cloud defined as staged future target if not completed in v1

#### Core backend work

- add `connectors` and `data_sources` tables
- add connector registry interface
- add connector-specific validators
- add secret reference storage model
- add `/api/v1/connectors`
- add `/api/v1/data-sources`
- add `/api/v1/data-sources/test`
- add `/api/v1/data-sources/discover`
- add `/api/v1/data-sources/preview`

#### Core frontend work

- add data sources page
- add connector type picker
- add connector-specific forms
- add connection test UX
- add discovery and preview UX
- add "generate ingest job" action

#### Acceptance criteria

- an operator can create a local-file data source in the UI
- an operator can create a Postgres data source in the UI
- connection test success/failure is visible in the UI
- schema discovery and preview work for supported connectors
- generated ingest config can be attached to a pipeline draft

### Phase 2. Structured Pipeline Editor

This should come before the visual DAG builder. Build a dependable forms-based
editor first.

#### Deliverables

- pipeline CRUD in backend and UI
- draft vs published pipeline model
- job add/edit/remove UI
- dependency editing UI
- schedule editor
- retry/timeout/idempotency editor
- input/output asset editor
- validation panel
- publish workflow
- diff view between draft and published

#### Core backend work

- add `pipeline_drafts`, `pipeline_versions`, and `pipeline_jobs` persistence
- add publish transaction model
- add validation endpoint for draft pipelines
- add import/export support between DB authoring model and manifest shape
- add audit events for create/edit/publish/rollback

#### Core frontend work

- add pipeline editor page
- add job-type-specific editors
- add dependency list editing
- add schedule and retry settings editors
- add validation error panel
- add publish and rollback controls

#### Acceptance criteria

- an operator can create a draft pipeline in the UI
- an operator can add ingest, SQL, Python, quality, metric, and external-tool jobs
- validation prevents publish when required fields or dependency rules are broken
- publish produces a runnable pipeline version
- scheduler and worker use the published version only

### Phase 3. Platform Terminal / TUI

Replace the current button-driven console with a real platform session model.

#### Deliverables

- terminal sessions persisted server-side
- prompt and command history
- streamed command output
- command palette and help
- guided operational workflows
- evidence/runbook/follow-up docking
- resumable session transcripts

#### Terminal scope

Allowed command classes:

- pipeline and run management
- data source testing and preview
- validation and publish checks
- backup and restore workflows
- retention and artifact inspection
- user/admin management
- diagnostics and health checks

Not allowed:

- arbitrary `bash`
- arbitrary file deletion
- unrestricted process execution

#### Core backend work

- add terminal session store
- add streaming command transport
- add session transcript persistence
- add command registry and schemas
- add cancellation and timeout handling

#### Core frontend work

- add terminal page with true input/output flow
- add command history
- add streaming render
- add inline help/autocomplete
- add runbook/evidence side panes

#### Acceptance criteria

- an operator can open a terminal session and run real platform commands
- output streams into the UI without page refreshes
- session history is available after reload
- evidence and follow-up items can be linked from a session

### Phase 4. Visual DAG IDE

Only do this after the structured editor is stable.

#### Deliverables

- visual pipeline graph canvas
- drag/drop job nodes
- dependency edge editing
- layout persistence
- inline validation on graph edits
- node inspector panel

#### Acceptance criteria

- visual edits and form edits stay in sync
- operators can build a valid DAG visually without touching YAML
- graph save/publish uses the same draft/publish model as the structured editor

## Cross-Cutting Work Required In Every Phase

### Auth and permissions

Need explicit role gates for:

- viewing connectors
- editing connectors
- testing connections
- editing pipeline drafts
- publishing pipelines
- running privileged terminal commands

Recommended roles:

- `viewer`
- `editor`
- `admin`

Potential future role split:

- `operator`
- `author`
- `platform_admin`

### Audit

Audit every sensitive authoring action:

- create/update/delete connector
- connection test
- schema preview
- pipeline draft save
- pipeline publish
- rollback
- terminal privileged commands

### Secret handling

Secrets must not be stored directly in plain pipeline definitions.

Need:

- secret reference model
- masked display in UI
- test connection using resolved secret values
- export behavior that omits raw secrets

### Validation

Each phase must add:

- unit tests
- backend integration tests
- frontend test coverage
- smoke path updates
- docs/runbook updates

## Suggested Parallel Ownership For Three Agents

### Agent 1: Backend authoring and validation

Own:

- connector registry
- data source persistence
- pipeline draft/publish persistence
- validation endpoints
- audit and permissions

### Agent 2: Frontend authoring UX

Own:

- data source forms
- pipeline editor UI
- visual graph editor
- terminal UX
- form validation and ergonomics

### Agent 3: Platform/runtime integration

Own:

- scheduler/runtime consumption of published pipeline versions
- secret resolution flow
- smoke/release workflows
- restore/import/export compatibility
- performance and reliability checks

## Key Risks

### 1. Source-of-truth confusion

If DB drafts and YAML manifests both mutate live runtime state without a clear
rule, the product will become untrustworthy quickly.

Mitigation:

- DB-backed live authoring
- manifest import/export only
- explicit publish pipeline versioning

### 2. Terminal scope creep

If the terminal becomes arbitrary shell access, security posture and support
cost will degrade.

Mitigation:

- keep command registry explicit
- keep terminal platform-scoped

### 3. Connector sprawl

Adding many connector types without a capability contract will create brittle
one-off code.

Mitigation:

- connector registry
- schema-driven config
- typed capability model

## Definition Of Done For This Program

This admin IDE program is complete when all of the following are true:

- a new data source can be created entirely in the UI
- a source can be tested and previewed in the UI
- a pipeline can be created and edited entirely in the UI
- a pipeline can be validated and published in the UI
- the published pipeline runs end to end through the real scheduler/worker path
- the terminal supports real session-based platform command execution
- audit and permissions cover all authoring actions
- docs explain how the UI authoring model relates to manifests and runtime

## Recommended Immediate Next Work

Do these first:

1. lock the DB-backed authoring source-of-truth decision
2. build the backend connector registry and data-source model
3. build the frontend connector studio for local files and Postgres
4. add draft pipeline persistence and structured pipeline editing

Do not start the visual DAG builder before the structured pipeline editor is
working and published pipelines are runtime-safe.
