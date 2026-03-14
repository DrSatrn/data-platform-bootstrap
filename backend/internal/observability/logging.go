// Package observability provides runtime diagnostics such as logging and health
// endpoints. This file focuses on cheap, structured logging built on the Go
// standard library so the local runtime stays lightweight.
package observability

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/streanor/data-platform/backend/internal/config"
)

// NewLogger creates a structured JSON logger so logs remain machine-readable
// and consistent across local and containerized environments.
func NewLogger(level string, service *Service) *slog.Logger {
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

	base := slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: slogLevel},
	)
	if service == nil {
		return slog.New(base)
	}

	return slog.New(&recordingHandler{
		next:    base,
		service: service,
	})
}

// HealthHandler returns a simple but useful readiness endpoint that reports
// static configuration context alongside status information.
func HealthHandler(cfg config.Settings) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":      "ok",
			"environment": cfg.Environment,
		})
	})
}

// RequestLoggingMiddleware emits a compact request log around each HTTP
// request. The timing here is intentionally cheap and avoids excessive
// allocations because this path sits on all API requests.
func RequestLoggingMiddleware(logger *slog.Logger, service *Service, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		recorder := &statusRecorder{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}
		next.ServeHTTP(recorder, r)
		duration := time.Since(start)
		if service != nil {
			service.RecordRequest(r.Method, r.URL.Path, recorder.statusCode, duration)
		}
		logger.Info(
			"http_request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.Int("status_code", recorder.statusCode),
			slog.Duration("duration", duration),
		)
	})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(mustMarshal(payload)))
}

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

type recordingHandler struct {
	next    slog.Handler
	service *Service
}

func (h *recordingHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.next.Enabled(ctx, level)
}

func (h *recordingHandler) Handle(ctx context.Context, record slog.Record) error {
	fields := map[string]string{}
	record.Attrs(func(attr slog.Attr) bool {
		fields[attr.Key] = attr.Value.String()
		return true
	})
	h.service.RecordLog(record.Level.String(), record.Message, fields)
	return h.next.Handle(ctx, record)
}

func (h *recordingHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &recordingHandler{
		next:    h.next.WithAttrs(attrs),
		service: h.service,
	}
}

func (h *recordingHandler) WithGroup(name string) slog.Handler {
	return &recordingHandler{
		next:    h.next.WithGroup(name),
		service: h.service,
	}
}
