package manifests

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileLoaderReadsPipelinesAssetsAndMetrics(t *testing.T) {
	root := t.TempDir()
	mustWriteManifest(t, filepath.Join(root, "pipelines", "pipeline.yaml"), ""+
		"id: sample_pipeline\n"+
		"name: Sample Pipeline\n"+
		"description: Sample pipeline.\n"+
		"owner: platform-team\n"+
		"jobs:\n"+
		"  - id: ingest_sample\n"+
		"    name: Ingest Sample\n"+
		"    type: ingest\n")
	mustWriteManifest(t, filepath.Join(root, "assets", "asset.yaml"), ""+
		"id: sample_asset\n"+
		"name: Sample Asset\n"+
		"layer: raw\n"+
		"description: Sample asset.\n"+
		"owner: platform-team\n"+
		"kind: table\n"+
		"freshness:\n"+
		"  expected_within: 1h\n"+
		"  warn_after: 2h\n"+
		"columns:\n"+
		"  - name: id\n"+
		"    type: text\n"+
		"    description: Identifier.\n"+
		"    is_pii: false\n")
	mustWriteManifest(t, filepath.Join(root, "metrics", "metric.yaml"), ""+
		"id: sample_metric\n"+
		"name: Sample Metric\n"+
		"description: Sample metric.\n"+
		"owner: platform-team\n"+
		"dataset_ref: sample_asset\n"+
		"time_dimension: month\n"+
		"dimensions:\n"+
		"  - month\n"+
		"measures:\n"+
		"  - total\n"+
		"default_visualization: line\n")

	loader := NewLoader(root)

	pipelines, err := loader.LoadPipelines()
	if err != nil {
		t.Fatalf("LoadPipelines returned error: %v", err)
	}
	if len(pipelines) != 1 || pipelines[0].ID != "sample_pipeline" {
		t.Fatalf("unexpected pipelines payload: %+v", pipelines)
	}

	assets, err := loader.LoadAssets()
	if err != nil {
		t.Fatalf("LoadAssets returned error: %v", err)
	}
	if len(assets) != 1 || assets[0].ID != "sample_asset" {
		t.Fatalf("unexpected assets payload: %+v", assets)
	}

	metrics, err := loader.LoadMetrics()
	if err != nil {
		t.Fatalf("LoadMetrics returned error: %v", err)
	}
	if len(metrics) != 1 || metrics[0].ID != "sample_metric" {
		t.Fatalf("unexpected metrics payload: %+v", metrics)
	}
}

func TestFileLoaderReturnsDecodeErrorForMalformedManifest(t *testing.T) {
	root := t.TempDir()
	mustWriteManifest(t, filepath.Join(root, "pipelines", "broken.yaml"), "id: broken_pipeline\njobs: [\n")

	loader := NewLoader(root)
	if _, err := loader.LoadPipelines(); err == nil {
		t.Fatal("expected malformed pipeline manifest to fail")
	}
}

func mustWriteManifest(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
