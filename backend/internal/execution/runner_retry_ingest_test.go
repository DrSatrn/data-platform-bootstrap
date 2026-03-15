package execution

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"encoding/csv"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/ingestion"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/storage"
)

func TestExecuteJobRetriesWithExponentialBackoffAndIdempotencyKeys(t *testing.T) {
	store := orchestration.NewInMemoryStore()
	runner := NewRunner(config.Settings{
		DataRoot:          t.TempDir(),
		ArtifactRoot:      t.TempDir(),
		JobRetryBaseDelay: 10 * time.Millisecond,
	}, nil, store, storage.NewService(t.TempDir(), nil), slog.New(slog.NewTextHandler(io.Discard, nil)))

	run := orchestration.PipelineRun{ID: "run_1", PipelineID: "pipeline_1"}
	job := orchestration.Job{ID: "job_1", Type: orchestration.JobTypeTransformSQL, Retries: 2}

	var (
		keys   []string
		delays []time.Duration
	)
	runner.executeAttemptOverride = func(ctx context.Context, run *orchestration.PipelineRun, job orchestration.Job, idempotencyKey string) error {
		keys = append(keys, idempotencyKey)
		if len(keys) < 3 {
			return errors.New("transient failure")
		}
		return nil
	}
	runner.sleep = func(ctx context.Context, delay time.Duration) error {
		delays = append(delays, delay)
		return nil
	}

	if err := runner.executeJob(context.Background(), &run, job); err != nil {
		t.Fatalf("executeJob returned error: %v", err)
	}

	if got, want := keys, []string{"run_1:job_1:1", "run_1:job_1:2", "run_1:job_1:3"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected idempotency keys: got %v want %v", got, want)
	}
	if got, want := delays, []time.Duration{10 * time.Millisecond, 20 * time.Millisecond}; !reflect.DeepEqual(got, want) {
		t.Fatalf("unexpected retry delays: got %v want %v", got, want)
	}
	jobRun := findJobRun(&run, job.ID)
	if jobRun.Attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", jobRun.Attempts)
	}
	if jobRun.Status != orchestration.RunStatusSucceeded {
		t.Fatalf("expected succeeded job status, got %s", jobRun.Status)
	}
}

func TestRunIngestExportsDatabaseRowsToCSV(t *testing.T) {
	tests := []struct {
		name          string
		sourceKind    string
		connectionEnv string
	}{
		{name: "postgres", sourceKind: "postgres", connectionEnv: "TEST_PLATFORM_POSTGRES_INGEST_DSN"},
		{name: "mysql", sourceKind: "mysql", connectionEnv: "TEST_PLATFORM_MYSQL_INGEST_DSN"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			root := t.TempDir()
			dataRoot := filepath.Join(root, "data")
			artifactRoot := filepath.Join(root, "artifacts")
			if err := os.Setenv(testCase.connectionEnv, "driver=test"); err != nil {
				t.Fatalf("set env: %v", err)
			}
			t.Cleanup(func() { _ = os.Unsetenv(testCase.connectionEnv) })

			runner := NewRunner(config.Settings{
				DataRoot:     dataRoot,
				ArtifactRoot: artifactRoot,
				SQLRoot:      filepath.Join(root, "sql"),
				DuckDBPath:   filepath.Join(root, "duckdb", "platform.duckdb"),
			}, nil, orchestration.NewInMemoryStore(), storage.NewService(artifactRoot, nil), slog.New(slog.NewTextHandler(io.Discard, nil)))
			runner.ingest = ingestion.NewExporterWithOpener(func(driverName, dsn string) (*sql.DB, error) {
				rows := fakeRows{
					columns: []string{"account_id", "balance"},
					values: [][]driver.Value{
						{"acct_1", "1250.50"},
						{"acct_2", "980.00"},
					},
				}
				return sql.OpenDB(fakeConnector{rows: rows}), nil
			})

			job := orchestration.Job{
				ID:   "ingest_db_rows",
				Type: orchestration.JobTypeIngest,
				Ingest: &orchestration.IngestSpec{
					SourceKind:    testCase.sourceKind,
					ConnectionEnv: testCase.connectionEnv,
					Query:         "select account_id, balance from balances",
					TargetPath:    "raw/raw_account_balances.csv",
					ArtifactPath:  "raw/raw_account_balances.csv",
					Format:        "csv",
				},
			}

			if err := runner.runIngest(context.Background(), "run_ingest_db", job); err != nil {
				t.Fatalf("runIngest returned error: %v", err)
			}

			targetPath := filepath.Join(dataRoot, "raw", "raw_account_balances.csv")
			assertFileExists(t, targetPath)
			assertFileExists(t, filepath.Join(artifactRoot, "runs", "run_ingest_db", "raw", "raw_account_balances.csv"))

			file, err := os.Open(targetPath)
			if err != nil {
				t.Fatalf("open csv target: %v", err)
			}
			defer file.Close()
			records, err := csv.NewReader(file).ReadAll()
			if err != nil {
				t.Fatalf("read csv target: %v", err)
			}
			if len(records) != 3 {
				t.Fatalf("expected header plus two rows, got %d", len(records))
			}
			if strings.Join(records[1], ",") != "acct_1,1250.50" {
				t.Fatalf("unexpected first data row %v", records[1])
			}
		})
	}
}

type fakeConnector struct {
	rows fakeRows
}

func (c fakeConnector) Connect(context.Context) (driver.Conn, error) {
	rows := c.rows
	return fakeConn{rows: &rows}, nil
}

func (c fakeConnector) Driver() driver.Driver {
	return fakeDriver{rows: c.rows}
}

type fakeDriver struct {
	rows fakeRows
}

func (d fakeDriver) Open(name string) (driver.Conn, error) {
	rows := d.rows
	return fakeConn{rows: &rows}, nil
}

type fakeConn struct {
	rows *fakeRows
}

func (c fakeConn) Prepare(query string) (driver.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (c fakeConn) Close() error              { return nil }
func (c fakeConn) Begin() (driver.Tx, error) { return nil, errors.New("not implemented") }
func (c fakeConn) QueryContext(context.Context, string, []driver.NamedValue) (driver.Rows, error) {
	return c.rows, nil
}

type fakeRows struct {
	columns []string
	values  [][]driver.Value
	index   int
}

func (r *fakeRows) Columns() []string { return append([]string{}, r.columns...) }
func (r *fakeRows) Close() error      { return nil }
func (r *fakeRows) Next(dest []driver.Value) error {
	if r.index >= len(r.values) {
		return io.EOF
	}
	copy(dest, r.values[r.index])
	r.index++
	return nil
}
