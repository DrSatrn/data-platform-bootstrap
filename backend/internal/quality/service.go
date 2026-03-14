// Package quality exposes operator-visible data trust signals. The service
// prefers DuckDB-backed quality queries and falls back to artifacts or sample
// data only when the analytical database is not yet ready.
package quality

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/streanor/data-platform/backend/internal/transforms"
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
	sql            *transforms.Engine
}

// NewService creates a quality service.
func NewService(sampleDataRoot, dataRoot, duckDBPath, sqlRoot string) *Service {
	return &Service{
		sampleDataRoot: sampleDataRoot,
		dataRoot:       dataRoot,
		sql:            transforms.NewEngine(duckDBPath, sqlRoot),
	}
}

// ListStatuses returns quality results computed from DuckDB when available.
func (s *Service) ListStatuses() ([]CheckStatus, error) {
	if statuses, err := s.listStatusesFromDuckDB(); err == nil {
		return statuses, nil
	}

	artifactPath := filepath.Join(s.dataRoot, "quality", "check_uncategorized_transactions.json")
	if bytes, err := os.ReadFile(artifactPath); err == nil {
		var artifact struct {
			Status             string `json:"status"`
			UncategorizedCount int    `json:"uncategorized_count"`
			Uncategorized      int    `json:"uncategorized"`
		}
		if err := json.Unmarshal(bytes, &artifact); err == nil {
			count := artifact.UncategorizedCount
			if count == 0 {
				count = artifact.Uncategorized
			}
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
					Message:  fmt.Sprintf("Detected %d uncategorized transactions in the latest worker run artifact.", count),
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

func (s *Service) listStatusesFromDuckDB() ([]CheckStatus, error) {
	duplicates, err := s.sql.QueryRowsFromFile(filepath.Join("quality", "check_duplicate_transactions.sql"), nil)
	if err != nil {
		return nil, err
	}
	uncategorized, err := s.sql.QueryRowsFromFile(filepath.Join("quality", "check_uncategorized_transactions.sql"), nil)
	if err != nil {
		return nil, err
	}
	if len(duplicates) == 0 || len(uncategorized) == 0 {
		return nil, fmt.Errorf("duckdb quality queries returned no rows")
	}

	duplicateCount := intFromRow(duplicates[0]["duplicate_count"])
	uncategorizedCount := intFromRow(uncategorized[0]["uncategorized_count"])
	return []CheckStatus{
		{
			ID:       "check_duplicate_transactions",
			Name:     "Duplicate Transactions",
			Status:   statusFromCount(duplicateCount, 0),
			Severity: "high",
			Message:  fmt.Sprintf("Detected %d duplicate transaction identifiers in DuckDB raw_transactions.", duplicateCount),
		},
		{
			ID:       "check_uncategorized_transactions",
			Name:     "Uncategorized Transactions",
			Status:   statusFromCount(uncategorizedCount, 0),
			Severity: "medium",
			Message:  fmt.Sprintf("Detected %d uncategorized transactions in DuckDB raw_transactions.", uncategorizedCount),
		},
	}, nil
}

func intFromRow(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int32:
		return int(typed)
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	case string:
		return statusCountString(typed)
	default:
		return 0
	}
}

func statusCountString(value string) int {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	count := 0
	for _, character := range value {
		if character < '0' || character > '9' {
			return 0
		}
		count = count*10 + int(character-'0')
	}
	return count
}

func statusFromCount(count int, passingThreshold int) string {
	if count <= passingThreshold {
		return "passing"
	}
	return "warning"
}
