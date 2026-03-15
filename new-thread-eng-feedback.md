# Engineering Feedback

Review basis: snapshot of the repository observed on March 15, 2026.

Important context:
- This repo is actively in flight and being edited concurrently.
- Feedback below should be read as a second-pass engineering review of the current visible state, not a judgment on the final intended implementation.
- I have biased this toward durable architectural and product-engineering concerns over transient polish or in-progress feature gaps.

## What This Repo Is Clearly Going For

This project continues to aim to be a local-first, self-hosted internal data platform with first-party ownership of orchestration, scheduling, execution, metadata, and reporting.

The previous worksteps successfully brought the platform to a v2-ready state for the personal-finance slice, pushing the boundaries of what a built-in UI and a self-contained runtime can do. The goals of educational clarity and "learning by reading the implementation" remain strong. 

## Strong Parts

### 1. The PostgreSQL control-plane is becoming real
The project is successfully making the transition towards a robust database-backed metadata and queue state while gracefully preserving local-first resilience. The duality of storage models is being managed carefully.

### 2. Operational surfaces are meaningfully expanding
First-party reporting with charts, metric browsers, dashboard presets, and audit logs proves that the platform does not need to outsource the most valuable user-facing workflows to third-party tools. 

### 3. The recovery story is half-complete but solid
Building a native `platformctl backup create` that bundles Postgres state, JSON manifests, and DuckDB analytical layers into a portable artifact is a huge step for self-hosted confidence.

### 4. Built-in testing and telemetry
The platform has a genuine first-party benchmark tool, smoke validations that now assert administrative behaviors like backups, and an execution loop that doesn't just guess at performance.

## Hyper-Critical Docs Review

The documentation has matured significantly since the last major review. The runbooks have been consolidated, roles are explicitly mapped, and the differences between Compose and host-run deployments are much clearer.

However, the docs now face a new class of challenges:

### 1. The "recovery" docs still leave the operator hanging
The backup bundle generation is impressively automated, but the restore process is a manual extraction drill. If the docs advertise "backups," the operator expects a symmetrical "restore" process. Right now, the docs describe the contents of a backup tarball instead of providing a bulletproof one-command restore path.

### 2. The metadata normalization story is aspirational
The docs and architecture notes discuss PostgreSQL as the control plane, but for catalogs and dashboards, it's really just a projection of the file system. The docs blur the line between "lives natively in Postgres" and "gets pushed to Postgres from the filesystem on startup."

### 3. The authorization model is well-documented but inherently brittle
The documentation correctly states that RBAC uses lightweight static tokens. While true, relying on an environment variable (`PLATFORM_ACCESS_TOKENS`) for role assignment is a major friction point as the platform claims to handle "enterprise-readiness." The docs lean heavily on this model being "sufficient" when it is likely going to break down for any team larger than two people.

## Documentation Recommendations

1. Address the Restore Symmetry
If there is a `backup create` command in `platformctl`, document the explicit strategy for a `restore` command. Even if it's not built yet, the docs should explicitly state "Restoring from backups is currently a manual extraction process (see below), but automation is prioritized."

2. Clarify Metadata Primacy
Explicitly document that the reporting and catalog services are still fundamentally manifest-first, and the PostgreSQL representation is an index/projection, not the primary source of truth (yet).

## Main Engineering Concerns

### 1. Restore Automation is Missing
It cannot be considered a mature platform feature if backups are created continuously but restoring them requires the user to manually coordinate file extractions, SQLite/DuckDB replacements, and Postgres imports. Taking a backup is easy; safely destroying and recreating state from a backup is hard. That automation needs to exist in `platformctl`.

### 2. Partial Postgres Normalization
Currently, `metadata` and `dashboards` are heavily tied to the `FileStore` and sync-on-read or projection-on-startup paradigms. True normalization means the database becomes the undisputed source of truth, and edits (like dashboard layouts) happen via the API and write directly to relational tables, optionally exporting to files, rather than the other way around.

### 3. Lightweight RBAC limits governance
The `PLATFORM_ACCESS_TOKENS` environment variable injection is an elegant hack for the scaffolding phase. However, a local-first application looking to provide audit trails and real platform safety needs a localized identity store with real session tracking, not just static tokens.

### 4. Reporting features are rigid
The reporting feature has bounded widget types, which is excellent, but it lacks dynamic reporting layout configurations, fine-grained sharing semantics, and report-level advanced interactivity. It works for a simple finance slice but will struggle to scale against generic analytical queries.

### 5. Benchmark depth is superficial
The `infra/scripts/benchmark_suite.sh` covers basic latency. It does not stress-test queue depth, concurrency bottlenecks under load, or the limits of the file-backed reporting store versus the Postgres store. 

## Architectural Advice

### 1. Close the loop on Recovery Automation
Stop relying on manual runbooks for restore operations. Embed the logic into Go so it is testable, auditable, and repeatable.

