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
   - Expand the benchmark suite with scheduled-run, artifact, report-save,
     queue-depth, scheduler-latency, and concurrency budgets.
   - Tighten verification around concurrent operator workflows and packaged
     upgrade paths.
   - Revisit richer report sharing/export once the benchmark contract is
     strong enough to act as a release gate.

## Latest Completed Workstep

- Completed the frontend completion and wiring pass for the staged management
  modules.
- Added a real `/api/v1/opsview` backend seam so the frontend can consume
  backend-owned operator summaries instead of rebuilding them from raw run
  events in the browser.
- Wired the staged management console, guided terminal, evidence board,
  runbook dock, follow-up board, and external-tool inspection modules into the
  live product through a new Management route.
- Kept exact artifact paths and failure classes visible in the operator flows.

## Next Likely Follow-On Work

- Reconcile and review the parallel dbt-runner work that other models are
  landing in the external-tool area.
- Decide whether the next platform pass should focus on broader connector and
  execution depth, richer governance features, or a stronger release/upgrade
  workflow on top of the new management surface.

## Deferred / Prompt-Requiring Tests

These should be planned now and executed later if human input or a longer
interactive pass is useful:

- UI usability pass for dashboard editing workflow
- browser-level responsive review across desktop and narrow widths
- manual product review of information architecture and copy
- exploratory testing of admin terminal commands and operator ergonomics
- user preference review for default dashboard layouts and widget presets
- manual operator review of identity-management UX and copy
- manual dataset-editor UX review and column-doc editing ergonomics

## Update Rule

Every major workstep in this session should update this file and `codex.md` so
the next fresh-context session can resume with minimal reconstruction.
