# This Dockerfile builds the frontend once and then serves it with the repo's
# own lightweight Node-based static host. That keeps the packaged web runtime
# fast, ARM64-friendly, and independent of a dev-only Vite server.
FROM node:22-bookworm-slim AS build

WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci

COPY web/ ./
RUN npm run build

FROM node:22-bookworm-slim

WORKDIR /app
COPY --from=build /src/web/dist /app/dist
COPY web/server.mjs /app/server.mjs

ENV PORT=3000
ENV HOST=0.0.0.0
ENV PLATFORM_WEB_DIST_ROOT=/app/dist

CMD ["node", "/app/server.mjs"]
