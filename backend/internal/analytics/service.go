// Package analytics provides a deliberately constrained analytics-serving
// surface. The service prefers DuckDB-backed curated datasets and falls back to
// materialized artifacts or sample data only when the analytical database is
// not yet ready.
package analytics

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/streanor/data-platform/backend/internal/transforms"
)

// Service owns curated analytics responses.
type Service struct {
	sampleDataRoot string
	dataRoot       string
	sql            *transforms.Engine
}

// QueryResult is shaped for chart-friendly frontend consumption.
type QueryResult struct {
	Dataset string           `json:"dataset"`
	Series  []map[string]any `json:"series"`
}

// NewService creates an analytics service.
func NewService(sampleDataRoot, dataRoot, duckDBPath, sqlRoot string) *Service {
	return &Service{
		sampleDataRoot: sampleDataRoot,
		dataRoot:       dataRoot,
		sql:            transforms.NewEngine(duckDBPath, sqlRoot),
	}
}

// QueryDataset returns one curated dataset by identifier.
func (s *Service) QueryDataset(dataset string) (QueryResult, error) {
	switch dataset {
	case "", "mart_monthly_cashflow":
		return s.SampleDashboard()
	case "metrics_savings_rate":
		return s.QueryMetric(dataset)
	default:
		return QueryResult{}, fmt.Errorf("unknown curated dataset %q", dataset)
	}
}

// QueryMetric returns one curated metric dataset by identifier.
func (s *Service) QueryMetric(metricID string) (QueryResult, error) {
	if metricID != "metrics_savings_rate" {
		return QueryResult{}, fmt.Errorf("unknown metric %q", metricID)
	}

	if rows, err := s.sql.QueryRows(`
		select month, savings_rate
		from metrics_savings_rate
		order by month
	`); err == nil && len(rows) > 0 {
		return QueryResult{
			Dataset: metricID,
			Series:  rows,
		}, nil
	}

	artifactPath := filepath.Join(s.dataRoot, "metrics", "metrics_savings_rate.json")
	if bytes, err := os.ReadFile(artifactPath); err == nil {
		var artifact struct {
			Series []map[string]any `json:"series"`
		}
		if err := json.Unmarshal(bytes, &artifact); err == nil && len(artifact.Series) > 0 {
			return QueryResult{
				Dataset: metricID,
				Series:  artifact.Series,
			}, nil
		}
	}

	return QueryResult{}, fmt.Errorf("metric %q is not available yet", metricID)
}

// SampleDashboard returns the curated cashflow dashboard dataset.
func (s *Service) SampleDashboard() (QueryResult, error) {
	if rows, err := s.sql.QueryRows(`
		select month, income, expenses, savings_rate
		from mart_monthly_cashflow
		order by month
	`); err == nil && len(rows) > 0 {
		return QueryResult{
			Dataset: "metrics_monthly_cashflow",
			Series:  rows,
		}, nil
	}

	artifactPath := filepath.Join(s.dataRoot, "mart", "mart_monthly_cashflow.json")
	if bytes, err := os.ReadFile(artifactPath); err == nil {
		var rows []map[string]any
		if err := json.Unmarshal(bytes, &rows); err == nil && len(rows) > 0 {
			return QueryResult{
				Dataset: "metrics_monthly_cashflow",
				Series:  rows,
			}, nil
		}
	}

	path := filepath.Join(s.sampleDataRoot, "personal_finance", "transactions.csv")
	file, err := os.Open(path)
	if err != nil {
		return QueryResult{}, fmt.Errorf("open transactions sample: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return QueryResult{}, fmt.Errorf("read transactions sample: %w", err)
	}

	type summary struct {
		income   float64
		expenses float64
	}

	summaries := map[string]summary{}
	for index, row := range rows {
		if index == 0 || len(row) < 5 {
			continue
		}
		month := ""
		if len(row[1]) >= 7 {
			month = row[1][:7]
		}
		amount, err := strconv.ParseFloat(strings.TrimSpace(row[4]), 64)
		if err != nil {
			return QueryResult{}, fmt.Errorf("parse amount on row %d: %w", index+1, err)
		}
		current := summaries[month]
		if amount >= 0 {
			current.income += amount
		} else {
			current.expenses += amount * -1
		}
		summaries[month] = current
	}

	months := make([]string, 0, len(summaries))
	for month := range summaries {
		months = append(months, month)
	}
	sort.Strings(months)

	series := make([]map[string]any, 0, len(months))
	for _, month := range months {
		item := summaries[month]
		savingsRate := 0.0
		if item.income > 0 {
			savingsRate = (item.income - item.expenses) / item.income
		}
		series = append(series, map[string]any{
			"month":        month,
			"income":       item.income,
			"expenses":     item.expenses,
			"savings_rate": savingsRate,
		})
	}

	return QueryResult{
		Dataset: "metrics_monthly_cashflow",
		Series:  series,
	}, nil
}
