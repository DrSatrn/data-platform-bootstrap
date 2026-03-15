# Config Package

This package parses runtime configuration from environment variables and
provides strongly typed settings to the rest of the backend. Defensive parsing
matters here because bad config should fail fast and clearly.

Host-run binaries now auto-load `.env` and `.env.local` files from the current
working directory and its parent directory unless a process environment value is
already set. This makes local `go run` and built-binary workflows easier to
follow without changing the Compose behavior, which still relies on explicit
container environment injection.
