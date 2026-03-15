package orchestration

import (
	"fmt"
	"path/filepath"
	"strings"
)

func validateJob(job Job) error {
	switch job.Type {
	case JobTypeIngest:
		return validateIngestJob(job)
	case JobTypeTransformSQL:
		return validateTransformSQLJob(job)
	case JobTypePublishMetric:
		return validatePublishMetricJob(job)
	case JobTypeExternalTool:
		return validateExternalToolJob(job)
	default:
		return nil
	}
}

func validateIngestJob(job Job) error {
	if job.Ingest == nil {
		return fmt.Errorf("ingest jobs must declare an ingest block")
	}
	if strings.TrimSpace(job.Command) != "" {
		return fmt.Errorf("ingest jobs must not set command")
	}
	if strings.TrimSpace(job.TransformRef) != "" {
		return fmt.Errorf("ingest jobs must not set transform_ref")
	}
	if !isRepoRelativeRef(job.Ingest.SourceRef) {
		return fmt.Errorf("ingest.source_ref must be repo-relative")
	}
	if !isDataRelativePath(job.Ingest.TargetPath) {
		return fmt.Errorf("ingest.target_path must be data-root-relative")
	}
	if artifactPath := strings.TrimSpace(job.Ingest.ArtifactPath); artifactPath != "" && !isDataRelativePath(artifactPath) {
		return fmt.Errorf("ingest.artifact_path must be data-root-relative")
	}
	return nil
}

func validateTransformSQLJob(job Job) error {
	if strings.TrimSpace(job.TransformRef) == "" {
		return fmt.Errorf("transform_sql jobs must set transform_ref")
	}
	for index, bootstrap := range job.Bootstrap {
		if !strings.HasPrefix(strings.TrimSpace(bootstrap.SQLRef), "bootstrap.") {
			return fmt.Errorf("bootstrap[%d].sql_ref must start with bootstrap.", index)
		}
		if strings.TrimSpace(bootstrap.Placeholder) == "" {
			return fmt.Errorf("bootstrap[%d].placeholder is required", index)
		}
		if !isDataRelativePath(bootstrap.SourcePath) {
			return fmt.Errorf("bootstrap[%d].source_path must be data-root-relative", index)
		}
	}
	return nil
}

func validatePublishMetricJob(job Job) error {
	if len(job.MetricRefs) == 0 {
		return fmt.Errorf("publish_metric jobs must declare metric_refs")
	}
	seen := make(map[string]struct{}, len(job.MetricRefs))
	for index, metricRef := range job.MetricRefs {
		metricRef = strings.TrimSpace(metricRef)
		if metricRef == "" {
			return fmt.Errorf("metric_refs[%d] is empty", index)
		}
		if _, exists := seen[metricRef]; exists {
			return fmt.Errorf("metric_refs[%d] duplicates %q", index, metricRef)
		}
		seen[metricRef] = struct{}{}
	}
	return nil
}

func isDataRelativePath(path string) bool {
	clean := filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
	if clean == "." || clean == "" {
		return false
	}
	if strings.HasPrefix(clean, "../") || strings.HasPrefix(clean, "/") {
		return false
	}
	return true
}
