# Infra Scripts

This directory contains small, well-documented operational scripts such as
migration runners, manifest validators, and smoke-test helpers.

- `localhost_smoke.sh`: starts an isolated localhost API, worker, and
  scheduler stack on loopback, verifies scheduled and manual runs, checks the
  artifact API, exercises the admin terminal, and proves the CLI path.
- `compose_smoke.sh`: boots the packaged Docker Compose stack, waits for
  migration and service health, validates the built web service, and proves the
  API, worker, scheduler, analytics, quality, artifacts, and remote CLI paths.
