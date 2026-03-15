# Infra Scripts

This directory contains small, well-documented operational scripts such as
migration runners, manifest validators, and smoke-test helpers.

- `localhost_smoke.sh`: starts an isolated localhost API, worker, and
  scheduler stack on loopback, verifies scheduled and manual runs, checks the
  artifact API, exercises the admin terminal, and proves the CLI path. It now
  authenticates write-path requests with the configured admin token.
- `compose_smoke.sh`: boots the packaged Docker Compose stack, waits for
  migration and service health, validates the built web service, and proves the
  API, worker, scheduler, analytics, quality, artifacts, and remote CLI paths.
  It now authenticates manual run triggers with the configured admin token.
- `benchmark_suite.sh`: runs the repo-owned platform benchmark command against
  a running stack and writes a timestamped JSON report for future latency and
  regression tracking.
- `backup_snapshot.sh`: creates and verifies a first-party recovery bundle
  using `platformctl backup`, giving operators a repeatable backup path
  outside the broader smoke workflows.
- `restore_drill.sh`: runs the real restore command into an isolated temporary
  runtime root so operators can validate restore behavior without touching live
  state.
- `restore_e2e.sh`: creates a real bundle from an isolated smoke runtime,
  restores that bundle into a second isolated runtime root, and boots the API
  against the restored state.
