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
- saved dashboard persistence at the API/store layer

The biggest remaining product gap is that operators still cannot shape the
reporting surface from the UI. The dashboard experience is readable, but it is
not yet an editable internal tool.

## Current Plan

1. Dashboard lifecycle hardening
   - Complete browser-based create/edit/delete/duplicate flows.
   - Mirror dashboard persistence into PostgreSQL while preserving the
     local-first file-backed source.
   - Add tests around reporting persistence and lifecycle behavior.

2. Richer reporting UX
   - Support KPI, table, line, and bar widgets without third-party charting
     dependencies.
   - Keep live widget hydration and preview during editing.
   - Improve the saved-dashboard experience so it behaves like a real internal
     product surface.

3. Metadata trust surfaces
   - Keep freshness status attached to catalog assets.
   - Surface stale and missing artifacts clearly in Datasets and System.
   - Add backend tests so freshness behavior remains reliable across refactors.

4. Next likely follow-on work
   - Dashboard sharing/preset workflows, more report-level controls, and
     deeper catalog or dataset drill-down pages.

## Latest Completed Workstep

- Added browser-based dashboard creation, duplication, editing, deletion, and
  widget reordering flows.
- Added first-party line and bar chart widgets alongside KPI and table widgets.
- Mirrored dashboard persistence into PostgreSQL via the reporting store
  boundary.
- Added backend freshness tests and surfaced freshness status in the datasets
  and system views.
- Turned the Datasets page into a catalog/detail workbench with owner, source,
  quality, docs, and column metadata inspection.

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
