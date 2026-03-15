// Package backup creates and verifies first-party recovery bundles for the
// platform. The implementation favors transparent archive contents and a
// machine-readable manifest so self-hosted operators can inspect what would be
// recoverable before they need it in an incident.
package backup

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/streanor/data-platform/backend/internal/audit"
	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/reporting"
)

const (
	// FormatVersion makes future format migrations explicit.
	FormatVersion = "2026-03-15"
)

// QueueSnapshotter describes the optional queue export behavior used during
// bundle creation. Runtime queues are not required to implement this, but both
// current local-first implementations do.
type QueueSnapshotter interface {
	ListRequests() ([]orchestration.QueueSnapshot, error)
}

// Service owns backup bundle creation and verification.
type Service struct {
	cfg      config.Settings
	loader   manifests.Loader
	runs     orchestration.Store
	queue    QueueSnapshotter
	reports  reporting.Store
	audit    audit.Store
	metadata metadata.Store
}

// BundleResult describes a created backup bundle.
type BundleResult struct {
	Path     string         `json:"path"`
	Manifest BackupManifest `json:"manifest"`
}

// BundleFile describes one archived file captured in the bundle manifest.
type BundleFile struct {
	Path       string `json:"path"`
	SizeBytes  int64  `json:"size_bytes"`
	SHA256     string `json:"sha256"`
	SourceKind string `json:"source_kind"`
}

// BundleCounts summarizes the main control-plane entities captured by the
// bundle so operators can quickly sanity-check coverage.
type BundleCounts struct {
	PipelineRuns  int `json:"pipeline_runs"`
	QueueRequests int `json:"queue_requests"`
	Dashboards    int `json:"dashboards"`
	AuditEvents   int `json:"audit_events"`
	DataAssets    int `json:"data_assets"`
	BundleFiles   int `json:"bundle_files"`
}

// BackupManifest records bundle metadata and inventory.
type BackupManifest struct {
	FormatVersion string       `json:"format_version"`
	GeneratedAt   time.Time    `json:"generated_at"`
	Environment   string       `json:"environment"`
	Counts        BundleCounts `json:"counts"`
	Files         []BundleFile `json:"files"`
	Warnings      []string     `json:"warnings,omitempty"`
	ConfigSummary ConfigExport `json:"config_summary"`
}

// ConfigExport is a sanitized runtime summary suitable for public-safe bundles.
// Secrets, DSNs, hostnames, and bearer tokens are intentionally excluded.
type ConfigExport struct {
	HTTPAddr          string `json:"http_addr"`
	APIBaseURL        string `json:"api_base_url"`
	LogLevel          string `json:"log_level"`
	SchedulerTick     string `json:"scheduler_tick"`
	WorkerPoll        string `json:"worker_poll"`
	MaxConcurrentJobs int    `json:"max_concurrent_jobs"`
	DataRoot          string `json:"data_root"`
	ArtifactRoot      string `json:"artifact_root"`
	DuckDBPath        string `json:"duckdb_path"`
	ManifestRoot      string `json:"manifest_root"`
	DashboardRoot     string `json:"dashboard_root"`
	SQLRoot           string `json:"sql_root"`
	PythonTaskRoot    string `json:"python_task_root"`
}

// NewService constructs a backup service for the current runtime.
func NewService(
	cfg config.Settings,
	loader manifests.Loader,
	runs orchestration.Store,
	queue QueueSnapshotter,
	reports reporting.Store,
	auditStore audit.Store,
	metadataStore metadata.Store,
) *Service {
	return &Service{
		cfg:      cfg,
		loader:   loader,
		runs:     runs,
		queue:    queue,
		reports:  reports,
		audit:    auditStore,
		metadata: metadataStore,
	}
}

