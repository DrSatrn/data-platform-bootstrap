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

	"github.com/streanor/data-platform/backend/internal/audit"
	"github.com/streanor/data-platform/backend/internal/backup"
	"github.com/streanor/data-platform/backend/internal/config"
	"github.com/streanor/data-platform/backend/internal/db"
	"github.com/streanor/data-platform/backend/internal/manifests"
	"github.com/streanor/data-platform/backend/internal/metadata"
	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/reporting"
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

	for _, pipeline := range pipelines {
		if err := orchestration.ValidatePipeline(pipeline); err != nil {
			return fmt.Errorf("pipeline %s: %w", pipeline.ID, err)
		}
	}

	_, err = loader.LoadAssets()
	return err
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
		return fmt.Errorf("usage: platformctl backup create [--out PATH] | list | verify --file PATH")
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
	default:
		return fmt.Errorf("usage: platformctl backup create [--out PATH] | list | verify --file PATH")
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
		{Name: "catalog", Method: http.MethodGet, Path: "/api/v1/catalog"},
		{Name: "analytics_dataset", Method: http.MethodGet, Path: "/api/v1/analytics?dataset=mart_budget_vs_actual"},
		{Name: "analytics_metric", Method: http.MethodGet, Path: "/api/v1/analytics?metric=metrics_category_variance"},
		{Name: "reports", Method: http.MethodGet, Path: "/api/v1/reports"},
		{Name: "system_overview", Method: http.MethodGet, Path: "/api/v1/system/overview"},
		{Name: "system_audit", Method: http.MethodGet, Path: "/api/v1/system/audit"},
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
