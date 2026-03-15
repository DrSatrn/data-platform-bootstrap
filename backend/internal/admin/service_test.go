// These tests exercise the product-specific admin command surface so operator
// workflows stay predictable as the management features evolve.
package admin

import (
	"errors"
	"strings"
	"testing"
	"time"

	analyticsinternal "github.com/streanor/data-platform/backend/internal/analytics"
	"github.com/streanor/data-platform/backend/internal/authz"
	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/observability"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/reporting"
)

func TestExecuteHelpListsSupportedCommands(t *testing.T) {
	service := newAdminTestService()

	result := service.Execute("help", authz.Principal{Subject: "alice", Role: authz.RoleAdmin})

	if !result.Success {
		t.Fatalf("expected help to succeed, got %+v", result)
	}
	if len(result.Output) == 0 || result.Output[0] != "help" {
		t.Fatalf("expected command list output, got %+v", result.Output)
	}
	if !containsLine(result.Output, "backup create") {
		t.Fatalf("expected backup command in help output, got %+v", result.Output)
	}
}

func TestExecuteUnknownCommandReturnsGuidance(t *testing.T) {
	service := newAdminTestService()

	result := service.Execute("wat", authz.Principal{Subject: "alice", Role: authz.RoleAdmin})

	if result.Success {
		t.Fatalf("expected unknown command failure, got %+v", result)
	}
	if len(result.Output) < 2 {
		t.Fatalf("expected guidance lines, got %+v", result.Output)
	}
	if !strings.Contains(result.Output[0], `unknown command "wat"`) {
		t.Fatalf("expected unknown-command message, got %+v", result.Output)
	}
	if result.Output[1] != "try: help" {
		t.Fatalf("expected help guidance, got %+v", result.Output)
	}
}

func TestExecuteStatusFormatsTelemetrySummary(t *testing.T) {
	service := newAdminTestService()
	service.telemetry.RecordRequest("GET", "/api/v1/catalog", 200, 12*time.Millisecond)
	service.telemetry.RecordRequest("GET", "/api/v1/system/overview", 500, 20*time.Millisecond)
	service.telemetry.RecordCommand("seed", true, "done")

	result := service.Execute("status", authz.Principal{Subject: "alice", Role: authz.RoleAdmin})

	if !result.Success {
		t.Fatalf("expected status to succeed, got %+v", result)
	}
	if !containsPrefix(result.Output, "environment: development") {
		t.Fatalf("expected environment line, got %+v", result.Output)
	}
	if !containsPrefix(result.Output, "total_requests: 2") {
		t.Fatalf("expected total_requests line, got %+v", result.Output)
	}
	if !containsPrefix(result.Output, "total_errors: 1") {
		t.Fatalf("expected total_errors line, got %+v", result.Output)
	}
	if !containsPrefix(result.Output, "total_commands: 1") {
		t.Fatalf("expected total_commands line, got %+v", result.Output)
	}
}

func TestExecuteAssetsPropagatesLoaderErrors(t *testing.T) {
	service := newAdminTestService()
	service.loader = adminLoaderStub{assetsErr: errors.New("assets unavailable")}

	result := service.Execute("assets", authz.Principal{Subject: "alice", Role: authz.RoleAdmin})

	if result.Success {
		t.Fatalf("expected assets command to fail, got %+v", result)
	}
	if len(result.Output) != 1 || result.Output[0] != "assets unavailable" {
		t.Fatalf("expected loader error output, got %+v", result.Output)
	}
}

func TestExecuteDashboardsFormatsSavedDashboards(t *testing.T) {
	service := newAdminTestService()
	service.reports = reporting.NewMemoryStore()

	result := service.Execute("dashboards", authz.Principal{Subject: "alice", Role: authz.RoleAdmin})

	if !result.Success {
		t.Fatalf("expected dashboards command to succeed, got %+v", result)
	}
	if len(result.Output) == 0 {
		t.Fatalf("expected dashboards output, got %+v", result.Output)
	}
	if !strings.Contains(result.Output[0], "widgets=") {
		t.Fatalf("expected dashboard formatting, got %+v", result.Output)
	}
}

func TestExecuteBackupsReturnsUnavailableWhenServiceMissing(t *testing.T) {
	service := newAdminTestService()
	service.backup = nil

	result := service.Execute("backups", authz.Principal{Subject: "alice", Role: authz.RoleAdmin})

	if result.Success {
		t.Fatalf("expected backup command failure, got %+v", result)
	}
	if len(result.Output) != 1 || result.Output[0] != "backup service is unavailable" {
		t.Fatalf("expected backup unavailable message, got %+v", result.Output)
	}
}

func TestResolveBackupPathRejectsTraversalAndOutsidePaths(t *testing.T) {
	service := newAdminTestService()
	service.cfg.DataRoot = "/tmp/platform-data"

	if _, err := service.resolveBackupPath("../escape.tar.gz"); err == nil {
		t.Fatal("expected traversal path rejection")
	}
	if _, err := service.resolveBackupPath("/tmp/elsewhere/backup.tar.gz"); err == nil {
		t.Fatal("expected outside-root absolute path rejection")
	}
}

func newAdminTestService() *Service {
	return &Service{
		cfg:       config.Settings{Environment: "development", APIBaseURL: "http://127.0.0.1:8080", DataRoot: "/tmp/platform-data"},
		loader:    adminLoaderStub{},
		store:     adminStoreStub{},
		reports:   reporting.NewMemoryStore(),
		telemetry: observability.NewService(),
	}
}

type adminLoaderStub struct {
	pipelines    []orchestration.Pipeline
	pipelinesErr error
	assets       []metadata.DataAsset
	assetsErr    error
	metrics      []analyticsinternal.MetricDefinition
	metricsErr   error
}

func (s adminLoaderStub) LoadPipelines() ([]orchestration.Pipeline, error) {
	if s.pipelinesErr != nil {
		return nil, s.pipelinesErr
	}
	return s.pipelines, nil
}

func (s adminLoaderStub) LoadAssets() ([]metadata.DataAsset, error) {
	if s.assetsErr != nil {
		return nil, s.assetsErr
	}
	return s.assets, nil
}

func (s adminLoaderStub) LoadMetrics() ([]analyticsinternal.MetricDefinition, error) {
	if s.metricsErr != nil {
		return nil, s.metricsErr
	}
	return s.metrics, nil
}

var _ manifests.Loader = adminLoaderStub{}

type adminStoreStub struct {
	runs []orchestration.PipelineRun
	err  error
}

func (s adminStoreStub) ListPipelineRuns() ([]orchestration.PipelineRun, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.runs, nil
}

func (s adminStoreStub) SavePipelineRun(orchestration.PipelineRun) error {
	return nil
}

func (s adminStoreStub) GetPipelineRun(string) (orchestration.PipelineRun, bool, error) {
	return orchestration.PipelineRun{}, false, nil
}

func containsLine(lines []string, needle string) bool {
	for _, line := range lines {
		if line == needle {
			return true
		}
	}
	return false
}

func containsPrefix(lines []string, prefix string) bool {
	for _, line := range lines {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}
