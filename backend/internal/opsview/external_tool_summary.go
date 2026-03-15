package opsview

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/storage"
)

// BuildExternalToolRunSummaries groups one pipeline run's external-tool events
// and artifacts by job identifier.
func BuildExternalToolRunSummaries(run orchestration.PipelineRun, artifacts []storage.Artifact) []ExternalToolRunSummary {
	summaries := map[string]*ExternalToolRunSummary{}

	for _, event := range run.Events {
		jobID := strings.TrimSpace(event.Fields["job_id"])
		tool := strings.TrimSpace(event.Fields["tool"])
		if jobID == "" || tool == "" {
			continue
		}
		summary := getOrCreateSummary(summaries, run, jobID)
		if summary.Tool == "" {
			summary.Tool = tool
		}
		if summary.Action == "" {
			summary.Action = strings.TrimSpace(event.Fields["action"])
		}
		if failureClass := strings.TrimSpace(event.Fields["failure_class"]); failureClass != "" {
			summary.FailureClass = failureClass
		}
		eventCopy := event
		summary.Events = append(summary.Events, eventCopy)
		updateStatus(summary, event)
		updateLastEventAt(summary, event)
	}

	for _, artifact := range artifacts {
		jobID, category, ok := splitExternalToolArtifactPath(artifact.RelativePath)
		if !ok {
			continue
		}
		summary := getOrCreateSummary(summaries, run, jobID)
		artifactCopy := artifact
		if category == "logs" {
			summary.LogArtifacts = append(summary.LogArtifacts, artifactCopy)
		} else {
			summary.OutputArtifacts = append(summary.OutputArtifacts, artifactCopy)
		}
	}

	out := make([]ExternalToolRunSummary, 0, len(summaries))
	for _, summary := range summaries {
		sortArtifacts(summary.LogArtifacts)
		sortArtifacts(summary.OutputArtifacts)
		summary.Evidence = BuildOperatorEvidenceSummary(run.ID, summary.LogArtifacts, summary.OutputArtifacts)
		if summary.Status == "" {
			summary.Status = "unknown"
		}
		out = append(out, *summary)
	}
	sort.Slice(out, func(left, right int) bool {
		return out[left].JobID < out[right].JobID
	})
	return out
}

// BuildOperatorEvidenceSummary flattens log and output artifacts into a small
// operator-facing evidence model that preserves linkable paths.
func BuildOperatorEvidenceSummary(runID string, logArtifacts, outputArtifacts []storage.Artifact) OperatorEvidenceSummary {
	evidence := OperatorEvidenceSummary{
		RunID:               runID,
		LogArtifactCount:    len(logArtifacts),
		OutputArtifactCount: len(outputArtifacts),
		ArtifactPaths:       []string{},
		LogPaths:            []string{},
		OutputPaths:         []string{},
	}
	for _, artifact := range logArtifacts {
		evidence.LogPaths = append(evidence.LogPaths, artifact.RelativePath)
		evidence.ArtifactPaths = append(evidence.ArtifactPaths, artifact.RelativePath)
	}
	for _, artifact := range outputArtifacts {
		evidence.OutputPaths = append(evidence.OutputPaths, artifact.RelativePath)
		evidence.ArtifactPaths = append(evidence.ArtifactPaths, artifact.RelativePath)
	}
	sort.Strings(evidence.LogPaths)
	sort.Strings(evidence.OutputPaths)
	sort.Strings(evidence.ArtifactPaths)
	evidence.TotalArtifacts = len(evidence.ArtifactPaths)
	return evidence
}

// BuildAttentionSummary gives a compact operator summary across grouped
// external-tool jobs.
func BuildAttentionSummary(summaries []ExternalToolRunSummary) AttentionSummary {
	attention := AttentionSummary{TotalJobs: len(summaries)}
	for _, summary := range summaries {
		switch summary.Status {
		case "failed":
			attention.FailedJobs++
		case "running":
			attention.RunningJobs++
		case "succeeded":
			attention.SucceededJobs++
		}
		if len(summary.LogArtifacts) == 0 {
			attention.JobsMissingLogs++
		}
		if len(summary.OutputArtifacts) == 0 {
			attention.JobsMissingOutputs++
		}
	}
	return attention
}

func getOrCreateSummary(summaries map[string]*ExternalToolRunSummary, run orchestration.PipelineRun, jobID string) *ExternalToolRunSummary {
	if summary, ok := summaries[jobID]; ok {
		return summary
	}
	summary := &ExternalToolRunSummary{
		RunID:           run.ID,
		PipelineID:      run.PipelineID,
		JobID:           jobID,
		Status:          "unknown",
		Events:          []orchestration.RunEvent{},
		LogArtifacts:    []storage.Artifact{},
		OutputArtifacts: []storage.Artifact{},
	}
	summaries[jobID] = summary
	return summary
}

func updateStatus(summary *ExternalToolRunSummary, event orchestration.RunEvent) {
	message := strings.ToLower(strings.TrimSpace(event.Message))
	if event.Level == "error" || strings.Contains(message, "failed") {
		summary.Status = "failed"
		return
	}
	if strings.Contains(message, "finished") || strings.Contains(message, "succeeded") {
		if summary.Status != "failed" {
			summary.Status = "succeeded"
		}
		return
	}
	if strings.Contains(message, "started") || strings.Contains(message, "selected") {
		if summary.Status == "" || summary.Status == "unknown" {
			summary.Status = "running"
		}
	}
}

func updateLastEventAt(summary *ExternalToolRunSummary, event orchestration.RunEvent) {
	if event.Time.IsZero() {
		return
	}
	if summary.LastEventAt == nil || event.Time.After(*summary.LastEventAt) {
		timeCopy := event.Time
		summary.LastEventAt = &timeCopy
	}
}

func splitExternalToolArtifactPath(path string) (jobID, category string, ok bool) {
	clean := filepath.ToSlash(filepath.Clean(path))
	parts := strings.Split(clean, "/")
	if len(parts) < 3 || parts[0] != "external_tools" {
		return "", "", false
	}
	jobID = parts[1]
	if jobID == "" {
		return "", "", false
	}
	if parts[2] == "logs" {
		return jobID, "logs", true
	}
	return jobID, "outputs", true
}

func sortArtifacts(artifacts []storage.Artifact) {
	sort.Slice(artifacts, func(left, right int) bool {
		return artifacts[left].RelativePath < artifacts[right].RelativePath
	})
}
