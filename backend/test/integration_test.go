// This package-level integration test documents the minimum environment shape
// the backend expects during local bootstrap. It intentionally stays light: the
// goal is to prove config loading from the tracked example file works and to
// give future integration suites a safe starting point for broader bring-up
// tests.
package test

import (
	"path/filepath"
	"testing"

	"github.com/streanor/data-platform/backend/internal/config"
)

func TestLoadSettingsFromEnvExampleProvidesNonEmptyRuntimePaths(t *testing.T) {
	// The config loader auto-discovers .env files, so the test pins the tracked
	// example file explicitly to keep the integration contract deterministic.
	t.Setenv("PLATFORM_ENV_FILE", filepath.Clean("../.env.example"))

	settings, err := config.Load()
	if err != nil {
		t.Fatalf("load config from env example: %v", err)
	}

	required := map[string]string{
		"Environment":      settings.Environment,
		"HTTPAddr":         settings.HTTPAddr,
		"WebAddr":          settings.WebAddr,
		"APIBaseURL":       settings.APIBaseURL,
		"LogLevel":         settings.LogLevel,
		"PostgresDSN":      settings.PostgresDSN,
		"DuckDBPath":       settings.DuckDBPath,
		"DataRoot":         settings.DataRoot,
		"ArtifactRoot":     settings.ArtifactRoot,
		"ManifestRoot":     settings.ManifestRoot,
		"DashboardRoot":    settings.DashboardRoot,
		"SQLRoot":          settings.SQLRoot,
		"PythonTaskRoot":   settings.PythonTaskRoot,
		"PythonBinary":     settings.PythonBinary,
		"ExternalToolRoot": settings.ExternalToolRoot,
		"SampleDataRoot":   settings.SampleDataRoot,
		"MigrationsRoot":   settings.MigrationsRoot,
	}
	for name, value := range required {
		if value == "" {
			t.Fatalf("expected %s to be non-empty", name)
		}
	}
	if settings.SchedulerTick <= 0 {
		t.Fatalf("expected positive scheduler tick, got %s", settings.SchedulerTick)
	}
	if settings.WorkerPoll <= 0 {
		t.Fatalf("expected positive worker poll, got %s", settings.WorkerPoll)
	}
	if settings.MaxConcurrentJobs <= 0 {
		t.Fatalf("expected positive max concurrent jobs, got %d", settings.MaxConcurrentJobs)
	}
}
