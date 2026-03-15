# On-Prem Deployment Guide

This runbook is the practical deployment path for a small on-prem team running
the platform on one managed machine.

## Target Team Shape

This guide assumes:

- 3 internal users
- one shared on-prem host
- local Docker or OrbStack packaging
- PostgreSQL and the web/API stack running on the same machine

## Hardware Baseline

Recommended starting point:

- CPU: 4 performance cores minimum, 8 preferred
- RAM: 16 GB minimum, 24 to 32 GB preferred if you expect larger DuckDB data
- Disk: 100 GB SSD minimum, 250 GB preferred for backups, artifacts, and growth
- Architecture: Apple Silicon or ARM64 Linux is the best-tested path today

Sizing guidance:

- DuckDB, run artifacts, and backup bundles all grow with retained history
- expect faster growth in `var/data`, `var/artifacts`, and `var/backups` than in
  PostgreSQL for this slice
- reserve at least 3x your active runtime footprint if you plan to keep
  multiple verified backup bundles on the same host

## Runtime Layout

The platform expects these major runtime surfaces:

- web UI on port `3000`
- API on port `8080`
- PostgreSQL inside the packaged stack
- local data roots for data files, artifacts, DuckDB, and backups

Recommended deployment posture:

- keep PostgreSQL unpublished outside the host
- expose only the web and API through a reverse proxy or LAN bind policy
- keep the default loopback-first behavior until you intentionally harden the
  network path

## Network Configuration

For a LAN-visible deployment:

1. bind the reverse proxy to the LAN interface
2. keep the Compose or host-run API behind that proxy
3. allow inbound traffic only from trusted internal subnets
4. do not expose PostgreSQL directly

Firewall guidance:

- allow `80` and `443` from trusted office/VPN ranges if using a reverse proxy
- block direct external access to `3000`, `8080`, and PostgreSQL unless you
  intentionally operate without a proxy inside a trusted LAN
- keep SSH restricted to operators

## TLS Termination

Preferred pattern:

- terminate TLS at a reverse proxy such as Caddy, Nginx, or Traefik
- proxy `/` to the web service
- proxy `/api/` to the API service

Operational recommendation:

- use internally trusted certificates for LAN use
- preserve `X-Forwarded-*` headers consistently
- keep the platform services themselves on private bindings behind the proxy

## Bootstrap Path

Start from:

1. [quickstart.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/quickstart.md)
2. [bootstrap.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/bootstrap.md)

Recommended first deployment sequence:

1. run `make doctor`
2. run `make bootstrap`
3. verify `http://127.0.0.1:8080/healthz`
4. create native users with the bootstrap admin token
5. sign in through `/api/v1/session`
6. run `make smoke`
7. run `make backup`

## User Provisioning For A 3-Person Team

Recommended initial roles:

- 1 admin for platform recovery, user management, backup verification, and the
  admin terminal
- 1 editor for day-to-day pipeline triggers, metadata edits, and dashboard
  changes
- 1 viewer for read-only reporting and catalog access

Provisioning pattern:

1. bootstrap with `PLATFORM_ADMIN_TOKEN`
2. create named native users through the System page or `POST /api/v1/admin/users`
3. stop using the bootstrap token for daily work
4. keep the bootstrap token only as a recovery path

## Backup Schedule

Minimum recommendation:

- daily `make backup`
- weekly restore drill with `make restore-drill`
- pre-release backup before any tagged rollout

Retention guidance:

- keep at least 7 daily bundles
- keep at least 4 weekly bundles
- keep one known-good pre-release bundle before schema or runtime changes

## Monitoring And Operations Rhythm

Check daily:

- `/healthz`
- System page source-of-truth summary
- scheduler freshness and queue summary
- recent failures and audit trail
- backup inventory freshness

Check weekly:

- `make smoke`
- `make benchmark`
- `make restore-drill`
- disk growth in `var/data`, `var/artifacts`, and `var/backups`

Investigate immediately if:

- runs remain queued unexpectedly
- scheduler freshness lags
- backups stop appearing
- audit or dashboard persistence diverges from the System page source-of-truth

## Data Directory Sizing

Watch these directories:

- `var/data`
- `var/artifacts`
- `var/backups`
- `var/duckdb`

Growth expectations:

- `var/data` grows with landed raw files, staged JSON, marts, metrics, quality,
  and profile outputs
- `var/artifacts` grows with every pipeline run
- `var/backups` grows with each retained bundle
- `var/duckdb` grows with analytical history and marts

For a 3-person internal team, start with alerts or manual checks when free disk
falls below 25%.

## OrbStack Notes For macOS

OrbStack-specific guidance:

- ensure Docker CLI access works before relying on `make bootstrap`
- keep the Docker socket permission model stable for operational scripts
- allocate enough RAM to OrbStack for PostgreSQL, DuckDB-backed work, and the
  web build container
- if Compose commands fail intermittently, verify OrbStack is healthy before
  debugging the platform itself

## Release Discipline

Before rollout:

1. run the release checklist in
   [release-checklist.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/release-checklist.md)
2. create a backup bundle
3. verify backend tests, frontend build/tests, manifest validation, and smoke
   checks

## If This Fails

Start with:

1. [operator-manual.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/operator-manual.md)
2. [backups.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/backups.md)
3. [benchmarking.md](/Users/streanor/Documents/Playground/data-platform/docs/runbooks/benchmarking.md)
