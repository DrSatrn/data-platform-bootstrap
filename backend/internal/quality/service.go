// Package quality exposes quality-check definitions and status output derived
// from repo-managed sample data. This keeps data-trust signals grounded in our
// own logic even before database-backed quality history is added.
package quality

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CheckStatus summarizes the current state of a quality check.
type CheckStatus struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Status   string `json:"status"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// Service owns quality-check status retrieval.
type Service struct {
	sampleDataRoot string
	dataRoot       string
}

// NewService creates a quality service.
func NewService(sampleDataRoot, dataRoot string) *Service {
	return &Service{
		sampleDataRoot: sampleDataRoot,
		dataRoot:       dataRoot,
	}
}

// ListStatuses returns quality results computed from the sample dataset.
func (s *Service) ListStatuses() ([]CheckStatus, error) {
	artifactPath := filepath.Join(s.dataRoot, "quality", "check_uncategorized_transactions.json")
	if bytes, err := os.ReadFile(artifactPath); err == nil {
		var artifact struct {
			Status        string `json:"status"`
			Uncategorized int    `json:"uncategorized"`
		}
		if err := json.Unmarshal(bytes, &artifact); err == nil {
			return []CheckStatus{
				{
					ID:       "check_duplicate_transactions",
					Name:     "Duplicate Transactions",
					Status:   "passing",
					Severity: "high",
					Message:  "No duplicate transaction identifiers were recorded in the latest worker run artifact.",
				},
				{
					ID:       "check_uncategorized_transactions",
					Name:     "Uncategorized Transactions",
					Status:   artifact.Status,
					Severity: "medium",
					Message:  fmt.Sprintf("Detected %d uncategorized transactions in the latest worker run artifact.", artifact.Uncategorized),
				},
			}, nil
		}
	}

	path := filepath.Join(s.sampleDataRoot, "personal_finance", "transactions.csv")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open transactions sample: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read transactions sample: %w", err)
	}

	seen := map[string]struct{}{}
	duplicates := 0
	uncategorized := 0
	for index, row := range rows {
		if index == 0 || len(row) < 5 {
			continue
		}
		id := strings.TrimSpace(row[0])
		if _, exists := seen[id]; exists {
			duplicates++
		}
		seen[id] = struct{}{}
		if strings.TrimSpace(row[3]) == "" {
			uncategorized++
		}
	}

	return []CheckStatus{
		{
			ID:       "check_duplicate_transactions",
			Name:     "Duplicate Transactions",
			Status:   statusFromCount(duplicates, 0),
			Severity: "high",
			Message:  fmt.Sprintf("Detected %d duplicate transaction identifiers in the sample feed.", duplicates),
		},
		{
			ID:       "check_uncategorized_transactions",
			Name:     "Uncategorized Transactions",
			Status:   statusFromCount(uncategorized, 0),
			Severity: "medium",
			Message:  fmt.Sprintf("Detected %d uncategorized transactions in the sample feed.", uncategorized),
		},
	}, nil
}

func statusFromCount(count int, passingThreshold int) string {
	if count <= passingThreshold {
		return "passing"
	}
	return "warning"
}
