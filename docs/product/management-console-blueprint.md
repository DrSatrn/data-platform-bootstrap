# Management Console Blueprint

This document is an additive blueprint for a future first-party management
surface inside the web app.

It is intentionally unwired and should be treated as guide-wire material for
parallel implementation work.

## Product Goal

The platform should expose two complementary operator surfaces inside the app:

1. a command-oriented in-app terminal for fast operational actions
2. a fully featured management GUI for inspection, governance, recovery, and
   lifecycle control

The terminal should accelerate experts.

The GUI should make the platform operable even when the user does not remember
every command or underlying subsystem.

## Why Both Surfaces Matter

The current admin terminal is useful, but a serious internal platform needs
more than command execution.

Operators also need:

- queue visibility
- service posture
- freshness attention
- dashboard inventory
- recovery and backup posture
- audit and governance context
- runbook shortcuts at the moment of action

## Proposed Management Surface Areas

### 1. Operator Workbench

Purpose:

- central place for command shortcuts
- command search
- recent operational actions
- runbook links

### 2. Control Plane Workspace

Purpose:

- show service health at a glance
- show current queue depth and active runs
- highlight datasets that need attention
- surface unresolved governance gaps

### 3. Governance Workspace

Purpose:

- missing docs
- missing quality checks
- ownership gaps
- PII visibility
- lineage completeness

### 4. Recovery Workspace

Purpose:

- recent backup bundles
- verification status
- restore drill history
- next recommended recovery action

### 5. Reporting Workspace

Purpose:

- saved dashboard inventory
- draft versus published definitions
- ownership and edit history
- widget health and stale references

## UX Principles

- terminal is not arbitrary shell access
- every action should point back to first-party runbooks
- role visibility should be obvious before the user clicks
- high-attention items should be visible without deep navigation
- the GUI should expose state, not just links to APIs

## Suggested Implementation Strategy

To minimize merge collisions, build this in phases:

1. additive utilities and mock-backed components
2. unwired routes/pages
3. integration behind existing auth/session context
4. final route wiring once hot files are stable

## Related Additive Frontend Modules

This blueprint is paired with the current unwired frontend staging modules:

- `web/src/features/management/terminal/OperatorWorkbench.tsx`
- `web/src/features/management/console/ControlPlaneWorkspace.tsx`
- `web/src/features/management/terminal/commandCatalog.ts`
- `web/src/features/management/inventory/assetAttention.ts`

These are not final product decisions. They are collision-safe building blocks
for later wiring.
