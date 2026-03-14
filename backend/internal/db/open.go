// Package db contains PostgreSQL persistence and migration helpers. The
// runtime uses these adapters opportunistically so the local-first platform can
// keep working even when PostgreSQL has not been bootstrapped yet.
package db

import (
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open creates a PostgreSQL connection pool using the pgx stdlib driver.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open postgres connection: %w", err)
	}
	return db, nil
}
