package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadReadsDotEnvFromWorkingDirectory(t *testing.T) {
	tempDir := t.TempDir()
	restoreDir := mustChdir(t, tempDir)
	defer restoreDir()

	mustWriteFile(t, filepath.Join(tempDir, ".env"), "PLATFORM_HTTP_ADDR=127.0.0.1:19090\nPLATFORM_ADMIN_TOKEN=dotenv-admin\n")
	restoreEnv := unsetAndRestoreEnv(t, "PLATFORM_ENV_FILE", "PLATFORM_HTTP_ADDR", "PLATFORM_ADMIN_TOKEN")
	defer restoreEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.HTTPAddr != "127.0.0.1:19090" {
		t.Fatalf("expected HTTPAddr from .env, got %q", cfg.HTTPAddr)
	}
	if cfg.AdminToken != "dotenv-admin" {
		t.Fatalf("expected AdminToken from .env, got %q", cfg.AdminToken)
	}
}

func TestLoadPrefersExistingEnvironmentOverDotEnv(t *testing.T) {
	tempDir := t.TempDir()
	restoreDir := mustChdir(t, tempDir)
	defer restoreDir()

	mustWriteFile(t, filepath.Join(tempDir, ".env"), "PLATFORM_HTTP_ADDR=127.0.0.1:19090\n")
	restoreEnv := unsetAndRestoreEnv(t, "PLATFORM_ENV_FILE", "PLATFORM_HTTP_ADDR")
	defer restoreEnv()
	if err := os.Setenv("PLATFORM_HTTP_ADDR", ":8088"); err != nil {
		t.Fatalf("set PLATFORM_HTTP_ADDR: %v", err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.HTTPAddr != ":8088" {
		t.Fatalf("expected existing environment to win, got %q", cfg.HTTPAddr)
	}
}

func TestLoadReturnsErrorForMalformedEnvLine(t *testing.T) {
	tempDir := t.TempDir()
	restoreDir := mustChdir(t, tempDir)
	defer restoreDir()

	mustWriteFile(t, filepath.Join(tempDir, ".env"), "PLATFORM_HTTP_ADDR=127.0.0.1:19090\nBROKEN_LINE\n")
	restoreEnv := unsetAndRestoreEnv(t, "PLATFORM_ENV_FILE", "PLATFORM_HTTP_ADDR")
	defer restoreEnv()

	if _, err := Load(); err == nil {
		t.Fatal("expected malformed .env file to fail")
	}
}

func mustChdir(t *testing.T, dir string) func() {
	t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to %s: %v", dir, err)
	}
	return func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore cwd: %v", err)
		}
	}
}

func mustWriteFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func unsetAndRestoreEnv(t *testing.T, keys ...string) func() {
	t.Helper()

	original := make(map[string]*string, len(keys))
	for _, key := range keys {
		value, found := os.LookupEnv(key)
		if found {
			copyValue := value
			original[key] = &copyValue
		} else {
			original[key] = nil
		}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}

	return func() {
		for _, key := range keys {
			if original[key] == nil {
				if err := os.Unsetenv(key); err != nil {
					t.Fatalf("restore unset %s: %v", key, err)
				}
				continue
			}
			if err := os.Setenv(key, *original[key]); err != nil {
				t.Fatalf("restore %s: %v", key, err)
			}
		}
	}
}
