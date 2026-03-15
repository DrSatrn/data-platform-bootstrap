# Audit Package

This package owns the platform's append-only audit trail for sensitive
operator-facing actions.

The current scope covers:

- manual pipeline triggers
- dashboard saves and deletes
- admin terminal command execution

The store is local-first with PostgreSQL mirroring when available so audit
history remains durable in the self-hosted path without giving up the simple
single-node development model.
