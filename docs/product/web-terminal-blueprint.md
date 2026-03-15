# Web Terminal Blueprint

This document describes the additive blueprint for an in-app web terminal that
operates alongside the broader management GUI.

## Why This Exists

The platform needs two operator surfaces at once:

- a full GUI for inspecting everything
- a terminal-grade workflow for controlled operational actions

The web terminal should not be a novelty shell. It should be the operator
surface for:

- restores and recovery drills
- external tool runs like dbt
- diagnostics and queue inspection
- artifact and log follow-up

## Product Direction

The GUI should answer:

- what is happening
- what is degraded
- what needs attention

The web terminal should answer:

- what exact command or action should I run
- what happened when I ran it
- what logs and artifacts came out of it
- what runbook should I read next

## Session Model

The additive frontend session model created in this thread supports:

- explicit command entries
- stdout and stderr transcript lines
- session status transitions
- pinned artifacts promoted from a run
- recommended runbooks tied to a session

This is the right shape for the eventual product because an operator often
needs to move from:

1. command
2. logs
3. artifacts
4. runbook
5. follow-up action

without context switching across tools.

## UI Building Blocks Added

New unwired staging pieces:

- `web/src/features/management/terminal/sessionModel.ts`
- `web/src/features/management/terminal/mockSessions.ts`
- `web/src/features/management/terminal/TerminalTranscript.tsx`
- `web/src/features/management/terminal/OperatorSessionDeck.tsx`

These are intentionally additive and can be wired later once the main app
navigation settles.

## Recommended Eventual Integration

The management GUI should eventually have three coordinated panels:

1. platform overview
2. terminal sessions
3. artifact and runbook follow-up

The operator should be able to:

- launch a guided command from a GUI card
- see the command appear in a terminal transcript
- inspect emitted artifacts immediately
- jump directly to the matching runbook

## Guardrails

The in-app terminal should not become a raw unrestricted shell.

It should prefer:

- templated commands
- role-aware actions
- artifact capture by default
- explicit session history
- runbook linking

This keeps it aligned with the platform’s auditability and control-plane goals.
