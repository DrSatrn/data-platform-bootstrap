# DB Package

This package contains PostgreSQL persistence and migration helpers for the platform.

Current responsibilities:

- opening PostgreSQL connections through the Go standard database APIs
- applying repo-managed SQL migrations through `platformctl migrate`
- mirroring pipeline run snapshots into PostgreSQL when the required tables exist

The file-backed local control plane still remains the primary read/write source for localhost execution. PostgreSQL is currently a hardening mirror, not yet the primary orchestration repository.
