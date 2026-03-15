// Package config parses environment variables into a strongly typed runtime
// configuration. The loader favors explicit defaults and descriptive failures
// so startup errors are easy to diagnose.
package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Settings contains the runtime configuration shared by the platform
// processes. Fields are grouped by operational concern rather than source.
type Settings struct {
	Environment             string
	HTTPAddr                string
	WebAddr                 string
	APIBaseURL              string
	LogLevel                string
	PostgresDSN             string
	DuckDBPath              string
	DataRoot                string
	ArtifactRoot            string
	ManifestRoot            string
	DashboardRoot           string
	SQLRoot                 string
	PythonTaskRoot          string
	PythonBinary            string
	DBTBinary               string
	DLTBinary               string
	PySparkBinary           string
	ExternalToolRoot        string
	SampleDataRoot          string
	MigrationsRoot          string
	AdminToken              string
	AccessTokens            string
	RunFailureWebhookURLs   []string
	AssetWarningWebhookURLs []string
	SchedulerTick           time.Duration
	WorkerPoll              time.Duration
	ExternalToolTimeout     time.Duration
	JobRetryBaseDelay       time.Duration
	AlertWebhookTimeout     time.Duration
	MaxConcurrentJobs       int
}

// Load reads environment variables and applies safe local-first defaults.
func Load() (Settings, error) {
	if err := loadEnvironmentFiles(); err != nil {
		return Settings{}, err
	}

	schedulerTick, err := durationFromEnv("PLATFORM_SCHEDULER_TICK", 15*time.Second)
	if err != nil {
		return Settings{}, err
	}

	workerPoll, err := durationFromEnv("PLATFORM_WORKER_POLL", 5*time.Second)
	if err != nil {
		return Settings{}, err
	}

	externalToolTimeout, err := durationFromEnv("PLATFORM_EXTERNAL_TOOL_TIMEOUT_DEFAULT", 5*time.Minute)
	if err != nil {
		return Settings{}, err
	}
	jobRetryBaseDelay, err := durationFromEnv("PLATFORM_JOB_RETRY_BASE_DELAY", 250*time.Millisecond)
	if err != nil {
		return Settings{}, err
	}

	alertWebhookTimeout, err := durationFromEnv("PLATFORM_ALERT_WEBHOOK_TIMEOUT", 5*time.Second)
	if err != nil {
		return Settings{}, err
	}

	maxJobs, err := intFromEnv("PLATFORM_MAX_CONCURRENT_JOBS", 4)
	if err != nil {
		return Settings{}, err
	}

	return Settings{
		Environment:             envOrDefault("PLATFORM_ENV", "development"),
		HTTPAddr:                envOrDefault("PLATFORM_HTTP_ADDR", ":8080"),
		WebAddr:                 envOrDefault("PLATFORM_WEB_ADDR", ":3000"),
		APIBaseURL:              envOrDefault("PLATFORM_API_BASE_URL", "http://127.0.0.1:8080"),
		LogLevel:                envOrDefault("PLATFORM_LOG_LEVEL", "debug"),
		PostgresDSN:             envOrDefault("PLATFORM_POSTGRES_DSN", "postgres://platform_app:local_dev_password@localhost:5432/platform?sslmode=disable"),
		DuckDBPath:              envOrDefault("PLATFORM_DUCKDB_PATH", "../var/duckdb/platform.duckdb"),
		DataRoot:                envOrDefault("PLATFORM_DATA_ROOT", "../var/data"),
		ArtifactRoot:            envOrDefault("PLATFORM_ARTIFACT_ROOT", "../var/artifacts"),
		ManifestRoot:            envOrDefault("PLATFORM_MANIFEST_ROOT", "../packages/manifests"),
		DashboardRoot:           envOrDefault("PLATFORM_DASHBOARD_ROOT", "../packages/dashboards"),
		SQLRoot:                 envOrDefault("PLATFORM_SQL_ROOT", "../packages/sql"),
		PythonTaskRoot:          envOrDefault("PLATFORM_PYTHON_TASK_ROOT", "../packages/python"),
		PythonBinary:            envOrDefault("PLATFORM_PYTHON_BINARY", "python3"),
		DBTBinary:               envOrDefault("PLATFORM_DBT_BINARY", "dbt"),
		DLTBinary:               envOrDefault("PLATFORM_DLT_BINARY", "dlt"),
		PySparkBinary:           envOrDefault("PLATFORM_PYSPARK_BINARY", "pyspark"),
		ExternalToolRoot:        envOrDefault("PLATFORM_EXTERNAL_TOOL_ROOT", ".."),
		SampleDataRoot:          envOrDefault("PLATFORM_SAMPLE_DATA_ROOT", "../packages/sample_data"),
		MigrationsRoot:          envOrDefault("PLATFORM_MIGRATIONS_ROOT", "../infra/migrations"),
		AdminToken:              envOrDefault("PLATFORM_ADMIN_TOKEN", ""),
		AccessTokens:            envOrDefault("PLATFORM_ACCESS_TOKENS", ""),
		RunFailureWebhookURLs:   csvListFromEnv("PLATFORM_ALERT_RUN_FAILURE_WEBHOOK_URLS"),
		AssetWarningWebhookURLs: csvListFromEnv("PLATFORM_ALERT_ASSET_WARNING_WEBHOOK_URLS"),
		SchedulerTick:           schedulerTick,
		WorkerPoll:              workerPoll,
		ExternalToolTimeout:     externalToolTimeout,
		JobRetryBaseDelay:       jobRetryBaseDelay,
		AlertWebhookTimeout:     alertWebhookTimeout,
		MaxConcurrentJobs:       maxJobs,
	}, nil
}

func loadEnvironmentFiles() error {
	originalEnv := map[string]bool{}
	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) == 2 {
			originalEnv[parts[0]] = true
		}
	}

	candidates := []string{}
	if explicit := strings.TrimSpace(os.Getenv("PLATFORM_ENV_FILE")); explicit != "" {
		candidates = append(candidates, explicit)
	} else {
		candidates = append(candidates, ".env", ".env.local", filepath.Join("..", ".env"), filepath.Join("..", ".env.local"))
	}

	seen := map[string]struct{}{}
	for _, candidate := range candidates {
		cleaned := filepath.Clean(candidate)
		if _, duplicate := seen[cleaned]; duplicate {
			continue
		}
		seen[cleaned] = struct{}{}

		if err := loadEnvironmentFile(cleaned, originalEnv); err != nil {
			return err
		}
	}
	return nil
}

func loadEnvironmentFile(path string, originalEnv map[string]bool) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("open env file %s: %w", path, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, found := strings.Cut(line, "=")
		if !found {
			return fmt.Errorf("env file %s line %d must be KEY=VALUE", path, lineNumber)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			return fmt.Errorf("env file %s line %d has an empty key", path, lineNumber)
		}
		if originalEnv[key] {
			continue
		}
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("set env %s from %s: %w", key, path, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan env file %s: %w", path, err)
	}
	return nil
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

func csvListFromEnv(name string) []string {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}