// Create writes a portable `.tar.gz` bundle and returns its manifest.
func (s *Service) Create(outputPath string) (BundleResult, error) {
	if outputPath == "" {
		outputPath = filepath.Join(s.cfg.DataRoot, "backups", "platform-backup-"+time.Now().UTC().Format("20060102T150405Z")+".tar.gz")
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return BundleResult{}, fmt.Errorf("create backup output dir: %w", err)
	}

	runs, err := s.runs.ListPipelineRuns()
	if err != nil {
		return BundleResult{}, fmt.Errorf("list pipeline runs: %w", err)
	}
	dashboards, err := s.reports.ListDashboards()
	if err != nil {
		return BundleResult{}, fmt.Errorf("list dashboards: %w", err)
	}
	auditEvents, err := s.audit.ListRecent(100000)
	if err != nil {
		return BundleResult{}, fmt.Errorf("list audit events: %w", err)
	}
	assets, assetWarning, err := s.loadAssets()
	if err != nil {
		return BundleResult{}, err
	}
	queueRequests, queueWarning, err := s.loadQueueRequests()
	if err != nil {
		return BundleResult{}, err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return BundleResult{}, fmt.Errorf("create backup bundle: %w", err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	manifest := BackupManifest{
		FormatVersion: FormatVersion,
		GeneratedAt:   time.Now().UTC(),
		Environment:   s.cfg.Environment,
		ConfigSummary: sanitizeConfig(s.cfg),
		Warnings:      compactWarnings(assetWarning, queueWarning),
	}

	if err := writeJSONExport(tarWriter, &manifest, "exports/pipeline_runs.json", runs, "control_plane_export"); err != nil {
		return BundleResult{}, err
	}
	if err := writeJSONExport(tarWriter, &manifest, "exports/queue_requests.json", queueRequests, "control_plane_export"); err != nil {
		return BundleResult{}, err
	}
	if err := writeJSONExport(tarWriter, &manifest, "exports/dashboards.json", dashboards, "control_plane_export"); err != nil {
		return BundleResult{}, err
	}
	if err := writeJSONExport(tarWriter, &manifest, "exports/audit_events.json", auditEvents, "control_plane_export"); err != nil {
		return BundleResult{}, err
	}
	if err := writeJSONExport(tarWriter, &manifest, "exports/data_assets.json", assets, "metadata_export"); err != nil {
		return BundleResult{}, err
	}
	if err := writeJSONExport(tarWriter, &manifest, "exports/config_summary.json", manifest.ConfigSummary, "config_export"); err != nil {
		return BundleResult{}, err
	}

	for _, include := range []struct {
		source string
		target string
		kind   string
	}{
		{source: filepath.Join(s.cfg.DataRoot, "control_plane"), target: "files/data/control_plane", kind: "data_root"},
		{source: filepath.Join(s.cfg.DataRoot, "raw"), target: "files/data/raw", kind: "data_root"},
		{source: filepath.Join(s.cfg.DataRoot, "mart"), target: "files/data/mart", kind: "data_root"},
		{source: filepath.Join(s.cfg.DataRoot, "metrics"), target: "files/data/metrics", kind: "data_root"},
		{source: filepath.Join(s.cfg.DataRoot, "quality"), target: "files/data/quality", kind: "data_root"},
		{source: s.cfg.ArtifactRoot, target: "files/artifacts", kind: "artifact_root"},
		{source: s.cfg.DuckDBPath, target: "files/duckdb/platform.duckdb", kind: "duckdb"},
		{source: s.cfg.ManifestRoot, target: "files/repo/manifests", kind: "repo_snapshot"},
		{source: s.cfg.DashboardRoot, target: "files/repo/dashboards", kind: "repo_snapshot"},
		{source: s.cfg.SQLRoot, target: "files/repo/sql", kind: "repo_snapshot"},
		{source: s.cfg.PythonTaskRoot, target: "files/repo/python", kind: "repo_snapshot"},
	} {
		if err := addPath(tarWriter, &manifest, include.source, include.target, include.kind); err != nil {
			return BundleResult{}, err
		}
	}

	manifest.Counts = BundleCounts{
		PipelineRuns:  len(runs),
		QueueRequests: len(queueRequests),
		Dashboards:    len(dashboards),
		AuditEvents:   len(auditEvents),
		DataAssets:    len(assets),
		BundleFiles:   len(manifest.Files),
	}

	manifestBytes, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return BundleResult{}, fmt.Errorf("encode backup manifest: %w", err)
	}
	if err := writeTarFile(tarWriter, "manifest.json", manifestBytes, manifest.GeneratedAt); err != nil {
		return BundleResult{}, fmt.Errorf("write backup manifest: %w", err)
	}

	return BundleResult{Path: outputPath, Manifest: manifest}, nil
}

// Verify validates that a bundle contains the required manifest and exports,
// and that the manifest inventory matches the actual archive payload.
func (s *Service) Verify(path string) (BackupManifest, error) {
	file, err := os.Open(path)
	if err != nil {
		return BackupManifest{}, fmt.Errorf("open backup bundle: %w", err)
	}
	defer file.Close()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return BackupManifest{}, fmt.Errorf("open gzip reader: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	seen := map[string]BundleFile{}
	var manifest BackupManifest
	foundManifest := false

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return BackupManifest{}, fmt.Errorf("read backup archive: %w", err)
		}
		if header.Typeflag != tar.TypeReg {
			continue
		}
		bytes, err := io.ReadAll(tarReader)
		if err != nil {
			return BackupManifest{}, fmt.Errorf("read archive member %s: %w", header.Name, err)
		}
		if header.Name == "manifest.json" {
			if err := json.Unmarshal(bytes, &manifest); err != nil {
				return BackupManifest{}, fmt.Errorf("decode backup manifest: %w", err)
			}
			foundManifest = true
			continue
		}
		sum := sha256.Sum256(bytes)
		seen[header.Name] = BundleFile{
			Path:      header.Name,
			SizeBytes: int64(len(bytes)),
			SHA256:    hex.EncodeToString(sum[:]),
		}
	}

	if !foundManifest {
		return BackupManifest{}, fmt.Errorf("backup bundle is missing manifest.json")
	}
	for _, required := range []string{
		"exports/pipeline_runs.json",
		"exports/queue_requests.json",
		"exports/dashboards.json",
		"exports/audit_events.json",
		"exports/data_assets.json",
		"exports/config_summary.json",
	} {
		if _, ok := seen[required]; !ok {
			return BackupManifest{}, fmt.Errorf("backup bundle is missing required export %s", required)
		}
	}
	for _, file := range manifest.Files {
		actual, ok := seen[file.Path]
		if !ok {
			return BackupManifest{}, fmt.Errorf("manifest entry %s not found in archive", file.Path)
		}
		if actual.SizeBytes != file.SizeBytes {
			return BackupManifest{}, fmt.Errorf("size mismatch for %s", file.Path)
		}
		if actual.SHA256 != file.SHA256 {
			return BackupManifest{}, fmt.Errorf("checksum mismatch for %s", file.Path)
		}
	}
	return manifest, nil
}

