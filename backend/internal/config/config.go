// Package config parses environment variables into a strongly typed runtime
// configuration. The loader favors explicit defaults and descriptive failures
// so startup errors are easy to diagnose.
package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Settings contains the runtime configuration shared by the platform
// processes. Fields are grouped by operational concern rather than source.
type Settings struct {
	Environment       string
	HTTPAddr          string
	WebAddr           string
	APIBaseURL        string
	LogLevel          string
	PostgresDSN       string
	DuckDBPath        string
	DataRoot          string
	ArtifactRoot      string
	ManifestRoot      string
	DashboardRoot     string
	SQLRoot           string
	SampleDataRoot    string
	MigrationsRoot    string
	AdminToken        string
	SchedulerTick     time.Duration
	WorkerPoll        time.Duration
	MaxConcurrentJobs int
}

// Load reads environment variables and applies safe local-first defaults.
func Load() (Settings, error) {
	schedulerTick, err := durationFromEnv("PLATFORM_SCHEDULER_TICK", 15*time.Second)
	if err != nil {
		return Settings{}, err
	}

	workerPoll, err := durationFromEnv("PLATFORM_WORKER_POLL", 5*time.Second)
	if err != nil {
		return Settings{}, err
	}

	maxJobs, err := intFromEnv("PLATFORM_MAX_CONCURRENT_JOBS", 4)
	if err != nil {
		return Settings{}, err
	}

	return Settings{
		Environment:       envOrDefault("PLATFORM_ENV", "development"),
		HTTPAddr:          envOrDefault("PLATFORM_HTTP_ADDR", ":8080"),
		WebAddr:           envOrDefault("PLATFORM_WEB_ADDR", ":3000"),
		APIBaseURL:        envOrDefault("PLATFORM_API_BASE_URL", "http://127.0.0.1:8080"),
		LogLevel:          envOrDefault("PLATFORM_LOG_LEVEL", "debug"),
		PostgresDSN:       envOrDefault("PLATFORM_POSTGRES_DSN", "postgres://platform_app:local_dev_password@localhost:5432/platform?sslmode=disable"),
		DuckDBPath:        envOrDefault("PLATFORM_DUCKDB_PATH", "../var/duckdb/platform.duckdb"),
		DataRoot:          envOrDefault("PLATFORM_DATA_ROOT", "../var/data"),
		ArtifactRoot:      envOrDefault("PLATFORM_ARTIFACT_ROOT", "../var/artifacts"),
		ManifestRoot:      envOrDefault("PLATFORM_MANIFEST_ROOT", "../packages/manifests"),
		DashboardRoot:     envOrDefault("PLATFORM_DASHBOARD_ROOT", "../packages/dashboards"),
		SQLRoot:           envOrDefault("PLATFORM_SQL_ROOT", "../packages/sql"),
		SampleDataRoot:    envOrDefault("PLATFORM_SAMPLE_DATA_ROOT", "../packages/sample_data"),
		MigrationsRoot:    envOrDefault("PLATFORM_MIGRATIONS_ROOT", "../infra/migrations"),
		AdminToken:        envOrDefault("PLATFORM_ADMIN_TOKEN", ""),
		SchedulerTick:     schedulerTick,
		WorkerPoll:        workerPoll,
		MaxConcurrentJobs: maxJobs,
	}, nil
}

func envOrDefault(name, fallback string) string {
	value := os.Getenv(name)
	if value == "" {
		return fallback
	}
	return value
}

func durationFromEnv(name string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid duration: %w", name, err)
	}

	return duration, nil
}

func intFromEnv(name string, fallback int) (int, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be a valid integer: %w", name, err)
	}

	return parsed, nil
}
