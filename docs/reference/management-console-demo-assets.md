# Management Console Demo Assets

This reference doc tracks the additive demo assets created for the future
management-console integration work.

Location:

- `packages/demo/management_console/`

Current assets:

- `opsview_snapshot.json`
- `terminal_sessions.json`
- `terminal_session_detail.json`
- `followups.json`
- `runbook_dock.json`
- `evidence_board.json`
- `operator_signals.json`
- `management_console_bundle.json`

Purpose:

- give future wiring work stable sample payloads
- support demos and visual integration without depending on live handlers
- avoid colliding with canonical runtime contracts while the frontend wiring
  pass is in flight

These files are intentionally additive and should be treated as demo payloads,
not source-of-truth API schemas.

Recommended mapping:

- `opsview_snapshot.json`: backend-style opsview summary payload
- `operator_signals.json`: derived operator signal cards for summary strips
- `terminal_sessions.json`: lightweight session list data
- `terminal_session_detail.json`: transcript-capable session detail data
- `followups.json`: follow-up board actions
- `runbook_dock.json`: runbook shortcuts and urgency
- `evidence_board.json`: operator evidence panel inputs
- `management_console_bundle.json`: top-level scenario manifest for demos

Current scenario:

- one failed dbt external-tool job requiring triage
- one successful recovery validation command with preserved evidence