### 2. Commit to Postgres for dynamic entities
Dashboards and metadata annotations should transition to being fully normalized CRUD entities in Postgres. The manifests should become "deployment" mechanisms (e.g., seeding a clean environment) rather than the dynamic source of truth that powers the UI. 

### 3. Build a real Identity Layer
Shift from purely static token injection to a local database-driven user and role registry. Keep it simple (local SQLite/Postgres tables), but let it be manageable via the UI/API.

## Product Advice

Operator ergonomics have come a long way. The next milestone for the platform is to make it feel "bulletproof" rather than just "usable." A massive driver of this will be the one-click restore flow. If an operator knows they can wipe their environment and restore it flawlessly with `platformctl restore <bundle>`, trust will skyrocket.

## Suggested Next Priorities

If I were sequencing the next high-leverage pass, I would do this:

1. Restore Automation
- Build `platformctl backup restore` to symmetrically complement `backup create`.

2. Complete Identity and Session Layer
- Move away from `PLATFORM_ACCESS_TOKENS`.
- Establish a normalized Identity model in the control plane for better Governance and Audit trails.

3. Reporting and Dashboard Normalization
- Decouple dashboard persistence from manifests during runtime. Make Postgres the primary mutable store.
- Add advanced layout tools and sharing semantics to the dashboard entities.

4. Performance Stress Testing
- Expand the benchmark suite into concurrency and queue-depth load tests.

## Bottom Line

This repo continues to be incredibly strong.

The platform is delivering on its promise of being self-contained and highly operable. The transition to the v2-ready state was executed with discipline, preserving the local-first fallback mechanisms.

The remaining gaps are natural growing pains when transitioning from a scaffolded architecture to an enterprise-trusted platform. The focus must be placed on normalizing statefully mutable concepts entirely into the database and closing the loop on critical operational flows like disaster recovery.

## What I Verified During Review

I reviewed:
- `backend/internal/reporting/store.go` and the multi-store logic.
- `backend/internal/metadata/catalog.go` and projection capabilities.
- `backend/internal/authz/service.go` defining the lightweight static RBAC.
- `backend/internal/app/runtime.go` to inspect the startup routines and persistence modes.
- `backend/internal/analytics/service.go` and the duckdb constraints.

Current state:
- The previous feedback checklist items are complete. PostgreSQL fallback, backup creation, RBAC token passing, and API documentation are substantially improved.

## Async Handoff Framework For The Next Model

This section is written as an execution contract for the next model working on the repo. 

## Priority Order

Work in this order unless there is a strong reason not to:

1. One-command restore automation
2. Identity/auth full system implementation
3. Deep PostgreSQL normalization for dashboards and metadata
4. Reporting layout tools and extended dataset drill-downs
5. Benchmark expansion

## Global Gating Rules

Before any area is considered complete:
- The happy path must be runnable without hidden assumptions.
- There must be an explicit verification step.
- Any unfinished edge must be called out plainly in docs.

## Workstream 1: Restore Automation
Checklist:
- [x] Build `platformctl backup restore` that cleanly ingests a `.tar.gz`.
- [x] Automate Postgres and filesystem reconstitution.
- [x] Update `docs/runbooks/backups.md` reflecting the automated process.
- [x] Add a `make restore-e2e` drill.

## Workstream 2: Identity and Auth Native System
Checklist:
- [ ] Replace static `PLATFORM_ACCESS_TOKENS` with a SQLite/Postgres-backed Identity/User store.
- [ ] Update `/api/v1/session` to support real session management (login/logout).
- [ ] Expand the Audit trail to reference database-backed user IDs, not just static subjects.
- [ ] Ensure backward compatibility with `PLATFORM_ADMIN_TOKEN` for bootstrapping.

## Workstream 3: Deep Postgres Normalization
Checklist:
- [ ] Move dashboards to be purely Postgres-backed in runtime, treating manifests only as initial seeds.
- [ ] Do the same for `metadata` catalogs so UI-driven annotations persist directly to the database.
- [ ] Update `backend/internal/reporting/store.go` and `metadata/catalog.go` to reflect database-first supremacy over sync-on-read mechanisms.

## Workstream 4: Reporting Polish
Checklist:
- [ ] Add explicit layout grid metadata to the dashboard widget domain schema.
- [ ] Allow dynamic reordering and resizing of widgets in the UI.
- [ ] Enhance dataset drill-down capabilities within the analytics service.

## Workstream 5: Benchmark Breadth
Checklist:
- [ ] Add concurrent load testing to `benchmark_suite.sh`.
- [ ] Add queue-depth and scheduler latency assertions to the benchmark artifacts.

## Required Closeout Format For Each Completed Area

When the next model finishes a workstream, it should leave behind a short note in its own summary covering:
- what changed
- what is now verifiably true
- what remains intentionally unfinished
- what command or workflow was used to verify it

## Final Instruction To The Next Model

Do not move forward with completely new product features (e.g. extending domain boundaries) until the platform architecture supports native database state handling and one-click restore capabilities. Optimization for operator ease and truthfulness is the priority.
