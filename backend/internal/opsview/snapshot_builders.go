package opsview

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/storage"
)

// BuildRunOperatorSnapshot creates a compact operator-facing snapshot for one
// pipeline run.
func BuildRunOperatorSnapshot(run orchestration.PipelineRun, artifacts []storage.Artifact) RunOperatorSnapshot {
	externalToolRuns := BuildExternalToolRunSummaries(run, artifacts)
	evidenceGroups := BuildEvidenceGroups(artifacts)

	snapshot := RunOperatorSnapshot{
		RunID:               run.ID,
		PipelineID:          run.PipelineID,
		Status:              string(run.Status),
		Trigger:             run.Trigger,
		UpdatedAt:           run.UpdatedAt,
		JobCount:            len(run.JobRuns),
		FailedJobCount:      countFailedJobs(run.JobRuns),
		ExternalToolRuns:    externalToolRuns,
		EvidenceGroups:      evidenceGroups,
		Attention:           BuildAttentionSummary(externalToolRuns),
		HasExternalToolRuns: len(externalToolRuns) > 0,
	}
	return snapshot
}

// BuildEvidenceGroups groups artifacts into stable operator-facing evidence
// buckets while preserving their original relative paths.
func BuildEvidenceGroups(artifacts []storage.Artifact) []EvidenceGroup {
	groups := map[string]*EvidenceGroup{}
	for _, artifact := range artifacts {
		key, kind, label := classifyEvidenceArtifact(artifact.RelativePath)
		group, ok := groups[key]
		if !ok {
			group = &EvidenceGroup{
				Key:           key,
				Kind:          kind,
				Label:         label,
				ArtifactPaths: []string{},
			}
			groups[key] = group
		}
		group.ArtifactPaths = append(group.ArtifactPaths, artifact.RelativePath)
	}

	out := make([]EvidenceGroup, 0, len(groups))
	for _, group := range groups {
		sort.Strings(group.ArtifactPaths)
		group.ArtifactCount = len(group.ArtifactPaths)
		out = append(out, *group)
	}
	sort.Slice(out, func(left, right int) bool {
		return out[left].Key < out[right].Key
	})
	return out
}

// BuildAttentionRollup collapses multiple run snapshots into an overview
// suitable for future management-console summary panels.
func BuildAttentionRollup(snapshots []RunOperatorSnapshot) AttentionRollup {
	rollup := AttentionRollup{TotalRuns: len(snapshots)}
	for _, snapshot := range snapshots {
		switch snapshot.Status {
		case "failed":
			rollup.FailedRuns++
		case "running":
			rollup.RunningRuns++
		case "succeeded":
			rollup.SucceededRuns++
		}
		if snapshot.FailedJobCount > 0 || snapshot.Attention.FailedJobs > 0 {
			rollup.RunsWithExternalToolFailures++
		}
		if len(snapshot.EvidenceGroups) == 0 {
			rollup.RunsMissingEvidence++
		}
		rollup.ExternalToolJobCount += len(snapshot.ExternalToolRuns)
	}
	return rollup
}

func countFailedJobs(jobRuns []orchestration.JobRun) int {
	count := 0
	for _, jobRun := range jobRuns {
		if jobRun.Status == orchestration.RunStatusFailed {
			count++
		}
	}
	return count
}

func classifyEvidenceArtifact(path string) (key, kind, label string) {
	clean := filepath.ToSlash(filepath.Clean(path))
	parts := strings.Split(clean, "/")
	if len(parts) >= 3 && parts[0] == "external_tools" {
		jobID := parts[1]
		if parts[2] == "logs" {
			return "external_tool_logs:" + jobID, "external_tool_logs", "External Tool Logs: " + jobID
		}
		return "external_tool_outputs:" + jobID, "external_tool_outputs", "External Tool Outputs: " + jobID
	}
	namespace := parts[0]
	if namespace == "." || namespace == "" {
		namespace = "artifacts"
	}
	return "artifacts:" + namespace, "artifacts", "Artifacts: " + namespace
}
