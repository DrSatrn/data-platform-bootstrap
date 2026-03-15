# Release Checklist

Use this checklist before calling a tagged build releasable.

## Pre-Release Verification

Run from the repo root unless noted otherwise:

1. `cd backend && go test ./...`
2. `cd backend && go run ./cmd/platformctl validate-manifests`
3. `cd web && npm run build`
4. `cd web && npm test`
5. `make smoke`
6. review [uat-checklist.md](/Users/streanor/Documents/Playground/data-platform/uat-checklist.md) and confirm the latest annotated pass is still representative

Do not cut a release if any of those fail.

## Changelog Format

Record each release with:

- version tag
- release date
- operator-visible changes
- schema or runtime compatibility notes
- known issues or deferred items

Prefer concise release notes over raw commit dumps.

## Version Tagging Strategy

Recommended format:

- `v1.0.0` for the first credible self-hosted release
- `v1.0.x` for fixes that preserve operational compatibility
- `v1.x.0` for minor feature releases that do not intentionally break upgrade expectations

Tag only after verification and backup steps complete.

## Migration Compatibility Check

Before release:

1. confirm migrations apply cleanly in a fresh environment
2. confirm the current runtime can start against the migrated schema
3. confirm backup restore still works with the release candidate state

If a release changes migration behavior, call that out explicitly in the
release notes.

## Backup Requirement

Always create and verify a backup before release:

1. `make backup`
2. confirm output includes both `backup bundle created:` and `backup bundle verified:`
3. retain at least one known-good pre-release bundle

## Smoke And Benchmark Gates

Required:

- `make smoke`

Strongly recommended:

- `make benchmark`
- `make restore-drill`

Investigate before release if benchmark output shows unexpected regressions or
the restore drill fails.

## Post-Release Validation

After the release is tagged and deployed:

1. verify `/healthz`
2. verify the web app loads
3. verify login through `/api/v1/session`
4. trigger one real pipeline run
5. confirm artifacts, catalog, dashboards, and the System page still load
6. confirm backup inventory is still visible

## Rollback Procedure

If the release is not acceptable:

1. stop the target runtime
2. restore the last known-good verified backup bundle
3. restart the runtime
4. repeat the post-restore checks from
   [backups.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/backups.md)
5. document the failed release and the rollback reason before retrying
