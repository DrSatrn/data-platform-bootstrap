# Authz Package

This package provides the platform's first real access-control boundary.

The current design is intentionally pragmatic and local-first:

- bearer tokens are configured through environment variables
- each token resolves to a subject and role
- roles map to coarse capabilities used by the UI and API handlers
- the existing admin token remains supported for backward compatibility

This is not yet a full identity system, but it is a meaningful step beyond a
single opaque admin token and gives the platform real write-path protection.
