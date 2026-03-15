# Model 1 — Security & Identity Hardening

**Priority:** HIGHEST — this is the most critical work for production readiness.  
**Owner:** Model 1 (strongest model)  
**Merge order:** 1st  

---

## Mission

Make the platform safe for a 3-person on-prem data team by hardening the identity layer, sanitizing API error responses, and closing session management gaps. This work is the single biggest blocker to internal production.

## Context

Read these first:
- `backend/internal/authz/service.go` — the entire identity/session layer
- `backend/internal/metadata/handler.go` — example of error response leakage
- `prompts/api-error-audit.md` — existing audit of error response issues
- `docs/runbooks/operator-manual.md` — operator-facing access model docs

## Tasks (In Priority Order)

### Task 1: Replace Password Hashing with bcrypt

**Current state:** `derivePasswordHash` in `authz/service.go` L451–459 uses 120,000 SHA-256 iterations. This is non-standard and not resistant to GPU-based cracking.

**Required change:**
- Replace `hashPassword` and `verifyPassword` with `golang.org/x/crypto/bcrypt`
- Use a cost factor of 12 (adjustable via config)
- Migration: existing stored hashes must still work during a transition period OR you must document that existing users will need password resets after the upgrade

**Files allowed:**
- `backend/internal/authz/service.go` — modify hashing functions
- `backend/go.mod` and `backend/go.sum` — add bcrypt dependency

**Files forbidden:**
- ❌ `backend/internal/app/runtime.go`
- ❌ `backend/internal/db/` (schema changes)
- ❌ Any frontend files

**Completion signal:** `cd backend && go test ./internal/authz/...` passes with bcrypt-based hashing.

### Task 2: Add Login Rate Limiting

**Current state:** `Login()` in `authz/service.go` accepts unlimited authentication attempts. No throttle anywhere.

**Required change:**
- Add an in-memory per-IP rate limiter (e.g., `golang.org/x/time/rate` or a simple sliding window)
- Return HTTP 429 after 5 failed attempts within 1 minute
- Add a corresponding test

**Files allowed:**
- `backend/internal/authz/service.go` — add rate limiter
- `backend/internal/authz/service_test.go` — add rate limit tests
- New file: `backend/internal/authz/ratelimit.go` if separating concerns

**Completion signal:** Rate limiting is testable and `go test ./internal/authz/...` passes.

### Task 3: Sanitize API Error Responses

**Current state:** Multiple handlers return raw `err.Error()` to API consumers. This can leak internal paths, database errors, and stack information.

**Required change:**
- Audit every handler file under `backend/internal/`
- Replace raw `err.Error()` returns with user-safe messages
- Log the full error server-side
- Use `prompts/api-error-audit.md` as the starting checklist

**Files allowed:**
- All `handler.go` files under `backend/internal/`
- `backend/internal/shared/` — add error wrapping utilities

**Files forbidden:**
- ❌ `backend/internal/app/runtime.go`

**Completion signal:** No handler returns raw `err.Error()` to the client. `go test ./...` passes.

### Task 4: Session Management Hardening

**Current state:**
- No session expiry cleanup (expired sessions accumulate in Postgres)
- No concurrent session limit per user
- `TouchSession` errors are silently ignored

**Required change:**
- Add a background goroutine or startup sweep that deletes sessions where `expires_at < now()`
- Limit active sessions per user to 5 (configurable)
- Log `TouchSession` errors instead of swallowing them

**Files allowed:**
- `backend/internal/authz/service.go`
- `backend/internal/db/identity_store.go`

**Completion signal:** Session cleanup is testable. `go test ./...` passes.

### Task 5: Document the Security Model

**Required change:**
- Create `docs/security.md` covering:
  - Authentication model (bootstrap + native sessions)
  - Password storage approach (after bcrypt migration)
  - Session lifecycle (creation, expiry, cleanup)
  - Rate limiting behavior
  - Role hierarchy and enforcement
  - Network binding defaults and LAN safety
  - Known limitations and future work

**Files allowed:**
- `docs/security.md` (NEW)

## Escalation Triggers

- If bcrypt migration requires schema changes → escalate to reviewer
- If rate limiting needs Postgres-backed state → document the trade-off, implement in-memory first
- If error sanitization requires changing API response contracts → document the breaking change

## Completion Note Format

After completing, leave `prompts/v2-model1-completion.md` containing:
1. Files changed
2. What is now verifiably true
3. Verification commands and results
4. What was explicitly NOT changed
5. Any escalation items
