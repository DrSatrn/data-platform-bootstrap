// Package observability provides runtime diagnostics such as logging and health
// endpoints. This file focuses on cheap, structured logging built on the Go
// standard library so the local runtime stays lightweight.
package observability

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/streanor/data-platform/backend/internal/config"
)

// NewLogger creates a structured JSON logger so logs remain machine-readable
// and consistent across local and containerized environments.
func NewLogger(level string) *slog.Logger {
	var slogLevel slog.Level
	switch level {
	case "debug":
		slogLevel = slog.LevelDebug
	case "warn":
		slogLevel = slog.LevelWarn
	case "error":
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	return slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{Level: slogLevel},
		),
	)
}

// HealthHandler returns a simple but useful readiness endpoint that reports
// static configuration context alongside status information.
func HealthHandler(cfg config.Settings) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":      "ok",
			"environment": cfg.Environment,
			"http_addr":   cfg.HTTPAddr,
			"web_addr":    cfg.WebAddr,
		})
	})
}

// RequestLoggingMiddleware emits a compact request log around each HTTP
// request. The timing here is intentionally cheap and avoids excessive
// allocations because this path sits on all API requests.
func RequestLoggingMiddleware(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Info(
			"http_request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Duration("duration", time.Since(start)),
		)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(mustMarshal(payload)))
}
