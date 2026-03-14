// Package transforms owns SQL-based analytical execution against DuckDB. The
// engine is intentionally lightweight: it loads version-controlled SQL files,
// materializes tables, and returns query results in backend-friendly shapes.
package transforms

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb"
)

// Engine executes version-controlled SQL against the configured DuckDB file.
type Engine struct {
	dbPath  string
	sqlRoot string
}

// NewEngine constructs a DuckDB-backed analytical execution engine.
func NewEngine(dbPath, sqlRoot string) *Engine {
	return &Engine{
		dbPath:  dbPath,
		sqlRoot: sqlRoot,
	}
}

// MaterializeRawTables loads landed raw artifacts into DuckDB tables.
func (e *Engine) MaterializeRawTables(rawTransactionsPath, rawBalancesPath string) error {
	if err := e.ExecFile(filepath.Join("bootstrap", "raw_transactions.sql"), map[string]string{
		"RAW_TRANSACTIONS_PATH": quotedSQLString(rawTransactionsPath),
	}); err != nil {
		return err
	}
	return e.ExecFile(filepath.Join("bootstrap", "raw_account_balances.sql"), map[string]string{
		"RAW_ACCOUNT_BALANCES_PATH": quotedSQLString(rawBalancesPath),
	})
}

// RunTransform resolves a manifest transform reference into a SQL file and
// executes it.
func (e *Engine) RunTransform(transformRef string) error {
	name := strings.TrimPrefix(transformRef, "transform.")
	if name == "" || name == transformRef {
		return fmt.Errorf("unsupported transform reference %q", transformRef)
	}
	return e.ExecFile(filepath.Join("transforms", name+".sql"), nil)
}

// RunMetric materializes a metric table from its SQL definition.
func (e *Engine) RunMetric(metricID string) error {
	return e.ExecFile(filepath.Join("metrics", metricID+".sql"), nil)
}

// QueryRows executes a read query and returns JSON-friendly row maps.
func (e *Engine) QueryRows(query string) ([]map[string]any, error) {
	db, err := e.open()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query duckdb rows: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("load duckdb columns: %w", err)
	}

	results := []map[string]any{}
	for rows.Next() {
		scanTargets := make([]any, len(columns))
		values := make([]any, len(columns))
		for index := range scanTargets {
			scanTargets[index] = &values[index]
		}
		if err := rows.Scan(scanTargets...); err != nil {
			return nil, fmt.Errorf("scan duckdb row: %w", err)
		}

		row := make(map[string]any, len(columns))
		for index, name := range columns {
			row[name] = normalizeValue(values[index])
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate duckdb rows: %w", err)
	}
	return results, nil
}

// QueryRowsFromFile loads a SQL file, applies placeholders, and executes it as
// a read query.
func (e *Engine) QueryRowsFromFile(relativePath string, placeholders map[string]string) ([]map[string]any, error) {
	query, err := e.loadSQL(relativePath, placeholders)
	if err != nil {
		return nil, err
	}
	return e.QueryRows(query)
}

// ExecFile loads a SQL file, applies placeholders, and executes it.
func (e *Engine) ExecFile(relativePath string, placeholders map[string]string) error {
	query, err := e.loadSQL(relativePath, placeholders)
	if err != nil {
		return err
	}
	db, err := e.open()
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Exec(query); err != nil {
		return fmt.Errorf("exec duckdb sql %s: %w", relativePath, err)
	}
	return nil
}

func (e *Engine) open() (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(e.dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("create duckdb dir: %w", err)
	}
	db, err := sql.Open("duckdb", e.dbPath)
	if err != nil {
		return nil, fmt.Errorf("open duckdb database: %w", err)
	}
	db.SetMaxOpenConns(1)
	return db, nil
}

func (e *Engine) loadSQL(relativePath string, placeholders map[string]string) (string, error) {
	path := filepath.Join(e.sqlRoot, relativePath)
	bytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read sql file %s: %w", path, err)
	}
	sqlText := string(bytes)
	for key, value := range placeholders {
		sqlText = strings.ReplaceAll(sqlText, "{{"+key+"}}", value)
	}
	return sqlText, nil
}

func quotedSQLString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func normalizeValue(value any) any {
	switch typed := value.(type) {
	case []byte:
		return string(typed)
	case time.Time:
		return typed.UTC().Format(time.RFC3339)
	default:
		return typed
	}
}
