# Admin Package

This package implements the built-in management terminal for the platform. The
terminal exposes platform-specific commands over HTTP instead of arbitrary
shell access so the admin surface remains safe, auditable, and tailored to the
product.

Current command groups include:

- platform status, metrics, runs, pipelines, assets, and logs
- artifact inspection for completed runs
- backup bundle discovery and verification
- controlled pipeline triggering for admins