// ListBundles returns the locally created bundles under the configured data
// root. This keeps the admin terminal and CLI able to discover previous
// recovery points without reaching into arbitrary paths.
func (s *Service) ListBundles() ([]BundleFile, error) {
	root := filepath.Join(s.cfg.DataRoot, "backups")
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []BundleFile{}, nil
		}
		return nil, fmt.Errorf("read backup dir: %w", err)
	}
	files := []BundleFile{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".tar.gz") {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, fmt.Errorf("stat backup bundle %s: %w", entry.Name(), err)
		}
		files = append(files, BundleFile{
			Path:      filepath.Join(root, entry.Name()),
			SizeBytes: info.Size(),
		})
	}
	sort.Slice(files, func(left, right int) bool {
		return files[left].Path > files[right].Path
	})
	return files, nil
}

func (s *Service) loadAssets() ([]metadata.DataAsset, string, error) {
	if s.metadata != nil {
		assets, err := s.metadata.ListAssets()
		if err == nil && len(assets) > 0 {
			return assets, "", nil
		}
	}
	assets, err := s.loader.LoadAssets()
	if err != nil {
		return nil, "", fmt.Errorf("load assets for backup: %w", err)
	}
	return assets, "metadata exported from manifests because the normalized metadata store was empty or unavailable", nil
}

