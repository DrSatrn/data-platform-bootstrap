// Package opsview now also exposes a small HTTP seam so the frontend
// management console can consume backend-owned operator summaries without
// duplicating grouping logic in the browser.
package opsview

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/streanor/data-platform/backend/internal/orchestration"
	"github.com/streanor/data-platform/backend/internal/shared"
	"github.com/streanor/data-platform/backend/internal/storage"
)

// RunStore is the narrow run-history dependency needed for the opsview API.
type RunStore interface {
	ListPipelineRuns() ([]orchestration.PipelineRun, error)
}

// ArtifactStore is the narrow artifact-listing dependency needed for the
// opsview API.
type ArtifactStore interface {
	ListRunArtifacts(runID string) ([]storage.Artifact, error)
}

// Handler serves backend-owned operator snapshots for the management console.
type Handler struct {
	runs      RunStore
	artifacts ArtifactStore
}

// NewHandler constructs the opsview API handler.
func NewHandler(runs RunStore, artifacts ArtifactStore) http.Handler {
	return &Handler{runs: runs, artifacts: artifacts}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		shared.WriteJSON(w, http.StatusMethodNotAllowed, map[string]any{
			"error": "method not allowed",
		})
		return
	}

	limit := parseOpsviewLimit(r.URL.Query().Get("limit"))
	runs, err := h.runs.ListPipelineRuns()
	if err != nil {
		shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
			"error": "failed to load run history",
		})
		return
	}
	sort.Slice(runs, func(left, right int) bool {
		return runs[left].UpdatedAt.After(runs[right].UpdatedAt)
	})
	if len(runs) > limit {
		runs = runs[:limit]
	}

	snapshots := make([]RunOperatorSnapshot, 0, len(runs))
	allExternalToolRuns := []ExternalToolRunSummary{}
	for _, run := range runs {
		artifacts, err := h.artifacts.ListRunArtifacts(run.ID)
		if err != nil {
			shared.WriteJSON(w, http.StatusInternalServerError, map[string]any{
				"error": "failed to load run artifacts",
			})
			return
		}
		snapshot := BuildRunOperatorSnapshot(run, artifacts)
		snapshots = append(snapshots, snapshot)
		allExternalToolRuns = append(allExternalToolRuns, snapshot.ExternalToolRuns...)
	}

	shared.WriteJSON(w, http.StatusOK, map[string]any{
		"snapshots":               snapshots,
		"attention_rollup":        BuildAttentionRollup(snapshots),
		"external_tool_attention": BuildAttentionSummary(allExternalToolRuns),
	})
}

func parseOpsviewLimit(raw string) int {
	if raw == "" {
		return 8
	}
	parsed, err := strconv.Atoi(raw)
	if err != nil || parsed <= 0 {
		return 8
	}
	if parsed > 25 {
		return 25
	}
	return parsed
}
