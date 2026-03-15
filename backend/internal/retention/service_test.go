package retention

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
)

func TestPurgeRemovesExpiredMaterializationsAndRunHistory(t *testing.T) {
	root := t.TempDir()
	dataRoot := filepath.Join(root, "data")
	artifactRoot := filepath.Join(root, "artifacts")
	now := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)

	materializationPath := metadata.MaterializationPath(dataRoot, "raw_transactions")
	if err := os.MkdirAll(filepath.Dir(materializationPath), 0o755); err != nil {
		t.Fatalf("mkdir materialization dir: %v", err)
	}
	if err := os.WriteFile(materializationPath, []byte("id,amount\n1,20\n"), 0o644); err != nil {
		t.Fatalf("write materialization: %v", err)
	}
	oldMaterialization := now.Add(-49 * time.Hour)
	if err := os.Chtimes(materializationPath, oldMaterialization, oldMaterialization); err != nil {
		t.Fatalf("chtimes materialization: %v", err)
	}

	runRoot := filepath.Join(dataRoot, "control_plane", "runs")
	if err := os.MkdirAll(runRoot, 0o755); err != nil {
		t.Fatalf("mkdir run root: %v", err)
	}
	run := orchestration.PipelineRun{
		ID:         "run_old",
		PipelineID: "finance_pipeline",
		Status:     orchestration.RunStatusSucceeded,
		UpdatedAt:  now.Add(-10 * 24 * time.Hour),
	}
	bytes, err := json.Marshal(run)
	if err != nil {
		t.Fatalf("marshal run: %v", err)
	}
	runPath := filepath.Join(runRoot, run.ID+".json")
	if err := os.WriteFile(runPath, bytes, 0o644); err != nil {
		t.Fatalf("write run snapshot: %v", err)
	}

	artifactDir := filepath.Join(artifactRoot, "runs", run.ID)
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		t.Fatalf("mkdir artifact dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(artifactDir, "artifact.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("write artifact: %v", err)
	}

	fakeDB := &recordingDB{
		rowsByQuery: map[string]int{
			`delete from artifact_snapshots where run_id = $1`:                      2,
			`delete from queue_requests where run_id = $1 and status = 'completed'`: 1,
			`delete from run_snapshots where run_id = $1`:                           1,
		},
	}

	service := NewService(Settings{
		DataRoot:              dataRoot,
		ArtifactRoot:          artifactRoot,
		Now:                   now,
		DefaultRunHistoryTTL:  7 * 24 * time.Hour,
		DefaultRunArtifactTTL: 7 * 24 * time.Hour,
	}, fakeDB)

	report, err := service.Purge(context.Background(), []metadata.DataAsset{
		{
			ID: "raw_transactions",
			Retention: metadata.Retention{
				Materializations: "48h",
				RunArtifacts:     "168h",
				RunHistory:       "168h",
			},
		},
	}, []orchestration.Pipeline{
		{
			ID: "finance_pipeline",
			Jobs: []orchestration.Job{
				{
					ID:   "ingest_transactions",
					Type: orchestration.JobTypeIngest,
					Ingest: &orchestration.IngestSpec{
						TargetPath: "raw/raw_transactions.csv",
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("purge returned error: %v", err)
	}

	if _, err := os.Stat(materializationPath); !os.IsNotExist(err) {
		t.Fatalf("expected materialization to be removed, stat err=%v", err)
	}
	if _, err := os.Stat(runPath); !os.IsNotExist(err) {
		t.Fatalf("expected run snapshot to be removed, stat err=%v", err)
	}
	if _, err := os.Stat(artifactDir); !os.IsNotExist(err) {
		t.Fatalf("expected run artifact dir to be removed, stat err=%v", err)
	}
	if report.PostgresRunRowsRemoved != 1 || report.PostgresQueueRowsRemoved != 1 || report.PostgresArtifactRowsRemoved != 2 {
		t.Fatalf("unexpected postgres row counts: %+v", report)
	}
}

type recordingDB struct {
	rowsByQuery map[string]int
}

func (r *recordingDB) ExecContext(_ context.Context, query string, _ ...any) (sql.Result, error) {
	return rowsAffected(r.rowsByQuery[query]), nil
}

type rowsAffected int64

func (r rowsAffected) LastInsertId() (int64, error) { return 0, nil }
func (r rowsAffected) RowsAffected() (int64, error) { return int64(r), nil }
