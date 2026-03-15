// platformctl provides operator-facing administrative commands such as manifest
// validation, migrations, remote admin actions, and benchmark execution. The
// CLI keeps operational behavior scriptable without pushing every task through
// the browser.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/streanor/data-platform/backend/internal/analytics"
	"github.com/streanor/data-platform/backend/internal/audit"
	"github.com/streanor/data-platform/backend/internal/backup"
	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/db"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/quality"
	"github.com/streanor/data-platform/backend/internal/reporting"
	"gopkg.in/yaml.v3"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("usage: platformctl validate-manifests | migrate | benchmark | backup | remote [--server URL] [--token TOKEN] <command>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "validate-manifests":
		if err := validateManifests(); err != nil {
			fmt.Fprintf(os.Stderr, "manifest validation failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("manifest validation passed")
	case "migrate":
		if err := migrate(); err != nil {
			fmt.Fprintf(os.Stderr, "migration failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("migrations applied")
	case "benchmark":
		if err := runBenchmark(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "benchmark failed: %v\n", err)
			os.Exit(1)
		}
	case "backup":
		if err := runBackup(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "backup command failed: %v\n", err)
			os.Exit(1)
		}
	case "remote":
		if err := runRemote(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "remote command failed: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}

type benchmarkTarget struct {
	Name   string `json:"name"`
	Method string `json:"method"`
	Path   string `json:"path"`
	Body   string `json:"body,omitempty"`
	Token  bool   `json:"token,omitempty"`
}

type benchmarkResult struct {
	Name       string  `json:"name"`
	Method     string  `json:"method"`
	Path       string  `json:"path"`
	Iterations int     `json:"iterations"`
	Successes  int     `json:"successes"`
	Failures   int     `json:"failures"`
	AverageMS  float64 `json:"average_ms"`
	P50MS      float64 `json:"p50_ms"`
	P95MS      float64 `json:"p95_ms"`
	MinMS      float64 `json:"min_ms"`
	MaxMS      float64 `json:"max_ms"`
	LastStatus int     `json:"last_status"`
	LastError  string  `json:"last_error,omitempty"`
}

type benchmarkReport struct {
	ServerURL   string            `json:"server_url"`
	GeneratedAt time.Time         `json:"generated_at"`
	Iterations  int               `json:"iterations"`
	Results     []benchmarkResult `json:"results"`
}

func validateManifests() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	loader := manifests.NewLoader(cfg.ManifestRoot)
	pipelines, err := loader.LoadPipelines()
	if err != nil {
		return err
	}
	assets, err := loader.LoadAssets()
	if err != nil {
		return err
	}
	metrics, err := loader.LoadMetrics()
	if err != nil {
		return err
	}
	owners, err := loadYAMLFiles[metadata.Owner](filepath.Join(cfg.ManifestRoot, "owners", "*.yaml"))
	if err != nil {
		return err
	}
	qualityChecks, err := loadYAMLFiles[quality.Definition](filepath.Join(cfg.ManifestRoot, "quality", "*.yaml"))
	if err != nil {
		return err
	}
	dashboards, err := loadYAMLFiles[reporting.Dashboard](filepath.Join(cfg.DashboardRoot, "*.yaml"))
	if err != nil {
		return err
	}
	repoRoot := filepath.Dir(filepath.Dir(filepath.Clean(cfg.ManifestRoot)))

	ownerByID := map[string]metadata.Owner{}
	for _, owner := range owners {
		ownerByID[owner.ID] = owner
	}
	assetByID := map[string]metadata.DataAsset{}
	for _, asset := range assets {
		assetByID[asset.ID] = asset
	}
	metricByID := map[string]analytics.MetricDefinition{}
	for _, metric := range metrics {
		metricByID[metric.ID] = metric
	}
	qualityByID := map[string]quality.Definition{}
	for _, check := range qualityChecks {
		qualityByID[check.ID] = check
	}

	errors := []string{}
	errors = append(errors, validateOwners(ownerByID)...)
	errors = append(errors, validateAssets(assets, ownerByID, qualityByID, assetByID)...)
	errors = append(errors, validateMetrics(metrics, ownerByID, assetByID)...)
	errors = append(errors, validateQualityChecks(qualityChecks, assetByID)...)
	errors = append(errors, validatePipelines(pipelines, ownerByID, assetByID, metricByID, qualityByID, cfg, repoRoot)...)
	errors = append(errors, validateDashboards(dashboards, assetByID, metricByID)...)

	if len(errors) > 0 {
		return fmt.Errorf("%d validation errors:\n- %s", len(errors), strings.Join(errors, "\n- "))
	}
	return nil
}

func validateOwners(owners map[string]metadata.Owner) []string {
	errors := []string{}
	if len(owners) == 0 {
		return append(errors, "at least one owner manifest is required")
	}
	for id, owner := range owners {
		if strings.TrimSpace(id) == "" {
			errors = append(errors, "owner manifest is missing id")
		}
		if strings.TrimSpace(owner.DisplayName) == "" {
			errors = append(errors, fmt.Sprintf("owner %s is missing display_name", id))
		}
	}
	return errors
}

func validateAssets(assets []metadata.DataAsset, owners map[string]metadata.Owner, qualityChecks map[string]quality.Definition, knownAssets map[string]metadata.DataAsset) []string {
	errors := []string{}
	seenAssets := map[string]struct{}{}
	for _, asset := range assets {
		if asset.ID == "" {
			errors = append(errors, "asset manifest is missing id")
			continue
		}
		if _, exists := seenAssets[asset.ID]; exists {
			errors = append(errors, fmt.Sprintf("asset %s is defined more than once", asset.ID))
		}
		seenAssets[asset.ID] = struct{}{}
		if _, ok := owners[asset.Owner]; !ok {
			errors = append(errors, fmt.Sprintf("asset %s references unknown owner %s", asset.ID, asset.Owner))
		}
		if asset.Layer == "" {
			errors = append(errors, fmt.Sprintf("asset %s is missing layer", asset.ID))
		}
		if asset.Kind == "" {
			errors = append(errors, fmt.Sprintf("asset %s is missing kind", asset.ID))
		}
		if _, err := time.ParseDuration(asset.Freshness.ExpectedWithin); err != nil {
			errors = append(errors, fmt.Sprintf("asset %s has invalid freshness.expected_within: %v", asset.ID, err))
		}
		if _, err := time.ParseDuration(asset.Freshness.WarnAfter); err != nil {
			errors = append(errors, fmt.Sprintf("asset %s has invalid freshness.warn_after: %v", asset.ID, err))
		}
		columnNames := map[string]struct{}{}
		for _, column := range asset.Columns {
			if column.Name == "" {
				errors = append(errors, fmt.Sprintf("asset %s contains a column without a name", asset.ID))
				continue
			}
			if _, exists := columnNames[column.Name]; exists {
				errors = append(errors, fmt.Sprintf("asset %s defines column %s more than once", asset.ID, column.Name))
			}
			columnNames[column.Name] = struct{}{}
		}
		for _, checkID := range asset.QualityCheckRefs {
			if _, ok := qualityChecks[checkID]; !ok {
				errors = append(errors, fmt.Sprintf("asset %s references unknown quality check %s", asset.ID, checkID))
			}
		}
		for _, sourceRef := range asset.SourceRefs {
			if strings.HasPrefix(sourceRef, "sample.") || strings.HasPrefix(sourceRef, "api.") || strings.HasPrefix(sourceRef, "db.") {
				continue
			}
			if _, ok := knownAssets[sourceRef]; !ok {
				errors = append(errors, fmt.Sprintf("asset %s references unknown source_ref %s", asset.ID, sourceRef))
			}
		}
	}
	return errors
}

func validateMetrics(metrics []analytics.MetricDefinition, owners map[string]metadata.Owner, assets map[string]metadata.DataAsset) []string {
	errors := []string{}
	seenMetrics := map[string]struct{}{}
	for _, metric := range metrics {
		if metric.ID == "" {
			errors = append(errors, "metric manifest is missing id")
			continue
		}
		if _, exists := seenMetrics[metric.ID]; exists {
			errors = append(errors, fmt.Sprintf("metric %s is defined more than once", metric.ID))
		}
		seenMetrics[metric.ID] = struct{}{}
		if _, ok := owners[metric.Owner]; !ok {
			errors = append(errors, fmt.Sprintf("metric %s references unknown owner %s", metric.ID, metric.Owner))
		}
		asset, ok := assets[metric.DatasetRef]
		if !ok {
			errors = append(errors, fmt.Sprintf("metric %s references unknown dataset %s", metric.ID, metric.DatasetRef))
			continue
		}
		if !assetHasColumn(asset, metric.TimeDimension) {
			errors = append(errors, fmt.Sprintf("metric %s references missing time_dimension %s on asset %s", metric.ID, metric.TimeDimension, asset.ID))
		}
		for _, dimension := range metric.Dimensions {
			if !assetHasColumn(asset, dimension) {
				errors = append(errors, fmt.Sprintf("metric %s dimension %s does not exist on asset %s", metric.ID, dimension, asset.ID))
			}
		}
		for _, measure := range metric.Measures {
			if !assetHasColumn(asset, measure) {
				errors = append(errors, fmt.Sprintf("metric %s measure %s does not exist on asset %s", metric.ID, measure, asset.ID))
			}
		}
		if metric.DefaultVisualization == "" {
			errors = append(errors, fmt.Sprintf("metric %s is missing default_visualization", metric.ID))
		}
	}
	return errors
}

func validateQualityChecks(checks []quality.Definition, assets map[string]metadata.DataAsset) []string {
	errors := []string{}
	seenChecks := map[string]struct{}{}
	for _, check := range checks {
		if check.ID == "" {
			errors = append(errors, "quality manifest is missing id")
			continue
		}
		if _, exists := seenChecks[check.ID]; exists {
			errors = append(errors, fmt.Sprintf("quality check %s is defined more than once", check.ID))
		}
		seenChecks[check.ID] = struct{}{}
		asset, ok := assets[check.AssetRef]
		if !ok {
			errors = append(errors, fmt.Sprintf("quality check %s references unknown asset %s", check.ID, check.AssetRef))
			continue
		}
		if check.ColumnRef != "" && !assetHasColumn(asset, check.ColumnRef) {
			errors = append(errors, fmt.Sprintf("quality check %s references missing column %s on asset %s", check.ID, check.ColumnRef, asset.ID))
		}
		if check.Type == "" {
			errors = append(errors, fmt.Sprintf("quality check %s is missing type", check.ID))
		}
	}
	return errors
}

func validatePipelines(
	pipelines []orchestration.Pipeline,
	owners map[string]metadata.Owner,
	assets map[string]metadata.DataAsset,
	metrics map[string]analytics.MetricDefinition,
	qualityChecks map[string]quality.Definition,
	cfg config.Settings,
	repoRoot string,
) []string {
	errors := []string{}
	seenPipelines := map[string]struct{}{}
	for _, pipeline := range pipelines {
		if _, exists := seenPipelines[pipeline.ID]; exists {
			errors = append(errors, fmt.Sprintf("pipeline %s is defined more than once", pipeline.ID))
		}
		seenPipelines[pipeline.ID] = struct{}{}
		if _, ok := owners[pipeline.Owner]; !ok {
			errors = append(errors, fmt.Sprintf("pipeline %s references unknown owner %s", pipeline.ID, pipeline.Owner))
		}
		if err := validatePipelineStructure(pipeline); err != nil {
			errors = append(errors, fmt.Sprintf("pipeline %s failed structural validation: %v", pipeline.ID, err))
		}
		for _, job := range pipeline.Jobs {
			if job.Retries < 0 {
				errors = append(errors, fmt.Sprintf("pipeline %s job %s has negative retries", pipeline.ID, job.ID))
			}
			switch job.Type {
			case orchestration.JobTypeIngest:
				for _, output := range job.Outputs {
					if _, ok := assets[output]; !ok {
						errors = append(errors, fmt.Sprintf("pipeline %s ingest job %s outputs unknown asset %s", pipeline.ID, job.ID, output))
					}
				}
			case orchestration.JobTypeTransformSQL:
				if job.TransformRef == "" {
					errors = append(errors, fmt.Sprintf("pipeline %s job %s is missing transform_ref", pipeline.ID, job.ID))
				} else if !sqlTransformExists(cfg.SQLRoot, job.TransformRef) {
					errors = append(errors, fmt.Sprintf("pipeline %s job %s references missing SQL transform %s", pipeline.ID, job.ID, job.TransformRef))
				}
				for _, output := range job.Outputs {
					if _, ok := assets[output]; !ok {
						errors = append(errors, fmt.Sprintf("pipeline %s SQL job %s outputs unknown asset %s", pipeline.ID, job.ID, output))
					}
				}
			case orchestration.JobTypeTransformPy:
				if job.Command == "" {
					errors = append(errors, fmt.Sprintf("pipeline %s job %s is missing python command", pipeline.ID, job.ID))
				} else if !pythonTaskExists(cfg.PythonTaskRoot, job.Command) {
					errors = append(errors, fmt.Sprintf("pipeline %s job %s references missing python task in command %q", pipeline.ID, job.ID, job.Command))
				}
				for _, output := range job.Outputs {
					if _, ok := assets[output]; !ok {
						errors = append(errors, fmt.Sprintf("pipeline %s Python job %s outputs unknown asset %s", pipeline.ID, job.ID, output))
					}
				}
			case orchestration.JobTypeQualityCheck:
				if _, ok := qualityChecks[job.ID]; !ok {
					errors = append(errors, fmt.Sprintf("pipeline %s quality job %s has no matching quality manifest", pipeline.ID, job.ID))
				}
				if !sqlQualityExists(cfg.SQLRoot, job.ID) {
					errors = append(errors, fmt.Sprintf("pipeline %s quality job %s references missing SQL quality query", pipeline.ID, job.ID))
				}
			case orchestration.JobTypePublishMetric:
				for _, output := range job.Outputs {
					if _, ok := metrics[output]; !ok {
						errors = append(errors, fmt.Sprintf("pipeline %s metric job %s outputs unknown metric %s", pipeline.ID, job.ID, output))
						continue
					}
					if !sqlMetricExists(cfg.SQLRoot, output) {
						errors = append(errors, fmt.Sprintf("pipeline %s metric job %s references missing metric SQL for %s", pipeline.ID, job.ID, output))
					}
				}
			case orchestration.JobTypeExternalTool:
				if job.ExternalTool == nil {
					errors = append(errors, fmt.Sprintf("pipeline %s external_tool job %s is missing external_tool config", pipeline.ID, job.ID))
					continue
				}
				if strings.TrimSpace(job.ExternalTool.Tool) == "" {
					errors = append(errors, fmt.Sprintf("pipeline %s external_tool job %s is missing external_tool.tool", pipeline.ID, job.ID))
				}
				if strings.TrimSpace(job.ExternalTool.Action) == "" {
					errors = append(errors, fmt.Sprintf("pipeline %s external_tool job %s is missing external_tool.action", pipeline.ID, job.ID))
				}
				if !repoPathExists(repoRoot, job.ExternalTool.ProjectRef) {
					errors = append(errors, fmt.Sprintf("pipeline %s external_tool job %s references missing project_ref %s", pipeline.ID, job.ID, job.ExternalTool.ProjectRef))
				}
				if !repoPathExists(repoRoot, job.ExternalTool.ConfigRef) {
					errors = append(errors, fmt.Sprintf("pipeline %s external_tool job %s references missing config_ref %s", pipeline.ID, job.ID, job.ExternalTool.ConfigRef))
				}
				if len(job.ExternalTool.Artifacts) == 0 {
					errors = append(errors, fmt.Sprintf("pipeline %s external_tool job %s must declare at least one artifact", pipeline.ID, job.ID))
				}
			default:
				errors = append(errors, fmt.Sprintf("pipeline %s job %s uses unsupported type %s", pipeline.ID, job.ID, job.Type))
			}

			for _, input := range job.Inputs {
				if strings.HasPrefix(input, "sample.") {
					continue
				}
				if _, assetExists := assets[input]; assetExists {
					continue
				}
				if _, metricExists := metrics[input]; metricExists {
					continue
				}
				errors = append(errors, fmt.Sprintf("pipeline %s job %s references unknown input %s", pipeline.ID, job.ID, input))
			}
		}
	}
	return errors
}

func validatePipelineStructure(pipeline orchestration.Pipeline) error {
	if strings.TrimSpace(pipeline.ID) == "" {
		return fmt.Errorf("pipeline id is required")
	}
	if len(pipeline.Jobs) == 0 {
		return fmt.Errorf("pipeline must contain at least one job")
	}
	seenJobs := map[string]struct{}{}
	for _, job := range pipeline.Jobs {
		if strings.TrimSpace(job.ID) == "" {
			return fmt.Errorf("job id is required")
		}
		if _, exists := seenJobs[job.ID]; exists {
			return fmt.Errorf("duplicate job id %q", job.ID)
		}
		seenJobs[job.ID] = struct{}{}
	}
	for _, job := range pipeline.Jobs {
		for _, dependency := range job.DependsOn {
			if _, ok := seenJobs[dependency]; !ok {
				return fmt.Errorf("job %q depends on unknown job %q", job.ID, dependency)
			}
			if dependency == job.ID {
				return fmt.Errorf("job %q cannot depend on itself", job.ID)
			}
		}
	}
	return nil
}

func validateDashboards(dashboards []reporting.Dashboard, assets map[string]metadata.DataAsset, metrics map[string]analytics.MetricDefinition) []string {
	errors := []string{}
	seenDashboards := map[string]struct{}{}
	for _, dashboard := range dashboards {
		if _, exists := seenDashboards[dashboard.ID]; exists {
			errors = append(errors, fmt.Sprintf("dashboard %s is defined more than once", dashboard.ID))
		}
		seenDashboards[dashboard.ID] = struct{}{}
		if err := reporting.ValidateDashboardDefinition(dashboard); err != nil {
			errors = append(errors, fmt.Sprintf("dashboard %s failed validation: %v", dashboard.ID, err))
			continue
		}
		widgetIDs := map[string]struct{}{}
		for _, widget := range dashboard.Widgets {
			if _, exists := widgetIDs[widget.ID]; exists {
				errors = append(errors, fmt.Sprintf("dashboard %s widget %s is defined more than once", dashboard.ID, widget.ID))
			}
			widgetIDs[widget.ID] = struct{}{}
			if widget.DatasetRef != "" {
				if _, ok := assets[widget.DatasetRef]; !ok {
					errors = append(errors, fmt.Sprintf("dashboard %s widget %s references unknown dataset %s", dashboard.ID, widget.ID, widget.DatasetRef))
				}
			}
			if widget.MetricRef != "" {
				if _, ok := metrics[widget.MetricRef]; !ok {
					errors = append(errors, fmt.Sprintf("dashboard %s widget %s references unknown metric %s", dashboard.ID, widget.ID, widget.MetricRef))
				}
			}
		}
	}
	return errors
}

func assetHasColumn(asset metadata.DataAsset, name string) bool {
	for _, column := range asset.Columns {
		if column.Name == name {
			return true
		}
	}
	return false
}

func sqlTransformExists(sqlRoot, transformRef string) bool {
	name := strings.TrimPrefix(transformRef, "transform.")
	if name == transformRef {
		return false
	}
	_, err := os.Stat(filepath.Join(sqlRoot, "transforms", name+".sql"))
	return err == nil
}

func sqlMetricExists(sqlRoot, metricID string) bool {
	_, err := os.Stat(filepath.Join(sqlRoot, "metrics", metricID+".sql"))
	return err == nil
}

func sqlQualityExists(sqlRoot, checkID string) bool {
	_, err := os.Stat(filepath.Join(sqlRoot, "quality", checkID+".sql"))
	return err == nil
}

func pythonTaskExists(taskRoot, command string) bool {
	tokens := strings.Fields(strings.TrimSpace(command))
	if len(tokens) < 2 {
		return false
	}
	scriptPath := tokens[1]
	if !filepath.IsAbs(scriptPath) {
		scriptPath = filepath.Join(taskRoot, filepath.Clean(scriptPath))
	}
	_, err := os.Stat(scriptPath)
	return err == nil
}

func repoPathExists(repoRoot, path string) bool {
	cleaned := filepath.Clean(strings.TrimSpace(path))
	if cleaned == "." || cleaned == "" || filepath.IsAbs(cleaned) || strings.HasPrefix(cleaned, "..") {
		return false
	}
	_, err := os.Stat(filepath.Join(repoRoot, filepath.FromSlash(cleaned)))
	return err == nil
}

func loadYAMLFiles[T any](pattern string) ([]T, error) {
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("glob yaml files: %w", err)
	}
	out := make([]T, 0, len(matches))
	for _, match := range matches {
		bytes, err := os.ReadFile(match)
		if err != nil {
			return nil, fmt.Errorf("read yaml file %s: %w", match, err)
		}
		var item T
		if err := yaml.Unmarshal(bytes, &item); err != nil {
			return nil, fmt.Errorf("decode yaml file %s: %w", match, err)
		}
		out = append(out, item)
	}
	return out, nil
}

func migrate() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	conn, err := db.Open(cfg.PostgresDSN)
	if err != nil {
		return err
	}
	defer conn.Close()

	ctx := context.Background()
	if err := db.ApplyMigrations(ctx, conn, cfg.MigrationsRoot); err != nil {
		return err
	}
	return nil
}

func runBackup(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: platformctl backup create [--out PATH] | list | verify --file PATH | restore --file PATH --yes [--postgres-mode auto|required|skip]")
	}

	service, closeFn, err := newBackupService()
	if err != nil {
		return err
	}
	if closeFn != nil {
		defer closeFn()
	}

	switch args[0] {
	case "create":
		flagSet := flag.NewFlagSet("backup create", flag.ContinueOnError)
		out := flagSet.String("out", "", "optional backup bundle output path")
		if err := flagSet.Parse(args[1:]); err != nil {
			return err
		}
		result, err := service.Create(*out)
		if err != nil {
			return err
		}
		fmt.Printf("backup bundle created: %s\n", result.Path)
		fmt.Printf("pipeline_runs=%d dashboards=%d data_assets=%d bundle_files=%d\n",
			result.Manifest.Counts.PipelineRuns,
			result.Manifest.Counts.Dashboards,
			result.Manifest.Counts.DataAssets,
			result.Manifest.Counts.BundleFiles,
		)
		return nil
	case "list":
		bundles, err := service.ListBundles()
		if err != nil {
			return err
		}
		if len(bundles) == 0 {
			fmt.Println("no backup bundles recorded yet")
			return nil
		}
		for _, bundle := range bundles {
			fmt.Printf("%s | %d bytes\n", bundle.Path, bundle.SizeBytes)
		}
		return nil
	case "verify":
		flagSet := flag.NewFlagSet("backup verify", flag.ContinueOnError)
		filePath := flagSet.String("file", "", "backup bundle to verify")
		if err := flagSet.Parse(args[1:]); err != nil {
			return err
		}
		if *filePath == "" {
			return fmt.Errorf("--file is required")
		}
		manifest, err := service.Verify(*filePath)
		if err != nil {
			return err
		}
		fmt.Printf("backup bundle verified: %s\n", *filePath)
		fmt.Printf("generated_at=%s pipeline_runs=%d queue_requests=%d dashboards=%d data_assets=%d bundle_files=%d\n",
			manifest.GeneratedAt.Format(time.RFC3339),
			manifest.Counts.PipelineRuns,
			manifest.Counts.QueueRequests,
			manifest.Counts.Dashboards,
			manifest.Counts.DataAssets,
			manifest.Counts.BundleFiles,
		)
		return nil
	case "restore":
		flagSet := flag.NewFlagSet("backup restore", flag.ContinueOnError)
		filePath := flagSet.String("file", "", "backup bundle to restore")
		yes := flagSet.Bool("yes", false, "confirm overwriting the target runtime roots")
		postgresMode := flagSet.String("postgres-mode", string(backup.PostgresRestoreAuto), "postgres restore behavior: auto|required|skip")
		dataRoot := flagSet.String("data-root", "", "override restore target data root")
		artifactRoot := flagSet.String("artifact-root", "", "override restore target artifact root")
		duckdbPath := flagSet.String("duckdb-path", "", "override restore target DuckDB file")
		extractRoot := flagSet.String("extract-root", "", "optional extraction workspace to keep after restore")
		if err := flagSet.Parse(args[1:]); err != nil {
			return err
		}
		if *filePath == "" {
			return fmt.Errorf("--file is required")
		}
		result, err := service.Restore(backup.RestoreOptions{
			BundlePath:         *filePath,
			Confirm:            *yes,
			TargetDataRoot:     *dataRoot,
			TargetArtifactRoot: *artifactRoot,
			TargetDuckDBPath:   *duckdbPath,
			ExtractRoot:        *extractRoot,
			PostgresMode:       backup.PostgresRestoreMode(strings.ToLower(strings.TrimSpace(*postgresMode))),
		})
		if err != nil {
			return err
		}
		fmt.Printf("backup bundle restored: %s\n", result.BundlePath)
		fmt.Printf("data_root=%s artifact_root=%s duckdb_path=%s postgres_restored=%t requeued_requests=%d\n",
			result.DataRoot,
			result.ArtifactRoot,
			result.DuckDBPath,
			result.PostgresRestored,
			result.QueueRequestsRequeued,
		)
		for _, warning := range result.Warnings {
			fmt.Printf("warning=%s\n", warning)
		}
		return nil
	default:
		return fmt.Errorf("usage: platformctl backup create [--out PATH] | list | verify --file PATH | restore --file PATH --yes [--postgres-mode auto|required|skip]")
	}
}

