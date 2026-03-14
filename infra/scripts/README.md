# Infra Scripts

This directory contains small, well-documented operational scripts such as
migration runners, manifest validators, and smoke-test helpers.

- `localhost_smoke.sh`: starts an isolated localhost API, worker, and
  scheduler stack on loopback, verifies scheduled and manual runs, checks the
  artifact API, exercises the admin terminal, and proves the CLI path.
