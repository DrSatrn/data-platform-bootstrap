This feature directory now powers the live Management route in the main app.

The original staged modules were deliberately additive so parallel work could
build:

- command-taxonomy utilities
- operator workbench components
- inventory and attention summaries
- mock-backed layout drafts

without colliding with hot route or auth wiring.

Current reality:

- the Management page reuses these modules against real platform APIs
- `useManagementConsole.ts` is the integration hook that bridges live data into
  the staged components
- the backend `opsview` API now provides operator snapshots so the frontend
  does not need to re-group external-tool evidence from scratch

The mock-backed modules still exist as fallbacks and development scaffolding,
but the intended product path is now the live integrated surface.
