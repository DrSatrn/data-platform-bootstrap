# Active Build Plan

This file is the short-horizon implementation plan for the current autonomous
work block. It exists to keep the next session focused without forcing future
agents to infer intent from a large git diff or chat transcript.

## Current Build Focus

Move the project from a strong vertical slice into a more fully fledged
internal platform product by prioritizing features that deepen user-facing
product behavior, not just lower-level plumbing.

## Why This Slice

The platform already has:

- real orchestration and execution
- DuckDB-backed curated analytics
- local and packaged smoke paths
- dashboard editing and persistence through the browser
- first-party reporting, diagnostics, and metadata surfaces

The next highest-leverage gap is platform trust and repeatable validation:
operators need deeper catalog intelligence, and the engineering loop needs
first-party benchmarks that can grow alongside smoke tests.

## Current Plan

1. Metadata intelligence
   - Derive richer catalog coverage and lineage summaries from repo manifests.
   - Surface documentation coverage, quality coverage, freshness, and lineage
     context in the operator UI.
   - Keep the metadata API useful without introducing a separate graph service.

2. Validation and benchmark foundation
   - Add a first-party benchmark command to `platformctl`.
   - Add a repo-owned benchmark script that emits timestamped JSON reports.
   - Build a stronger future E2E validation baseline alongside smoke tests.

3. Next likely follow-on work
   - Report sharing/preset workflows, deeper dataset drill-downs, and broader
     control-plane normalization in PostgreSQL.
   - Expand the benchmark suite with scheduled-run, artifact, and report-save
     latency budgets.

## Latest Completed Workstep

- Added derived catalog coverage, lineage, and trust summaries to the metadata
  API.
- Turned the Datasets page into a stronger metadata workbench with coverage,
  lineage, governance, and column documentation context.
- Added a first-party `platformctl benchmark` command plus
  `infra/scripts/benchmark_suite.sh`.
- Captured an initial benchmark baseline against the packaged stack under
  `var/benchmarks/`.

## Deferred / Prompt-Requiring Tests

These should be planned now and executed later if human input or a longer
interactive pass is useful:

- UI usability pass for dashboard editing workflow
- browser-level responsive review across desktop and narrow widths
- manual product review of information architecture and copy
- exploratory testing of admin terminal commands and operator ergonomics
- user preference review for default dashboard layouts and widget presets

## Update Rule

Every major workstep in this session should update this file and `codex.md` so
the next fresh-context session can resume with minimal reconstruction.
