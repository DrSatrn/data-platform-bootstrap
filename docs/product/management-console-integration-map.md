# Management Console Integration Map

This document explains how the additive management-console staging pieces can be
wired together once the hot implementation threads settle.

Status note:

- the Management route now implements the first real integration pass using
  these staged pieces
- this file is still useful as a map of how the modules fit together, but it is
  no longer purely hypothetical

## Goal

The eventual in-app management surface should not be assembled as isolated
widgets.

It should read as one operator workflow:

1. inspect platform posture
2. pick or launch an operational action
3. observe tool and session output
4. inspect artifacts
5. follow the next recommended step

## Current Additive Building Blocks

The repo now has unwired staging pieces for:

- workbench command discovery
- control-plane overview
- external-tool run inspection
- terminal session transcripts
- post-command follow-up planning

The new composite preview file:

- `web/src/features/management/console/ManagementConsolePreview.tsx`

is not the final page. It is the merge sketch showing how these building blocks
can live together.

## Suggested Wiring Order

When Model 1 is ready to do the final integration pass, the safest order is:

1. choose the management route/page shell
2. wire the control-plane overview and workbench first
3. wire external-tool inspection and terminal sessions second
4. wire follow-up planning last once artifact and runbook behavior are stable

## Why This Order

The first two surfaces are mostly read-oriented.

The latter two depend more heavily on:

- artifact shape
- external tool event shape
- final operator workflows

That makes them more sensitive to ongoing backend changes.

## Guardrails For The Final Wiring Pass

- prefer using the existing staged modules over rewriting them in place
- keep role visibility explicit
- preserve artifact paths and runbook links in the UI
- ensure the terminal remains guided, not arbitrary shell access

## Success Criteria

The final integrated management console should let an operator:

- understand platform health quickly
- inspect external-tool activity clearly
- follow a terminal session without leaving the app
- know what artifact or runbook to open next
