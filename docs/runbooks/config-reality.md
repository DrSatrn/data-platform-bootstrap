# Configuration Reality

This runbook exists to remove ambiguity about how configuration is actually
loaded in the current repo.

## Why This Exists

Several platform workflows can look similar while using different configuration
sources:

- Docker Compose
- host-run Go binaries
- host-run web dev server
- smoke scripts

If we do not call those differences out explicitly, users are forced to infer
runtime behavior from a mix of `.env.example`, shell commands, and defaults.

## Current Configuration Sources

Backend processes resolve configuration in this order:

1. existing process environment variables
2. `PLATFORM_ENV_FILE` if explicitly set
3. `.env`
4. `.env.local`
5. `../.env`
6. `../.env.local`
7. built-in defaults in `backend/internal/config/config.go`

Important behavior:

- existing environment variables win over values loaded from env files
- env files are optional
- malformed env file lines fail startup

## Runtime Mode Differences

### Compose mode

Compose uses container-oriented paths such as:

- `/workspace/packages/...`
- `/var/lib/platform/...`

That is why `.env.example` is currently Compose-shaped.

### Host-run backend mode

When you run binaries from `backend/`, built-in defaults are repo-relative:

- `../packages/manifests`
- `../packages/sql`
- `../var/data`
- `../var/artifacts`

That means host-run local behavior can succeed even without exporting every
path manually, as long as you run from the expected working directory.

### Smoke scripts

The smoke scripts set their own explicit runtime environment and should be
treated as the safest executable reference for current expected behavior.

## Practical Guidance

Use this rule of thumb:

- use `make smoke` for the fastest verified path
- use `make bootstrap` for the packaged Compose path
- use manual `go run` commands only when debugging and be explicit about your
  working directory

## What To Verify Before Wiring This Into Canonical Docs

- confirm `backend/internal/config/config.go` still auto-loads env files
- confirm `.env.example` remains Compose-oriented
- confirm host-run defaults still resolve correctly from `backend/`
- confirm smoke scripts remain the most reliable proof path
