package orchestration

import "testing"

func TestValidatePipelineAcceptsManifestDrivenIngestAndMetrics(t *testing.T) {
	pipeline := Pipeline{
		ID: "inventory_pipeline",
		Jobs: []Job{
			{
				ID:   "ingest_stock_movements_csv",
				Type: JobTypeIngest,
				Ingest: &IngestSpec{
					SourceRef:    "inventory_operations/stock_movements.csv",
					TargetPath:   "raw/raw_stock_movements.csv",
					ArtifactPath: "raw/raw_stock_movements.csv",
				},
			},
			{
				ID:           "build_inventory_monthly_summary_mart",
				Type:         JobTypeTransformSQL,
				DependsOn:    []string{"ingest_stock_movements_csv"},
				TransformRef: "transform.inventory_monthly_summary",
				Bootstrap: []BootstrapSpec{{
					SQLRef:      "bootstrap.raw_stock_movements",
					Placeholder: "RAW_STOCK_MOVEMENTS_PATH",
					SourcePath:  "raw/raw_stock_movements.csv",
					Required:    true,
				}},
				Outputs: []string{"mart_inventory_monthly_summary"},
			},
			{
				ID:         "publish_inventory_metrics",
				Type:       JobTypePublishMetric,
				DependsOn:  []string{"build_inventory_monthly_summary_mart"},
				MetricRefs: []string{"metrics_inventory_net_change"},
			},
		},
	}

	if err := ValidatePipeline(pipeline); err != nil {
		t.Fatalf("expected manifest-driven pipeline to validate, got %v", err)
	}
}

func TestValidatePipelineRejectsIngestWithoutSourceOrTarget(t *testing.T) {
	pipeline := Pipeline{
		ID: "broken_ingest",
		Jobs: []Job{{
			ID:     "ingest",
			Type:   JobTypeIngest,
			Ingest: &IngestSpec{},
		}},
	}

	if err := ValidatePipeline(pipeline); err == nil {
		t.Fatal("expected ingest validation error")
	}
}

func TestValidatePipelineRejectsPublishMetricWithoutMetricRefs(t *testing.T) {
	pipeline := Pipeline{
		ID: "broken_metric_publish",
		Jobs: []Job{{
			ID:   "publish_metrics",
			Type: JobTypePublishMetric,
		}},
	}

	if err := ValidatePipeline(pipeline); err == nil {
		t.Fatal("expected publish_metric validation error")
	}
}

func TestValidatePipelineAcceptsDatabaseIngestJobs(t *testing.T) {
	pipeline := Pipeline{
		ID: "database_ingest_pipeline",
		Jobs: []Job{{
			ID:   "ingest_balances_postgres",
			Type: JobTypeIngest,
			Ingest: &IngestSpec{
				SourceKind:    "postgres",
				ConnectionEnv: "PLATFORM_SOURCE_FINANCE_POSTGRES_DSN",
				Query:         "select account_id, balance from balances",
				TargetPath:    "raw/raw_account_balances.csv",
				ArtifactPath:  "raw/raw_account_balances.csv",
				Format:        "csv",
			},
		}},
	}

	if err := ValidatePipeline(pipeline); err != nil {
		t.Fatalf("expected database ingest pipeline to validate, got %v", err)
	}
}

func TestValidatePipelineRejectsDatabaseIngestWithoutConnectionEnv(t *testing.T) {
	pipeline := Pipeline{
		ID: "broken_database_ingest",
		Jobs: []Job{{
			ID:   "ingest_mysql",
			Type: JobTypeIngest,
			Ingest: &IngestSpec{
				SourceKind: "mysql",
				Query:      "select * from inventory",
				TargetPath: "raw/raw_inventory.csv",
			},
		}},
	}

	if err := ValidatePipeline(pipeline); err == nil {
		t.Fatal("expected database ingest validation error")
	}
}
