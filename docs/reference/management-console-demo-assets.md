# Management Console Demo Assets

This reference doc tracks the additive demo assets created for the future
management-console integration work.

Location:

- `packages/demo/management_console/`

Current assets:

- `opsview_snapshot.json`
- `terminal_sessions.json`
- `runbook_dock.json`

Purpose:

- give future wiring work stable sample payloads
- support demos and visual integration without depending on live handlers
- avoid colliding with canonical runtime contracts while the frontend wiring
  pass is in flight

These files are intentionally additive and should be treated as demo payloads,
not source-of-truth API schemas.
