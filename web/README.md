# Web App

This directory contains the React and TypeScript frontend for the platform. The UI is designed to feel like a compact internal product for operators and analysts rather than a marketing site or generic dashboard shell.

## Frontend Priorities

- fast local startup and build behavior
- clear routing and feature ownership
- metadata-aware views
- operational status visibility
- chart-ready integration with curated analytics endpoints

## Runtime Modes

- Local development uses Vite for the quickest feedback loop.
- Packaged Compose deployment builds the frontend once and serves it through the
  repo's own `server.mjs` static host, which also proxies `/api` and `/healthz`
  requests to the Go API so the localhost service stack behaves like a real
  deployed product.