func (s *Service) loadQueueRequests() ([]orchestration.QueueSnapshot, string, error) {
	if s.queue == nil {
		return []orchestration.QueueSnapshot{}, "queue export unavailable because the runtime queue does not expose snapshots", nil
	}
	requests, err := s.queue.ListRequests()
	if err != nil {
		return nil, "", fmt.Errorf("list queue requests for backup: %w", err)
	}
	return requests, "", nil
}

func sanitizeConfig(cfg config.Settings) ConfigExport {
	return ConfigExport{
		HTTPAddr:          cfg.HTTPAddr,
		APIBaseURL:        cfg.APIBaseURL,
		LogLevel:          cfg.LogLevel,
		SchedulerTick:     cfg.SchedulerTick.String(),
		WorkerPoll:        cfg.WorkerPoll.String(),
		MaxConcurrentJobs: cfg.MaxConcurrentJobs,
		DataRoot:          cfg.DataRoot,
		ArtifactRoot:      cfg.ArtifactRoot,
		DuckDBPath:        cfg.DuckDBPath,
		ManifestRoot:      cfg.ManifestRoot,
		DashboardRoot:     cfg.DashboardRoot,
		SQLRoot:           cfg.SQLRoot,
		PythonTaskRoot:    cfg.PythonTaskRoot,
	}
}

func compactWarnings(values ...string) []string {
	warnings := []string{}
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			warnings = append(warnings, value)
		}
	}
	return warnings
}

func writeJSONExport(writer *tar.Writer, manifest *BackupManifest, target string, payload any, sourceKind string) error {
	bytes, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("encode backup export %s: %w", target, err)
	}
	if err := writeTarFile(writer, target, bytes, manifest.GeneratedAt); err != nil {
		return fmt.Errorf("write backup export %s: %w", target, err)
	}
	sum := sha256.Sum256(bytes)
	manifest.Files = append(manifest.Files, BundleFile{
		Path:       target,
		SizeBytes:  int64(len(bytes)),
		SHA256:     hex.EncodeToString(sum[:]),
		SourceKind: sourceKind,
	})
	return nil
}

func addPath(writer *tar.Writer, manifest *BackupManifest, sourcePath, targetPath, sourceKind string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("stat backup include %s: %w", sourcePath, err)
	}
	if !info.IsDir() {
		return addFile(writer, manifest, sourcePath, targetPath, sourceKind, info)
	}
	return filepath.WalkDir(sourcePath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("walk backup include %s: %w", path, walkErr)
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("load backup entry info %s: %w", path, err)
		}
		relative, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return fmt.Errorf("derive backup relative path for %s: %w", path, err)
		}
		return addFile(writer, manifest, path, filepath.ToSlash(filepath.Join(targetPath, relative)), sourceKind, info)
	})
}

func addFile(writer *tar.Writer, manifest *BackupManifest, sourcePath, targetPath, sourceKind string, info fs.FileInfo) error {
	bytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("read backup source %s: %w", sourcePath, err)
	}
	if err := writeTarFile(writer, filepath.ToSlash(targetPath), bytes, info.ModTime().UTC()); err != nil {
		return fmt.Errorf("archive backup source %s: %w", sourcePath, err)
	}
	sum := sha256.Sum256(bytes)
	manifest.Files = append(manifest.Files, BundleFile{
		Path:       filepath.ToSlash(targetPath),
		SizeBytes:  int64(len(bytes)),
		SHA256:     hex.EncodeToString(sum[:]),
		SourceKind: sourceKind,
	})
	return nil
}

func writeTarFile(writer *tar.Writer, path string, bytes []byte, modifiedAt time.Time) error {
	header := &tar.Header{
		Name:    filepath.ToSlash(path),
		Mode:    0o644,
		Size:    int64(len(bytes)),
		ModTime: modifiedAt,
	}
	if err := writer.WriteHeader(header); err != nil {
		return err
	}
	_, err := writer.Write(bytes)
	return err
}
