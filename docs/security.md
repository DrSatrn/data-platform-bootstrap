# Security Model

This document explains the current security posture of the platform for a
small on-prem team. It focuses on what is implemented today, how to operate it
safely, and what limitations still matter before broader deployment.

## Authentication Paths

The platform currently supports three authentication modes.

### Bootstrap admin token

Purpose:
- first-run access
- break-glass recovery
- emergency admin access when the native identity store is unavailable

How it works:
- set `PLATFORM_ADMIN_TOKEN`
- send `Authorization: Bearer <token>`

This token is effectively root access and should be reserved for recovery and
bootstrap workflows.

### Native users and sessions

Purpose:
- normal day-to-day operator access

How it works:
- administrators create users
- users log in through `POST /api/v1/session`
- the backend creates a PostgreSQL-backed session record and returns a bearer
  token

This is the intended operating model for a 3-person on-prem team.

### Legacy static access tokens

Purpose:
- compatibility bridge only

How it works:
- `PLATFORM_ACCESS_TOKENS=token:role:subject,...`

These still work, but they should not be the default operational model.

## Password Storage

New passwords are stored with bcrypt.

Current behavior:
- bcrypt is used for newly created passwords
- bcrypt is used for password resets
- default bcrypt cost is `12`
- cost can be adjusted with `PLATFORM_BCRYPT_COST`

Migration behavior:
- legacy SHA-based password records still verify during the transition period
- once a user resets a password, the stored record moves to bcrypt

Operational implication:
- existing users do not need an immediate forced reset to keep logging in
- password resets are still the cleanest path to converge older accounts onto
  the current scheme

## Session Lifecycle

Native sessions are stored in PostgreSQL.

Creation:
1. user logs in successfully
2. backend generates a random bearer token
3. backend stores only a SHA-256 token hash in `platform_sessions`
4. client keeps the bearer token locally

Expiry:
- default session TTL is `24h`

Cleanup:
- expired sessions are swept during auth-service startup
- expired sessions are also deleted during successful login flows

Touch behavior:
- valid sessions update `last_seen_at`
- touch failures are logged server-side instead of being silently ignored

Concurrency limit:
- active sessions per user are capped at `5` by default
- configurable with `PLATFORM_MAX_ACTIVE_SESSIONS`
- older sessions are trimmed when a new login would exceed the limit

## Login Rate Limiting

Failed login attempts are rate limited in memory.

Current behavior:
- `5` failed attempts
- within `1 minute`
- response after limit: HTTP `429`

Keying:
- `X-Forwarded-For` first
- otherwise request remote address

Important limitation:
- limiter state is in-memory only
- restarting the API clears the limiter state

For a small self-hosted team this is a meaningful improvement, but it is not a
replacement for edge protection at the reverse-proxy or firewall layer.

## Role Hierarchy

- `anonymous`
  - unauthenticated
  - limited to public health/session entry points

- `viewer`
  - read-only platform access
  - analytics, catalog, reports, system views, opsview, audit visibility

- `editor`
  - everything a viewer can do
  - pipeline triggers
  - dashboard writes
  - metadata annotation edits

- `admin`
  - everything an editor can do
  - user management
  - admin terminal access
  - high-privilege operational workflows

## Network Binding Defaults

There are two important runtime modes.

### Compose / packaged stack

The packaged deployment path is loopback-first by default, which is the safer
starting posture for local and pilot installs.

### Host-run binaries

Host-run defaults still use:
- `PLATFORM_HTTP_ADDR=:8080`
- `PLATFORM_WEB_ADDR=:3000`

That means host-run binaries listen on all interfaces unless overridden.

Recommended workstation-safe override:

```sh
export PLATFORM_HTTP_ADDR=127.0.0.1:8080
export PLATFORM_WEB_ADDR=127.0.0.1:3000
```

If you intentionally expose the platform on a LAN:
- put it behind a reverse proxy
- terminate TLS there
- restrict firewall ingress to the team subnet
- avoid relying on the bootstrap token for normal use

## Recommended Practice For A 3-Person Team

- keep one bootstrap admin token for recovery only
- create named native users for each team member
- minimize the number of `admin` users
- use `editor` for normal operators who need to trigger runs or edit metadata
- rotate any shared token immediately if it leaks into shell history, logs, or
  screenshots

## Known Limitations

- no external identity provider integration
- no MFA
- no team/group policy model yet
- rate limiting is in-memory only
- host-run defaults are broader than Compose defaults unless overridden
- session cleanup is startup/login driven, not a dedicated long-running janitor
- admin-terminal command failures may still contain detailed operator-visible
  backend output because that surface is intentionally privileged

## What To Check Regularly

Daily:
- recent audit events
- unusual failed-login activity
- unexpected admin-terminal usage

Weekly:
- active user list
- stale session posture
- whether the bootstrap token is being used outside recovery workflows
