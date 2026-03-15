package scheduler

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/streanor/data-platform/backend/internal/alerting"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
)

func TestRefreshCatalogPostsWebhookWhenAssetTurnsStale(t *testing.T) {
	root := t.TempDir()
	manifestRoot := filepath.Join(root, "manifests")
	dataRoot := filepath.Join(root, "data")
	assetID := "metrics_savings_rate"
	assetPath := metadata.MaterializationPath(dataRoot, assetID)

	mustWriteManifest(t, filepath.Join(manifestRoot, "assets", "asset.yaml"), ""+
		"id: metrics_savings_rate\n"+
		"name: Savings Rate\n"+
		"layer: mart\n"+
		"description: Savings rate metric.\n"+
		"owner: platform-team\n"+
		"kind: metric\n"+
		"freshness:\n"+
		"  expected_within: 1h\n"+
		"  warn_after: 2h\n"+
		"columns:\n"+
		"  - name: month\n"+
		"    type: text\n")
	mustWriteFile(t, assetPath, "{}")
	staleAt := time.Now().UTC().Add(-3 * time.Hour)
	if err := os.Chtimes(assetPath, staleAt, staleAt); err != nil {
		t.Fatalf("set stale modtime: %v", err)
	}

	var (
		mu    sync.Mutex
		calls int
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode webhook payload: %v", err)
		}
		mu.Lock()
		calls++
		mu.Unlock()
		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	store := orchestration.NewInMemoryStore()
	queue, err := orchestration.NewQueue(dataRoot)
	if err != nil {
		t.Fatalf("new queue: %v", err)
	}
	service := NewService(
		time.Hour,
		manifests.NewLoader(manifestRoot),
		store,
		orchestration.NewControlService(manifests.NewLoader(manifestRoot), store, queue),
		metadata.NewCatalog(),
		nil,
		alerting.NewDispatcher(alerting.Settings{
			Environment:      "test",
			AssetWarningURLs: []string{server.URL},
			WebhookTimeout:   5 * time.Second,
		}, nil),
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		dataRoot,
	)

	if err := service.refreshCatalog(context.Background()); err != nil {
		t.Fatalf("refreshCatalog returned error: %v", err)
	}
	if err := service.refreshCatalog(context.Background()); err != nil {
		t.Fatalf("second refreshCatalog returned error: %v", err)
	}

	mu.Lock()
	defer mu.Unlock()
	if calls != 1 {
		t.Fatalf("expected 1 asset alert webhook, got %d", calls)
	}
}

func mustWriteManifest(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func mustWriteFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
