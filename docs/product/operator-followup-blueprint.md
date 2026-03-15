# Operator Follow-up Blueprint

This blueprint covers the missing step after an operator runs a command in the
future web terminal: deciding what to do next.

## Why This Matters

A terminal transcript alone is not enough for platform operations.

After a command completes, the operator usually needs one of four things:

- the right artifact
- the next command
- the matching runbook
- a signal that the issue can wait

If the product does not surface this explicitly, operators fall back to memory
and guesswork.

## Additive Draft Pieces

New unwired staging files in this thread:

- `web/src/features/management/terminal/followupPlanner.ts`
- `web/src/features/management/terminal/followupPlanner.test.ts`
- `web/src/features/management/terminal/mockFollowups.ts`
- `web/src/features/management/terminal/ArtifactFollowupPanel.tsx`
- `web/src/features/management/terminal/NextStepBoard.tsx`

## Product Intent

The follow-up layer should answer:

- what is urgent right now
- what should be rerun
- what evidence should be preserved
- which runbook should be opened

That means a terminal session should not end at “exit code 2”.

It should produce a structured, operator-facing follow-up plan.

## Draft Rules Encoded In The Planner

Current additive rules are intentionally simple:

- failed pipeline sessions generate urgent failure review plus targeted rerun
- successful recovery sessions suggest promoting verification evidence
- sessions with a linked runbook always keep that runbook visible
- completed sessions without artifacts are flagged for weaker follow-up

These rules are meant as product scaffolding, not final policy.

## Eventual Integration Shape

The eventual management console should let an operator move through:

1. session transcript
2. suggested follow-up actions
3. artifact inspection
4. linked runbook
5. retry or escalation

This is the bridge between “web terminal” and “fully featured GUI for
management of everything.”
