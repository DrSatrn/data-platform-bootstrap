# Backend

This directory contains the Go backend for the platform. The code is organized as a modular monolith with clear bounded contexts so the system stays fast and easy to reason about while still modeling real internal-platform architecture.

## Structure

- `cmd/` contains runtime entrypoints.
- `internal/` contains product subsystems and shared implementation details.
- `pkg/` is reserved for carefully chosen public utility packages.
- `test/` contains broader integration-oriented test harnesses.

## Design Rules

- Keep orchestration state transitions explicit.
- Prefer standard library components unless a dependency materially improves clarity or reliability.
- Optimize for `darwin/arm64` and `linux/arm64`.
