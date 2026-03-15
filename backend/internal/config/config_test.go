// These tests cover the local env-file loading behavior so the documented
// host-run startup path stays trustworthy as the config package evolves.
package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsParentDotEnvForHostRun(t *testing.T) {
	repoRoot := t.TempDir()
	backendDir := filepath.Join(repoRoot, "backend")
	if err := os.MkdirAll(backendDir, 0o755); err != nil {
		t.Fatalf("mkdir backend dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, ".env"), []byte("PLATFORM_HTTP_ADDR=:19090\nPLATFORM_ADMIN_TOKEN=test-admin\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()
	if err := os.Chdir(backendDir); err != nil {
		t.Fatalf("chdir backend dir: %v", err)
	}

	unsetForTest(t, "PLATFORM_HTTP_ADDR")
	unsetForTest(t, "PLATFORM_ADMIN_TOKEN")
	unsetForTest(t, "PLATFORM_ENV_FILE")

	settings, err := Load()
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if settings.HTTPAddr != ":19090" {
		t.Fatalf("expected HTTP addr from .env, got %s", settings.HTTPAddr)
	}
	if settings.AdminToken != "test-admin" {
		t.Fatalf("expected admin token from .env, got %s", settings.AdminToken)
	}
}

func TestLoadRespectsExplicitEnvironmentOverDotEnv(t *testing.T) {
	repoRoot := t.TempDir()
	backendDir := filepath.Join(repoRoot, "backend")
	if err := os.MkdirAll(backendDir, 0o755); err != nil {
		t.Fatalf("mkdir backend dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(repoRoot, ".env"), []byte("PLATFORM_HTTP_ADDR=:19090\n"), 0o644); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer func() {
		_ = os.Chdir(originalWD)
	}()
	if err := os.Chdir(backendDir); err != nil {
		t.Fatalf("chdir backend dir: %v", err)
	}

	t.Setenv("PLATFORM_HTTP_ADDR", ":18080")
	unsetForTest(t, "PLATFORM_ENV_FILE")

	settings, err := Load()
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if settings.HTTPAddr != ":18080" {
		t.Fatalf("expected explicit env to win, got %s", settings.HTTPAddr)
	}
}

func unsetForTest(t *testing.T, key string) {
	t.Helper()
	original, hadOriginal := os.LookupEnv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset env %s: %v", key, err)
	}
	t.Cleanup(func() {
		if !hadOriginal {
			_ = os.Unsetenv(key)
			return
		}
		_ = os.Setenv(key, original)
	})
}
