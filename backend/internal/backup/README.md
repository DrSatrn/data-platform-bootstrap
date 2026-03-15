# Backup Package

This package owns first-party backup bundle creation and verification for the
local-first platform.

The goal is to make self-hosted recovery concrete and inspectable without
depending on external backup products or proprietary control-plane services.

Current responsibilities:

- build a portable `.tar.gz` bundle containing control-plane exports
- include local runtime state such as DuckDB, artifacts, manifests, and saved
  dashboards when present
- emit a machine-readable manifest with counts and checksummed entries
- verify a produced bundle before operators rely on it

This is intentionally a backup/export subsystem, not yet a full one-command
restore engine. The bundle format and verification path are the foundation for
later restore automation.
