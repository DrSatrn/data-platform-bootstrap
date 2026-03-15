# Operator Evidence Blueprint

This blueprint covers the evidence layer for the future management console.

## Purpose

After an operator runs a command or inspects an external tool run, the product
should make the important evidence obvious.

That evidence usually includes:

- stderr and stdout logs
- declared output artifacts
- verification reports
- the artifact most worth opening first

## Additive Draft Pieces

New unwired files created in this thread:

- `web/src/features/management/evidence/evidenceBoard.ts`
- `web/src/features/management/evidence/evidenceBoard.test.ts`
- `web/src/features/management/evidence/EvidenceBoard.tsx`
- `web/src/features/management/runbooks/runbookDock.ts`
- `web/src/features/management/runbooks/runbookDock.test.ts`
- `web/src/features/management/runbooks/RunbookDock.tsx`

## Product Direction

The terminal transcript should explain what happened.

The evidence board should explain what to inspect.

The runbook dock should explain what guidance is closest to the current issue.

Together they reduce the common operator failure mode of:

1. command failed
2. logs exist somewhere
3. artifact exists somewhere else
4. runbook exists in another tab

## Eventual Wiring

The final console should let an operator move through:

1. inspect session
2. inspect top evidence
3. open linked runbook
4. decide on retry or escalation

The additive modules created here are staging pieces for that path, not final
page wiring.
