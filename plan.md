# Active Build Plan

This file is the short-horizon implementation plan for the current autonomous
work block. It exists to keep the next session focused without forcing future
agents to infer intent from a large git diff or chat transcript.

## Current Build Focus

Move the project from a strong vertical slice into a more fully fledged
internal platform product by prioritizing operational safety, recovery
readiness, access control, and stronger self-hosted platform discipline.

## Why This Slice

The platform already has:

- real orchestration and execution
- DuckDB-backed curated analytics
- local and packaged smoke paths
- dashboard editing and persistence through the browser
- first-party reporting, diagnostics, and metadata surfaces

The next highest-leverage gap is platform hardening:

- access control needs to be real, not a single opaque admin token
- metadata trust signals need to be operator-facing
- validation needs to quantify performance, not just behavior
- recovery needs to be concrete, not implied

## Current Plan

1. Access control hardening
   - Introduce lightweight RBAC with bearer tokens and browser session
     awareness.
   - Protect write paths and admin surfaces.
   - Keep the local-first deployment model simple enough for self-hosted use.

2. Auditability
   - Persist privileged actions in a first-party audit trail.
   - Expose recent audit history in the operations UI and packaged runtime.
   - Use the audit layer as the foundation for later governance and recovery
     workflows.

3. Metadata intelligence and persistence
   - Derive richer catalog coverage and lineage summaries from repo manifests.
   - Surface documentation coverage, quality coverage, freshness, and lineage
     context in the operator UI.
   - Project the synchronized metadata catalog into PostgreSQL so the control
     plane has durable dataset state.

4. Validation and benchmark foundation
   - Add a first-party benchmark command to `platformctl`.
   - Add a repo-owned benchmark script that emits timestamped JSON reports.
   - Build a stronger future E2E validation baseline alongside smoke tests.

5. Recovery and backup foundation
   - Add a first-party backup/export bundle format.
   - Verify bundles in automation instead of assuming they are usable.
   - Expose recovery primitives through the CLI and admin terminal.

6. Next likely follow-on work
   - Normalize more mutable state into PostgreSQL so dashboards and metadata
     become database-first runtime entities rather than projections.
   - Build richer reporting layout and dataset drill-down behavior on top of
     the stronger control-plane state model.
   - Expand the benchmark suite with scheduled-run, artifact, report-save, and
     queue latency budgets.

## Latest Completed Workstep

- Implemented a PostgreSQL-backed native identity and session layer with
  bootstrap-admin compatibility.
- Added `/api/v1/session` login/logout plus admin user-management APIs.
- Expanded audit events to capture database-backed `actor_user_id` values.
- Updated the browser auth flow so operators can sign in with username and
  password while still keeping the bootstrap token override for recovery.
- Extended smoke coverage so packaged deployments now create a native user,
  log in, read product APIs via a session token, and log out again.
- Updated backup and restore handling so native users are now included in
  recovery bundles and restored with hashed credentials while live sessions are
  intentionally cleared.

## Deferred / Prompt-Requiring Tests

These should be planned now and executed later if human input or a longer
interactive pass is useful:

- UI usability pass for dashboard editing workflow
- browser-level responsive review across desktop and narrow widths
- manual product review of information architecture and copy
- exploratory testing of admin terminal commands and operator ergonomics
- user preference review for default dashboard layouts and widget presets
- manual operator review of identity-management UX and copy

## Update Rule

Every major workstep in this session should update this file and `codex.md` so
the next fresh-context session can resume with minimal reconstruction.
