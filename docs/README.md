# Documentation Guide

This is the map of the documentation set.

If you are new here, do not read the folders in alphabetical order. Read based
on what you are trying to do.

## I Want To Get This Running On My Machine

- [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md)
  One clear first-run path. Start here.
- [bootstrap.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/bootstrap.md)
  Run the packaged Docker Compose stack.
- [localhost-e2e.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/localhost-e2e.md)
  Run the API, worker, scheduler, and web app directly on your machine.
- [local-host-run.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/local-host-run.md)
  A compatibility pointer that sends you to the real host-run guide.
- [config-reality.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/config-reality.md)
  Understand which env files and defaults apply in each runtime mode.
- [deployment.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/deployment.md)
  Deploy the platform for a small on-prem team.

## I Want To Understand How The Platform Works

- [system-overview.md](/Users/streanor/Documents/Playground/data-platform/docs/architecture/system-overview.md)
  Short explanation of the runtime pieces.
- [runtime-wiring.md](/Users/streanor/Documents/Playground/data-platform/docs/architecture/runtime-wiring.md)
  Practical map of the running services, ports, and state.
- [service-boundaries.md](/Users/streanor/Documents/Playground/data-platform/docs/architecture/service-boundaries.md)
  Explains which subsystem owns which responsibility.
- [data-model.md](/Users/streanor/Documents/Playground/data-platform/docs/architecture/data-model.md)
  High-level domain entities and concepts.
- [trace-one-pipeline.md](/Users/streanor/Documents/Playground/data-platform/docs/tutorials/trace-one-pipeline.md)
  Shorter walkthrough of one real pipeline.
- [trace-one-pipeline-complete.md](/Users/streanor/Documents/Playground/data-platform/docs/tutorials/trace-one-pipeline-complete.md)
  Fuller walkthrough from manifest to UI.

## I Want To Test Or Verify It Works

- [uat-checklist.md](/Users/streanor/Documents/Playground/data-platform/uat-checklist.md)
  Manual user acceptance checklist.
- [benchmarking.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/benchmarking.md)
  Run the benchmark gate and understand the output.
- [backups.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/backups.md)
  Verify backup and restore behavior.

## I Want To Operate And Maintain It

- [operator-manual.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/operator-manual.md)
  Main operator handbook.
- [deployment.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/deployment.md)
  On-prem deployment and capacity guidance.
- [access-matrix.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/access-matrix.md)
  Which role should be able to do what.
- [release-checklist.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/release-checklist.md)
  Pre-release, post-release, and rollback discipline.
- [optional-external-tools.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/optional-external-tools.md)
  How dbt-style external tools fit into the platform.
- [external-tool-troubleshooting.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/external-tool-troubleshooting.md)
  Troubleshoot external-tool runs.
- [dbt-operator-checklist.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/dbt-operator-checklist.md)
  Quick operator checklist for dbt-backed runs.

## I Want To Contribute Code Or Docs

- [making-changes.md](/Users/streanor/Documents/Playground/data-platform/docs/tutorials/making-changes.md)
  Where to make changes and how to think about ownership.
- [contributing.md](/Users/streanor/Documents/Playground/data-platform/contributing.md)
  Full contributor guide.
- [README.md](/Users/streanor/Documents/Playground/data-platform/README.md)
  Product overview and repo-level orientation.

## I Want To Understand Technical Decisions

- [docs/decisions/README.md](/Users/streanor/Documents/Playground/data-platform/docs/decisions/README.md)
  Index of architecture decision records.
- [0001-modular-monolith.md](/Users/streanor/Documents/Playground/data-platform/docs/decisions/0001-modular-monolith.md)
  Why the backend is a modular monolith.

## I Want Product And Design Context

These are internal design notes, not first-stop user docs.

- [docs/product/README.md](/Users/streanor/Documents/Playground/data-platform/docs/product/README.md)
  Explains what the product docs are for.
- [management-console-blueprint.md](/Users/streanor/Documents/Playground/data-platform/docs/product/management-console-blueprint.md)
  Intent for the management surface.
- [management-console-integration-map.md](/Users/streanor/Documents/Playground/data-platform/docs/product/management-console-integration-map.md)
  How the management modules fit together.
- [operator-evidence-blueprint.md](/Users/streanor/Documents/Playground/data-platform/docs/product/operator-evidence-blueprint.md)
  Why evidence and artifact surfacing matters.
