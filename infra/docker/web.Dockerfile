# This Dockerfile packages the React frontend into a static bundle served by a
# tiny web server. It is intentionally conservative and ARM64-friendly.
FROM node:22-alpine AS build

WORKDIR /workspace/web
COPY web/package.json web/tsconfig.json web/vite.config.ts web/index.html ./
COPY web/src ./src
RUN npm install && npm run build

FROM nginx:1.27-alpine

COPY --from=build /workspace/web/dist /usr/share/nginx/html
