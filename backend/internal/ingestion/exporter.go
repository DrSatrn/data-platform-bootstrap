// Package ingestion implements the low-level source export logic used by ingest
// jobs. The worker remains the control plane and delegates only the physical
// copy/query-to-file step to this package.
package ingestion

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DBOpener allows tests to replace the real sql.Open behavior with a fake
// database while the production path continues to use native drivers.
type DBOpener func(driverName, dsn string) (*sql.DB, error)

// Exporter copies file-backed inputs and exports database query results into
// the raw file lake.
type Exporter struct {
	open DBOpener
}

// NewExporter constructs a production exporter.
func NewExporter() *Exporter {
	return &Exporter{open: sql.Open}
}

// NewExporterWithOpener constructs an exporter with a custom DB opener for
// tests and narrowly scoped integration seams.
func NewExporterWithOpener(open DBOpener) *Exporter {
	if open == nil {
		open = sql.Open
	}
	return &Exporter{open: open}
}

// DatabaseSpec defines one bounded query export from a relational source into
// a local CSV file.
type DatabaseSpec struct {
	Driver        string
	ConnectionEnv string
	Query         string
	TargetPath    string
}

// ExportQueryToCSV runs one bounded query against a supported database and
// materializes the rows as CSV under the local data root.
func (e *Exporter) ExportQueryToCSV(ctx context.Context, spec DatabaseSpec) error {
	driverName, err := normalizeDriver(spec.Driver)
	if err != nil {
		return err
	}
	dsn := strings.TrimSpace(os.Getenv(strings.TrimSpace(spec.ConnectionEnv)))
	if dsn == "" {
		return fmt.Errorf("connection env %s is empty", spec.ConnectionEnv)
	}
	db, err := e.open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("open %s ingest source: %w", driverName, err)
	}
	defer db.Close()

	rows, err := db.QueryContext(ctx, spec.Query)
	if err != nil {
		return fmt.Errorf("query %s ingest source: %w", driverName, err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("read column metadata: %w", err)
	}
	if len(columns) == 0 {
		return fmt.Errorf("query returned no columns")
	}
	if err := os.MkdirAll(filepath.Dir(spec.TargetPath), 0o755); err != nil {
		return fmt.Errorf("create ingest target dir: %w", err)
	}

	file, err := os.Create(spec.TargetPath)
	if err != nil {
		return fmt.Errorf("create ingest target %s: %w", spec.TargetPath, err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	if err := writer.Write(columns); err != nil {
		return fmt.Errorf("write csv header: %w", err)
	}

	values := make([]any, len(columns))
	scanTargets := make([]any, len(columns))
	for index := range values {
		scanTargets[index] = &values[index]
	}

	for rows.Next() {
		if err := rows.Scan(scanTargets...); err != nil {
			return fmt.Errorf("scan ingest row: %w", err)
		}
		record := make([]string, len(columns))
		for index, value := range values {
			record[index] = stringifyValue(value)
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("write csv row: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate ingest rows: %w", err)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return fmt.Errorf("flush csv writer: %w", err)
	}
	return nil
}

func normalizeDriver(driver string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(driver)) {
	case "postgres", "postgresql", "pgx":
		return "pgx", nil
	case "mysql":
		return "mysql", nil
	default:
		return "", fmt.Errorf("unsupported database driver %q", driver)
	}
}

func stringifyValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case []byte:
		return string(typed)
	default:
		return fmt.Sprint(typed)
	}
}
