package execution

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/storage"
)

func TestRunIngestUsesManifestDeclaredSourceAndTarget(t *testing.T) {
	root := t.TempDir()
	sampleRoot := filepath.Join(root, "sample")
	dataRoot := filepath.Join(root, "data")
	artifactRoot := filepath.Join(root, "artifacts")
	mustWriteFile(t, filepath.Join(sampleRoot, "inventory", "movements.csv"), "sku,movement_date,warehouse,quantity\nsku_a,2026-03-01,brisbane,4\n")

	runner := NewRunner(config.Settings{
		SampleDataRoot: sampleRoot,
		DataRoot:       dataRoot,
		ArtifactRoot:   artifactRoot,
		SQLRoot:        filepath.Join(root, "sql"),
		DuckDBPath:     filepath.Join(root, "duckdb", "platform.duckdb"),
	}, nil, orchestration.NewInMemoryStore(), storage.NewService(artifactRoot, nil), slog.New(slog.NewTextHandler(io.Discard, nil)))

	job := orchestration.Job{
		ID:   "ingest_inventory",
		Type: orchestration.JobTypeIngest,
		Ingest: &orchestration.IngestSpec{
			SourceRef:    "inventory/movements.csv",
			TargetPath:   "raw/raw_stock_movements.csv",
			ArtifactPath: "raw/raw_stock_movements.csv",
		},
	}

	if err := runner.runIngest(context.Background(), "run_ingest", job); err != nil {
		t.Fatalf("runIngest returned error: %v", err)
	}

	assertFileExists(t, filepath.Join(dataRoot, "raw", "raw_stock_movements.csv"))
	assertFileExists(t, filepath.Join(artifactRoot, "runs", "run_ingest", "raw", "raw_stock_movements.csv"))
}

func TestRunSQLTransformUsesBootstrapAndDeclaredOutputs(t *testing.T) {
	root := t.TempDir()
	sqlRoot := filepath.Join(root, "sql")
	dataRoot := filepath.Join(root, "data")
	artifactRoot := filepath.Join(root, "artifacts")

	mustWriteFile(t, filepath.Join(dataRoot, "raw", "raw_stock_movements.csv"), strings.TrimSpace(`
sku,movement_date,movement_type,warehouse,quantity,unit_cost
sku_a,2026-02-01,receipt,brisbane,10,2.5
sku_a,2026-02-03,sale,brisbane,-4,2.5
sku_b,2026-03-01,receipt,sydney,6,1.2
`))
	mustWriteFile(t, filepath.Join(sqlRoot, "bootstrap", "raw_stock_movements.sql"), strings.TrimSpace(`
create or replace table raw_stock_movements as
select
  sku,
  cast(movement_date as date) as movement_date,
  movement_type,
  warehouse,
  cast(quantity as integer) as quantity,
  cast(unit_cost as double) as unit_cost
from read_csv_auto({{RAW_STOCK_MOVEMENTS_PATH}}, header = true);
`))
	mustWriteFile(t, filepath.Join(sqlRoot, "transforms", "inventory_monthly_summary.sql"), strings.TrimSpace(`
create or replace table mart_inventory_monthly_summary as
select
  strftime(date_trunc('month', movement_date), '%Y-%m') as month,
  sku,
  warehouse,
  sum(case when quantity >= 0 then quantity else 0 end) as receipts,
  sum(case when quantity < 0 then -quantity else 0 end) as issues,
  sum(quantity) as net_quantity,
  sum(quantity) over (
    partition by sku, warehouse
    order by strftime(date_trunc('month', movement_date), '%Y-%m')
    rows between unbounded preceding and current row
  ) as closing_quantity,
  count(*) as movement_count
from raw_stock_movements
group by 1, 2, 3, movement_date, quantity;
`))

	runner := NewRunner(config.Settings{
		DataRoot:     dataRoot,
		ArtifactRoot: artifactRoot,
		SQLRoot:      sqlRoot,
		DuckDBPath:   filepath.Join(root, "duckdb", "platform.duckdb"),
	}, nil, orchestration.NewInMemoryStore(), storage.NewService(artifactRoot, nil), slog.New(slog.NewTextHandler(io.Discard, nil)))

	job := orchestration.Job{
		ID:           "build_inventory_monthly_summary_mart",
		Type:         orchestration.JobTypeTransformSQL,
		TransformRef: "transform.inventory_monthly_summary",
		Bootstrap: []orchestration.BootstrapSpec{
			{
				SQLRef:      "bootstrap.raw_stock_movements",
				Placeholder: "RAW_STOCK_MOVEMENTS_PATH",
				SourcePath:  "raw/raw_stock_movements.csv",
				Required:    true,
			},
		},
		Outputs: []string{"mart_inventory_monthly_summary"},
	}

	if err := runner.runSQLTransform("run_transform", job); err != nil {
		t.Fatalf("runSQLTransform returned error: %v", err)
	}

	bytes, err := os.ReadFile(filepath.Join(dataRoot, "mart", "mart_inventory_monthly_summary.json"))
	if err != nil {
		t.Fatalf("read mart artifact: %v", err)
	}
	if !strings.Contains(string(bytes), `"month": "2026-02"`) {
		t.Fatalf("expected materialized mart JSON, got %s", string(bytes))
	}
	assertFileExists(t, filepath.Join(artifactRoot, "runs", "run_transform", "mart", "mart_inventory_monthly_summary.json"))
}

func TestRunMetricsPublishUsesDeclaredMetricRefs(t *testing.T) {
	root := t.TempDir()
	sqlRoot := filepath.Join(root, "sql")
	dataRoot := filepath.Join(root, "data")
	artifactRoot := filepath.Join(root, "artifacts")

	mustWriteFile(t, filepath.Join(sqlRoot, "metrics", "metrics_inventory_net_change.sql"), strings.TrimSpace(`
create or replace table metrics_inventory_net_change as
select '2026-02' as month, 'brisbane' as warehouse, 8 as net_quantity;
`))

	runner := NewRunner(config.Settings{
		DataRoot:     dataRoot,
		ArtifactRoot: artifactRoot,
		SQLRoot:      sqlRoot,
		DuckDBPath:   filepath.Join(root, "duckdb", "platform.duckdb"),
	}, nil, orchestration.NewInMemoryStore(), storage.NewService(artifactRoot, nil), slog.New(slog.NewTextHandler(io.Discard, nil)))

	job := orchestration.Job{
		ID:         "publish_inventory_metrics",
		Type:       orchestration.JobTypePublishMetric,
		MetricRefs: []string{"metrics_inventory_net_change"},
	}

	if err := runner.runMetricsPublish("run_metric", job); err != nil {
		t.Fatalf("runMetricsPublish returned error: %v", err)
	}

	bytes, err := os.ReadFile(filepath.Join(dataRoot, "metrics", "metrics_inventory_net_change.json"))
	if err != nil {
		t.Fatalf("read metric artifact: %v", err)
	}
	if !strings.Contains(string(bytes), `"metric_id": "metrics_inventory_net_change"`) {
		t.Fatalf("expected metric artifact payload, got %s", string(bytes))
	}
}
