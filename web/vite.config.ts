// This Vite config keeps the frontend build small and direct. The API proxy is
// aligned with the local platform runtime so the UI can develop against the Go
// API without extra browser configuration.
import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  plugins: [react()],
  server: {
    port: 3000,
    proxy: {
      "/api": {
        target: "http://127.0.0.1:8080",
        changeOrigin: true
      },
      "/healthz": {
        target: "http://127.0.0.1:8080",
        changeOrigin: true
      }
    }
  }
});
