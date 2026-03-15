package externaltools

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func defaultBinary(configured, fallback string) string {
	if strings.TrimSpace(configured) == "" {
		return fallback
	}
	return strings.TrimSpace(configured)
}

func resolveRepoRoot(requestRoot, settingsRoot string) string {
	root := strings.TrimSpace(requestRoot)
	if root == "" {
		root = strings.TrimSpace(settingsRoot)
	}
	if root == "" {
		root = "."
	}
	return filepath.Clean(root)
}

func resolveRepoRelativePath(repoRoot, ref, field string) (string, error) {
	if !isRepoRelativeRef(ref) {
		return "", fmt.Errorf("%s must be repo-relative", field)
	}
	return filepath.Join(repoRoot, filepath.FromSlash(filepath.Clean(ref))), nil
}

func isRepoRelativeRef(value string) bool {
	clean := filepath.Clean(strings.TrimSpace(value))
	if clean == "" || clean == "." {
		return false
	}
	if filepath.IsAbs(clean) {
		return false
	}
	return !strings.HasPrefix(clean, "..")
}

func requireDirectory(path, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s %s: %w", label, path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s %s must be a directory", label, path)
	}
	return nil
}

func requireFile(path, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s %s: %w", label, path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("%s %s must be a file", label, path)
	}
	return nil
}

type reservedAdapter struct {
	tool string
}

func (a reservedAdapter) Tool() string {
	return a.tool
}

func (a reservedAdapter) Plan(AdapterRequest) (ExecutionPlan, error) {
	return ExecutionPlan{}, &ExecutionError{
		Kind:   FailureKindUnsupportedTool,
		Tool:   a.tool,
		Action: "run",
		Err:    fmt.Errorf("adapter is reserved for future implementation"),
	}
}
