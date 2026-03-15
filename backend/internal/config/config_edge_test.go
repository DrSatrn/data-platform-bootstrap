// These tests cover malformed and awkward config inputs so local bootstrap
// behavior stays predictable even when operators misconfigure environment
// values.
package config

import "testing"

func TestLoadReturnsErrorForInvalidSchedulerTick(t *testing.T) {
	restoreEnv := unsetAndRestoreEnv(t, "PLATFORM_ENV_FILE", "PLATFORM_SCHEDULER_TICK")
	defer restoreEnv()
	mustSetEnv(t, "PLATFORM_SCHEDULER_TICK", "not-a-duration")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid scheduler tick to fail")
	}
}

func TestLoadReturnsErrorForInvalidWorkerPoll(t *testing.T) {
	restoreEnv := unsetAndRestoreEnv(t, "PLATFORM_ENV_FILE", "PLATFORM_WORKER_POLL")
	defer restoreEnv()
	mustSetEnv(t, "PLATFORM_WORKER_POLL", "soon")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid worker poll to fail")
	}
}

func TestLoadReturnsErrorForInvalidExternalToolTimeout(t *testing.T) {
	restoreEnv := unsetAndRestoreEnv(t, "PLATFORM_ENV_FILE", "PLATFORM_EXTERNAL_TOOL_TIMEOUT_DEFAULT")
	defer restoreEnv()
	mustSetEnv(t, "PLATFORM_EXTERNAL_TOOL_TIMEOUT_DEFAULT", "forever")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid external tool timeout to fail")
	}
}

func TestLoadReturnsErrorForInvalidMaxConcurrentJobs(t *testing.T) {
	restoreEnv := unsetAndRestoreEnv(t, "PLATFORM_ENV_FILE", "PLATFORM_MAX_CONCURRENT_JOBS")
	defer restoreEnv()
	mustSetEnv(t, "PLATFORM_MAX_CONCURRENT_JOBS", "lots")

	if _, err := Load(); err == nil {
		t.Fatal("expected invalid max concurrent jobs to fail")
	}
}

func TestLoadFallsBackToDefaultsForExplicitEmptyPaths(t *testing.T) {
	restoreEnv := unsetAndRestoreEnv(
		t,
		"PLATFORM_ENV_FILE",
		"PLATFORM_DATA_ROOT",
		"PLATFORM_ARTIFACT_ROOT",
		"PLATFORM_MANIFEST_ROOT",
		"PLATFORM_SQL_ROOT",
	)
	defer restoreEnv()
	mustSetEnv(t, "PLATFORM_DATA_ROOT", "")
	mustSetEnv(t, "PLATFORM_ARTIFACT_ROOT", "")
	mustSetEnv(t, "PLATFORM_MANIFEST_ROOT", "")
	mustSetEnv(t, "PLATFORM_SQL_ROOT", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.DataRoot == "" || cfg.ArtifactRoot == "" || cfg.ManifestRoot == "" || cfg.SQLRoot == "" {
		t.Fatalf("expected default paths to be restored, got %+v", cfg)
	}
}

func TestLoadAllowsBootstrapAndLegacyTokensTogether(t *testing.T) {
	restoreEnv := unsetAndRestoreEnv(t, "PLATFORM_ENV_FILE", "PLATFORM_ADMIN_TOKEN", "PLATFORM_ACCESS_TOKENS")
	defer restoreEnv()
	mustSetEnv(t, "PLATFORM_ADMIN_TOKEN", "bootstrap-token")
	mustSetEnv(t, "PLATFORM_ACCESS_TOKENS", "viewer-token:viewer:alice")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.AdminToken != "bootstrap-token" || cfg.AccessTokens != "viewer-token:viewer:alice" {
		t.Fatalf("expected auth settings to round-trip, got %+v", cfg)
	}
}