- [operator-followup-blueprint.md](/Users/streanor/Documents/Playground/data-platform/docs/product/operator-followup-blueprint.md)
  Post-command follow-up workflow ideas.
- [opsview-ui-bridge.md](/Users/streanor/Documents/Playground/data-platform/docs/product/opsview-ui-bridge.md)
  Bridge between backend opsview read models and the UI.
- [web-terminal-blueprint.md](/Users/streanor/Documents/Playground/data-platform/docs/product/web-terminal-blueprint.md)
  Intent for the guided in-app terminal.

## I Want Stable Reference Material

- [docs/reference/README.md](/Users/streanor/Documents/Playground/data-platform/docs/reference/README.md)
  Reference section overview.
- [external-tool-jobs.md](/Users/streanor/Documents/Playground/data-platform/docs/reference/external-tool-jobs.md)
  Contract for `external_tool` jobs.
- [operator-command-taxonomy.md](/Users/streanor/Documents/Playground/data-platform/docs/reference/operator-command-taxonomy.md)
  Command families and operator intent.
- [opsview-read-models.md](/Users/streanor/Documents/Playground/data-platform/docs/reference/opsview-read-models.md)
  Backend operator read-model notes.
- [opsexport-bundles.md](/Users/streanor/Documents/Playground/data-platform/docs/reference/opsexport-bundles.md)
  Export-bundle notes for opsview snapshots.
- [management-console-demo-assets.md](/Users/streanor/Documents/Playground/data-platform/docs/reference/management-console-demo-assets.md)
  Demo payload notes for management-console work.

## Folder Index

### Architecture

- [docs/architecture/README.md](/Users/streanor/Documents/Playground/data-platform/docs/architecture/README.md)
  Reading order for the architecture docs.
- [system-overview.md](/Users/streanor/Documents/Playground/data-platform/docs/architecture/system-overview.md)
  Short system summary.
- [service-boundaries.md](/Users/streanor/Documents/Playground/data-platform/docs/architecture/service-boundaries.md)
  Responsibility boundaries.
- [runtime-wiring.md](/Users/streanor/Documents/Playground/data-platform/docs/architecture/runtime-wiring.md)
  Runtime process and state map.
- [data-model.md](/Users/streanor/Documents/Playground/data-platform/docs/architecture/data-model.md)
  Domain model summary.

### Decisions

- [docs/decisions/README.md](/Users/streanor/Documents/Playground/data-platform/docs/decisions/README.md)
  ADR index.
- [0001-modular-monolith.md](/Users/streanor/Documents/Playground/data-platform/docs/decisions/0001-modular-monolith.md)
  First ADR.

### Runbooks

- [docs/runbooks/README.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/README.md)
  Runbook section overview.
- [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md)
  Canonical first-run doc.
- [bootstrap.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/bootstrap.md)
  Docker Compose path.
- [localhost-e2e.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/localhost-e2e.md)
  Host-run end-to-end path.
- [local-host-run.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/local-host-run.md)
  Compatibility pointer.
- [operator-manual.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/operator-manual.md)
  Main operator handbook.
- [backups.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/backups.md)
  Backup and restore.
- [benchmarking.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/benchmarking.md)
  Benchmark gate.
- [config-reality.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/config-reality.md)
  Config source-of-truth.
- [access-matrix.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/access-matrix.md)
  Role matrix.
- [optional-external-tools.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/optional-external-tools.md)
  External-tool support.
- [external-tool-troubleshooting.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/external-tool-troubleshooting.md)
  External-tool troubleshooting.
- [dbt-operator-checklist.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/dbt-operator-checklist.md)
  DBT operator checklist.

### Tutorials

- [docs/tutorials/README.md](/Users/streanor/Documents/Playground/data-platform/docs/tutorials/README.md)
  Tutorial section overview.
- [making-changes.md](/Users/streanor/Documents/Playground/data-platform/docs/tutorials/making-changes.md)
  Contributor tutorial.
- [trace-one-pipeline.md](/Users/streanor/Documents/Playground/data-platform/docs/tutorials/trace-one-pipeline.md)
  Shorter pipeline walkthrough.
- [trace-one-pipeline-complete.md](/Users/streanor/Documents/Playground/data-platform/docs/tutorials/trace-one-pipeline-complete.md)
  Longer pipeline walkthrough.
