# Ops Export Bundles

`backend/internal/opsexport/` is the additive backend-only export layer for
`opsview` snapshots.

Current scope:

- build stable export bundles from `opsview.RunOperatorSnapshot` values
- serialize those bundles as indented JSON
- preserve a compact rollup plus ordered snapshots for demos, review, and
  future API publication work

Important boundary:

- this package is pure and unwired
- it does not own handlers, runtime wiring, CLI integration, or frontend work
- golden JSON in package `testdata/` is additive coverage, not a published API
  guarantee by itself

Primary files:

- [models.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/opsexport/models.go)
- [bundle.go](/Users/streanor/Documents/Playground/data-platform/backend/internal/opsexport/bundle.go)
