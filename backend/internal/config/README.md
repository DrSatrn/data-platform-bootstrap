# Config Package

This package parses runtime configuration from environment variables and
provides strongly typed settings to the rest of the backend. Defensive parsing
matters here because bad config should fail fast and clearly.

Host-run binaries now auto-load `.env` and `.env.local` files from the current
working directory and its parent directory unless a process environment value is
already set. This makes local `go run` and built-binary workflows easier to
follow without changing the Compose behavior, which still relies on explicit
container environment injection.

Optional external-tool settings are also parsed here so the Go control plane
can pass stable runtime defaults into bounded adapters:

- `PLATFORM_DBT_BINARY`
- `PLATFORM_DLT_BINARY`
- `PLATFORM_PYSPARK_BINARY`
- `PLATFORM_EXTERNAL_TOOL_ROOT`
- `PLATFORM_EXTERNAL_TOOL_TIMEOUT_DEFAULT`
