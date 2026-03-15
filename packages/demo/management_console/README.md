# Management Console Demo Assets

This directory contains additive, unwired demo assets for the future
management-console and web-terminal experience.

These files are not canonical runtime contracts.

They exist to give wiring and demo work a stable set of sample payloads without
touching routed pages or runtime handlers.

Current assets:

- `opsview_snapshot.json`
- `terminal_sessions.json`
- `terminal_session_detail.json`
- `followups.json`
- `runbook_dock.json`
- `evidence_board.json`
- `operator_signals.json`
- `management_console_bundle.json`

Suggested use:

- `terminal_sessions.json` for lightweight list or deck views
- `terminal_session_detail.json` for transcript-oriented views
- `followups.json` for next-step and action-board views
- `runbook_dock.json` for linked operator guidance
- `evidence_board.json` for artifact/evidence panes
- `opsview_snapshot.json` and `operator_signals.json` for external-tool overview
- `management_console_bundle.json` as a single entry point for demo scenarios
