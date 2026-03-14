// This file applies SQL migrations from the repo-managed migration directory.
// Keeping migrations in our own CLI preserves a self-built control plane while
// still leaning on PostgreSQL for durable storage.
package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ApplyMigrations runs all SQL files in lexical order inside a transaction.
func ApplyMigrations(ctx context.Context, conn *sql.DB, migrationsRoot string) error {
	files, err := filepath.Glob(filepath.Join(migrationsRoot, "*.sql"))
	if err != nil {
		return fmt.Errorf("glob migrations: %w", err)
	}
	sort.Strings(files)

	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration transaction: %w", err)
	}
	defer tx.Rollback()

	for _, file := range files {
		bytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", file, err)
		}
		if strings.TrimSpace(string(bytes)) == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx, string(bytes)); err != nil {
			return fmt.Errorf("apply migration %s: %w", filepath.Base(file), err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migrations: %w", err)
	}
	return nil
}
