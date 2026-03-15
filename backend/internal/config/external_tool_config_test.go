package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadReadsExternalToolSettings(t *testing.T) {
	restoreEnv := unsetAndRestoreEnv(
		t,
		"PLATFORM_ENV_FILE",
		"PLATFORM_EXTERNAL_TOOL_ROOT",
		"PLATFORM_DBT_BINARY",
		"PLATFORM_DLT_BINARY",
		"PLATFORM_PYSPARK_BINARY",
		"PLATFORM_EXTERNAL_TOOL_TIMEOUT_DEFAULT",
	)
	defer restoreEnv()

	mustSetEnv(t, "PLATFORM_EXTERNAL_TOOL_ROOT", "/tmp/platform-repo")
	mustSetEnv(t, "PLATFORM_DBT_BINARY", "/opt/bin/dbt")
	mustSetEnv(t, "PLATFORM_DLT_BINARY", "/opt/bin/dlt")
	mustSetEnv(t, "PLATFORM_PYSPARK_BINARY", "/opt/bin/pyspark")
	mustSetEnv(t, "PLATFORM_EXTERNAL_TOOL_TIMEOUT_DEFAULT", "7m")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.ExternalToolRoot != "/tmp/platform-repo" {
		t.Fatalf("expected ExternalToolRoot /tmp/platform-repo, got %q", cfg.ExternalToolRoot)
	}
	if cfg.DBTBinary != "/opt/bin/dbt" {
		t.Fatalf("expected DBTBinary /opt/bin/dbt, got %q", cfg.DBTBinary)
	}
	if cfg.DLTBinary != "/opt/bin/dlt" {
		t.Fatalf("expected DLTBinary /opt/bin/dlt, got %q", cfg.DLTBinary)
	}
	if cfg.PySparkBinary != "/opt/bin/pyspark" {
		t.Fatalf("expected PySparkBinary /opt/bin/pyspark, got %q", cfg.PySparkBinary)
	}
	if cfg.ExternalToolTimeout != 7*time.Minute {
		t.Fatalf("expected ExternalToolTimeout 7m, got %s", cfg.ExternalToolTimeout)
	}
}

func mustSetEnv(t *testing.T, key, value string) {
	t.Helper()
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("set %s: %v", key, err)
	}
}
