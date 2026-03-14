# DB Package

This package contains PostgreSQL persistence and migration helpers for the platform.

Current responsibilities:

- opening PostgreSQL connections through the Go standard database APIs
- applying repo-managed SQL migrations through `platformctl migrate`
- serving as the primary control-plane repository for run snapshots, queue
  state, and artifact metadata when the required tables exist
- falling back cleanly to the filesystem-backed control plane when PostgreSQL is
  unavailable or migrations have not been applied

The runtime now prefers PostgreSQL when bootstrapped, while keeping the
filesystem-backed path available as a deliberate local-first fallback.