func runRemote(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	flagSet := flag.NewFlagSet("remote", flag.ContinueOnError)
	server := flagSet.String("server", cfg.APIBaseURL, "platform API base URL")
	token := flagSet.String("token", cfg.AdminToken, "admin token")
	if err := flagSet.Parse(args); err != nil {
		return err
	}

	command := strings.TrimSpace(strings.Join(flagSet.Args(), " "))
	if command == "" {
		return fmt.Errorf("remote command is required")
	}

	body, err := json.Marshal(map[string]string{"command": command})
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodPost, strings.TrimRight(*server, "/")+"/api/v1/admin/terminal/execute", bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	if *token != "" {
		request.Header.Set("Authorization", "Bearer "+*token)
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	payload, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode >= 400 {
		return fmt.Errorf("status %d: %s", response.StatusCode, strings.TrimSpace(string(payload)))
	}

	var result struct {
		Output []string `json:"output"`
	}
	if err := json.Unmarshal(payload, &result); err != nil {
		return fmt.Errorf("decode remote response: %w", err)
	}

	for _, line := range result.Output {
		fmt.Println(line)
	}
	return nil
}

func runBenchmark(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	flagSet := flag.NewFlagSet("benchmark", flag.ContinueOnError)
	server := flagSet.String("server", cfg.APIBaseURL, "platform API base URL")
	token := flagSet.String("token", cfg.AdminToken, "admin token")
	iterations := flagSet.Int("iterations", 5, "iterations per target")
	out := flagSet.String("out", "", "optional file to write JSON benchmark report")
	if err := flagSet.Parse(args); err != nil {
		return err
	}
	if *iterations <= 0 {
		return fmt.Errorf("iterations must be positive")
	}

	report := benchmarkReport{
		ServerURL:   strings.TrimRight(*server, "/"),
		GeneratedAt: time.Now().UTC(),
		Iterations:  *iterations,
		Results:     make([]benchmarkResult, 0, len(defaultBenchmarkTargets())),
	}
	client := &http.Client{Timeout: 15 * time.Second}
	for _, target := range defaultBenchmarkTargets() {
		result := benchmarkResult{
			Name:       target.Name,
			Method:     target.Method,
			Path:       target.Path,
			Iterations: *iterations,
		}
		durations := make([]float64, 0, *iterations)
		for index := 0; index < *iterations; index++ {
			elapsed, statusCode, err := benchmarkRequest(client, report.ServerURL, *token, target)
			result.LastStatus = statusCode
			if err != nil {
				result.Failures++
				result.LastError = err.Error()
				continue
			}
			result.Successes++
			durations = append(durations, elapsed)
		}
		if len(durations) > 0 {
			sort.Float64s(durations)
			result.AverageMS = average(durations)
			result.P50MS = percentile(durations, 50)
			result.P95MS = percentile(durations, 95)
			result.MinMS = durations[0]
			result.MaxMS = durations[len(durations)-1]
		}
		report.Results = append(report.Results, result)
	}

	for _, result := range report.Results {
		fmt.Printf(
			"%s %s%s success=%d/%d p50=%.2fms p95=%.2fms avg=%.2fms\n",
			result.Method,
			report.ServerURL,
			result.Path,
			result.Successes,
			result.Iterations,
			result.P50MS,
			result.P95MS,
			result.AverageMS,
		)
		if result.LastError != "" {
			fmt.Printf("  last_error=%s\n", result.LastError)
		}
	}

	benchmarkFailed := false
	for _, result := range report.Results {
		if result.Successes == 0 {
			benchmarkFailed = true
			break
		}
	}

	if *out != "" {
		if err := os.MkdirAll(filepath.Dir(*out), 0o755); err != nil {
			return err
		}
		bytes, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(*out, bytes, 0o644); err != nil {
			return err
		}
		fmt.Printf("benchmark report written to %s\n", *out)
	}

	if benchmarkFailed {
		return fmt.Errorf("one or more benchmark targets recorded zero successful requests")
	}

	return nil
}

func defaultBenchmarkTargets() []benchmarkTarget {
	return []benchmarkTarget{
		{Name: "health", Method: http.MethodGet, Path: "/healthz"},
		{Name: "catalog", Method: http.MethodGet, Path: "/api/v1/catalog", Token: true},
		{Name: "analytics_dataset", Method: http.MethodGet, Path: "/api/v1/analytics?dataset=mart_budget_vs_actual", Token: true},
		{Name: "analytics_metric", Method: http.MethodGet, Path: "/api/v1/analytics?metric=metrics_category_variance", Token: true},
		{Name: "reports", Method: http.MethodGet, Path: "/api/v1/reports", Token: true},
		{Name: "system_overview", Method: http.MethodGet, Path: "/api/v1/system/overview", Token: true},
		{Name: "system_audit", Method: http.MethodGet, Path: "/api/v1/system/audit", Token: true},
		{Name: "admin_status", Method: http.MethodPost, Path: "/api/v1/admin/terminal/execute", Body: `{"command":"status"}`, Token: true},
	}
}

func benchmarkRequest(client *http.Client, server, token string, target benchmarkTarget) (float64, int, error) {
	var body io.Reader
	if target.Body != "" {
		body = strings.NewReader(target.Body)
	}
	request, err := http.NewRequest(target.Method, server+target.Path, body)
	if err != nil {
		return 0, 0, err
	}
	if target.Body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	if target.Token && token != "" {
		request.Header.Set("Authorization", "Bearer "+token)
	}

	started := time.Now()
	response, err := client.Do(request)
	elapsed := time.Since(started).Seconds() * 1000
	if err != nil {
		return elapsed, 0, err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, response.Body)
	if response.StatusCode >= 400 {
		return elapsed, response.StatusCode, fmt.Errorf("status %d", response.StatusCode)
	}
	return elapsed, response.StatusCode, nil
}

func average(values []float64) float64 {
	total := 0.0
	for _, value := range values {
		total += value
	}
	return total / float64(len(values))
}

func percentile(values []float64, target int) float64 {
	if len(values) == 0 {
		return 0
	}
	index := (float64(target) / 100) * float64(len(values)-1)
	lower := int(index)
	upper := lower
	if lower < len(values)-1 {
		upper = lower + 1
	}
	weight := index - float64(lower)
	return values[lower] + (values[upper]-values[lower])*weight
}

func newBackupService() (*backup.Service, func(), error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, err
	}

	loader := manifests.NewLoader(cfg.ManifestRoot)
	runStore, err := orchestration.NewFileStore(cfg.DataRoot)
	if err != nil {
		return nil, nil, err
	}
	queue, err := orchestration.NewQueue(cfg.DataRoot)
	if err != nil {
		return nil, nil, err
	}

	var reportStore reporting.Store
	fileReports, err := reporting.NewFileStore(cfg.DataRoot, cfg.DashboardRoot)
	if err == nil {
		reportStore = fileReports
	} else {
		reportStore = reporting.NewMemoryStore()
	}

	var auditStore audit.Store
	fileAudit, err := audit.NewFileStore(cfg.DataRoot)
	if err == nil {
		auditStore = fileAudit
	} else {
		auditStore = audit.NewMemoryStore()
	}

	var metadataStore metadata.Store
	controlPlane, err := db.NewControlPlane(context.Background(), cfg.PostgresDSN)
	if err == nil {
		closeFn := func() { _ = controlPlane.Conn.Close() }
		return backup.NewService(cfg, loader, controlPlane.RunStore, controlPlane.RunQueue, controlPlane.Dashboards, controlPlane.Audit, controlPlane.Metadata), closeFn, nil
	}

	return backup.NewService(cfg, loader, runStore, queue, reportStore, auditStore, metadataStore), nil, nil
}
