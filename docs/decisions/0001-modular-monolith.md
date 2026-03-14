# ADR 0001: Modular Monolith First

## Status

Accepted

## Context

The platform needs multiple distinct capabilities, but it must remain practical to run and study on a single developer machine.

## Decision

Use a modular-monolith backend with clear bounded contexts and multiple runtime entrypoints instead of a microservice-first decomposition.

## Consequences

- Local development remains fast and understandable.
- Boundaries still exist in code, making later extraction possible.
- We avoid unnecessary distributed-systems overhead in v1.
