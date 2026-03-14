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

// QueryOptions captures the constrained filter contract accepted by the
// analytics API. The service intentionally keeps this small so the reporting
// layer cannot drift into arbitrary-query behavior.
type QueryOptions struct {
	FromMonth string
	ToMonth   string
	Category  string
	Limit     int
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
func (s *Service) QueryDataset(dataset string, options QueryOptions) (QueryResult, error) {
	switch dataset {
	case "", "mart_monthly_cashflow":
		return s.queryMonthlyCashflow(options)
	case "mart_category_spend":
		return s.queryCategorySpend(options)
	case "mart_budget_vs_actual":
		return s.queryBudgetVariance(options)
	default:
		return QueryResult{}, fmt.Errorf("unknown curated dataset %q", dataset)
	}
}

// QueryMetric returns one curated metric dataset by identifier.
func (s *Service) QueryMetric(metricID string, options QueryOptions) (QueryResult, error) {
	switch metricID {
	case "metrics_savings_rate":
		return s.querySavingsRateMetric(options)
	case "metrics_category_variance":
		return s.queryCategoryVarianceMetric(options)
	default:
		return QueryResult{}, fmt.Errorf("unknown metric %q", metricID)
	}
}

// SampleDashboard returns the default monthly cashflow dataset for the
// dashboard landing view.
func (s *Service) SampleDashboard() (QueryResult, error) {
	return s.queryMonthlyCashflow(QueryOptions{})
}

func (s *Service) queryMonthlyCashflow(options QueryOptions) (QueryResult, error) {
	query := `
		select month, income, expenses, savings_rate
		from mart_monthly_cashflow
	`
	if rows, err := s.queryDuckDB(query, options, false, "month"); err == nil && len(rows) > 0 {
		return QueryResult{Dataset: "mart_monthly_cashflow", Series: rows}, nil
	}
	if rows, err := s.loadArtifactRows(filepath.Join("mart", "mart_monthly_cashflow.json"), options); err == nil && len(rows) > 0 {
		return QueryResult{Dataset: "mart_monthly_cashflow", Series: rows}, nil
	}

	rows, err := s.sampleMonthlyCashflow(options)
	if err != nil {
		return QueryResult{}, err
	}
	return QueryResult{Dataset: "mart_monthly_cashflow", Series: rows}, nil
}

func (s *Service) queryCategorySpend(options QueryOptions) (QueryResult, error) {
	query := `
		select month, category, actual_spend
		from mart_category_spend
	`
	if rows, err := s.queryDuckDB(query, options, true, "month, category"); err == nil && len(rows) > 0 {
		return QueryResult{Dataset: "mart_category_spend", Series: rows}, nil
	}
	if rows, err := s.loadArtifactRows(filepath.Join("mart", "mart_category_spend.json"), options); err == nil && len(rows) > 0 {
		return QueryResult{Dataset: "mart_category_spend", Series: rows}, nil
	}

	rows, err := s.sampleCategorySpend(options)
	if err != nil {
		return QueryResult{}, err
	}
	return QueryResult{Dataset: "mart_category_spend", Series: rows}, nil
}

func (s *Service) queryBudgetVariance(options QueryOptions) (QueryResult, error) {
	query := `
		select month, category, budget_amount, actual_spend, variance_amount
		from mart_budget_vs_actual
	`
	if rows, err := s.queryDuckDB(query, options, true, "month, category"); err == nil && len(rows) > 0 {
		return QueryResult{Dataset: "mart_budget_vs_actual", Series: rows}, nil
	}
	if rows, err := s.loadArtifactRows(filepath.Join("mart", "mart_budget_vs_actual.json"), options); err == nil && len(rows) > 0 {
		return QueryResult{Dataset: "mart_budget_vs_actual", Series: rows}, nil
	}

	rows, err := s.sampleBudgetVariance(options)
	if err != nil {
		return QueryResult{}, err
	}
	return QueryResult{Dataset: "mart_budget_vs_actual", Series: rows}, nil
}

func (s *Service) querySavingsRateMetric(options QueryOptions) (QueryResult, error) {
	query := `
		select month, savings_rate
		from metrics_savings_rate
	`
	if rows, err := s.queryDuckDB(query, options, false, "month"); err == nil && len(rows) > 0 {
		return QueryResult{Dataset: "metrics_savings_rate", Series: rows}, nil
	}
	if rows, err := s.loadMetricArtifact("metrics_savings_rate", options); err == nil && len(rows) > 0 {
		return QueryResult{Dataset: "metrics_savings_rate", Series: rows}, nil
	}

	rows, err := s.sampleSavingsRate(options)
	if err != nil {
		return QueryResult{}, err
	}
	return QueryResult{Dataset: "metrics_savings_rate", Series: rows}, nil
}

func (s *Service) queryCategoryVarianceMetric(options QueryOptions) (QueryResult, error) {
	query := `
		select month, category, variance_amount
		from metrics_category_variance
	`
	if rows, err := s.queryDuckDB(query, options, true, "month, category"); err == nil && len(rows) > 0 {
		return QueryResult{Dataset: "metrics_category_variance", Series: rows}, nil
	}
	if rows, err := s.loadMetricArtifact("metrics_category_variance", options); err == nil && len(rows) > 0 {
		return QueryResult{Dataset: "metrics_category_variance", Series: rows}, nil
	}

	rows, err := s.sampleCategoryVariance(options)
	if err != nil {
		return QueryResult{}, err
	}
	return QueryResult{Dataset: "metrics_category_variance", Series: rows}, nil
}

func (s *Service) queryDuckDB(baseQuery string, options QueryOptions, includeCategory bool, orderBy string) ([]map[string]any, error) {
	clauses := []string{}
	args := []any{}

	if options.FromMonth != "" {
		clauses = append(clauses, "month >= ?")
		args = append(args, options.FromMonth)
	}
	if options.ToMonth != "" {
		clauses = append(clauses, "month <= ?")
		args = append(args, options.ToMonth)
	}
	if includeCategory && options.Category != "" {
		clauses = append(clauses, "category = ?")
		args = append(args, options.Category)
	}

	query := baseQuery
	if len(clauses) > 0 {
		query += " where " + strings.Join(clauses, " and ")
	}
	query += " order by " + orderBy
	if options.Limit > 0 {
		query += " limit ?"
		args = append(args, options.Limit)
	}
	return s.sql.QueryRowsArgs(query, args...)
}

func (s *Service) loadArtifactRows(relativePath string, options QueryOptions) ([]map[string]any, error) {
	bytes, err := os.ReadFile(filepath.Join(s.dataRoot, relativePath))
	if err != nil {
		return nil, err
	}
	var rows []map[string]any
	if err := json.Unmarshal(bytes, &rows); err != nil {
		return nil, err
	}
	return filterRows(rows, options), nil
}

func (s *Service) loadMetricArtifact(metricID string, options QueryOptions) ([]map[string]any, error) {
	bytes, err := os.ReadFile(filepath.Join(s.dataRoot, "metrics", metricID+".json"))
	if err != nil {
		return nil, err
	}
	var payload struct {
		Series []map[string]any `json:"series"`
	}
	if err := json.Unmarshal(bytes, &payload); err != nil {
		return nil, err
	}
	return filterRows(payload.Series, options), nil
}

func (s *Service) sampleMonthlyCashflow(options QueryOptions) ([]map[string]any, error) {
	transactions, err := s.sampleTransactions()
	if err != nil {
		return nil, err
	}

	type summary struct {
		income   float64
		expenses float64
	}
	summaries := map[string]summary{}
	for _, transaction := range transactions {
		current := summaries[transaction.Month]
		if transaction.Amount >= 0 {
			current.income += transaction.Amount
		} else {
			current.expenses += -transaction.Amount
		}
		summaries[transaction.Month] = current
	}

	months := sortedKeys(summaries)
	rows := make([]map[string]any, 0, len(months))
	for _, month := range months {
		item := summaries[month]
		savingsRate := 0.0
		if item.income > 0 {
			savingsRate = (item.income - item.expenses) / item.income
		}
		rows = append(rows, map[string]any{
			"month":        month,
			"income":       item.income,
			"expenses":     item.expenses,
			"savings_rate": savingsRate,
		})
	}
	return filterRows(rows, options), nil
}

func (s *Service) sampleCategorySpend(options QueryOptions) ([]map[string]any, error) {
	transactions, err := s.sampleTransactions()
	if err != nil {
		return nil, err
	}
	spend := map[string]float64{}
	for _, transaction := range transactions {
		if transaction.Amount >= 0 {
			continue
		}
		key := transaction.Month + "|" + transaction.Category
		spend[key] += -transaction.Amount
	}

	keys := sortedKeys(spend)
	rows := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		parts := strings.SplitN(key, "|", 2)
		rows = append(rows, map[string]any{
			"month":        parts[0],
			"category":     parts[1],
			"actual_spend": spend[key],
		})
	}
	return filterRows(rows, options), nil
}

func (s *Service) sampleBudgetVariance(options QueryOptions) ([]map[string]any, error) {
	categorySpend, err := s.sampleCategorySpend(QueryOptions{})
	if err != nil {
		return nil, err
	}
	budgets, err := s.sampleBudgets()
	if err != nil {
		return nil, err
	}

	months := map[string]struct{}{}
	actualByKey := map[string]float64{}
	for _, row := range categorySpend {
		month := stringValue(row["month"])
		category := stringValue(row["category"])
		months[month] = struct{}{}
		actualByKey[month+"|"+category] = numericValue(row["actual_spend"])
	}

	monthKeys := sortedKeys(months)
	rows := make([]map[string]any, 0, len(monthKeys)*len(budgets))
	for _, month := range monthKeys {
		for category, budgetAmount := range budgets {
			actualSpend := actualByKey[month+"|"+category]
			rows = append(rows, map[string]any{
				"month":           month,
				"category":        category,
				"budget_amount":   budgetAmount,
				"actual_spend":    actualSpend,
				"variance_amount": actualSpend - budgetAmount,
			})
		}
	}
	return filterRows(rows, options), nil
}

func (s *Service) sampleSavingsRate(options QueryOptions) ([]map[string]any, error) {
	cashflow, err := s.sampleMonthlyCashflow(options)
	if err != nil {
		return nil, err
	}
	rows := make([]map[string]any, 0, len(cashflow))
	for _, row := range cashflow {
		rows = append(rows, map[string]any{
			"month":        row["month"],
			"savings_rate": row["savings_rate"],
		})
	}
	return rows, nil
}

func (s *Service) sampleCategoryVariance(options QueryOptions) ([]map[string]any, error) {
	varianceRows, err := s.sampleBudgetVariance(options)
	if err != nil {
		return nil, err
	}
	rows := make([]map[string]any, 0, len(varianceRows))
	for _, row := range varianceRows {
		rows = append(rows, map[string]any{
			"month":           row["month"],
			"category":        row["category"],
			"variance_amount": row["variance_amount"],
		})
	}
	return rows, nil
}

type sampleTransaction struct {
	Month    string
	Category string
	Amount   float64
}

func (s *Service) sampleTransactions() ([]sampleTransaction, error) {
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

	transactions := make([]sampleTransaction, 0, len(rows))
	for index, row := range rows {
		if index == 0 || len(row) < 5 {
			continue
		}
		amount, err := strconv.ParseFloat(strings.TrimSpace(row[4]), 64)
		if err != nil {
			return nil, fmt.Errorf("parse amount on row %d: %w", index+1, err)
		}
		category := strings.TrimSpace(row[3])
		if category == "" {
			category = "Uncategorized"
		}
		transactions = append(transactions, sampleTransaction{
			Month:    row[1][:7],
			Category: category,
			Amount:   amount,
		})
	}
	return transactions, nil
}

func (s *Service) sampleBudgets() (map[string]float64, error) {
	path := filepath.Join(s.sampleDataRoot, "personal_finance", "budget_rules.json")
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read budget rules sample: %w", err)
	}
	var records []struct {
		Category      string  `json:"category"`
		MonthlyBudget float64 `json:"monthly_budget"`
	}
	if err := json.Unmarshal(bytes, &records); err != nil {
		return nil, fmt.Errorf("decode budget rules sample: %w", err)
	}
	budgets := make(map[string]float64, len(records))
	for _, record := range records {
		budgets[record.Category] = record.MonthlyBudget
	}
	return budgets, nil
}

func filterRows(rows []map[string]any, options QueryOptions) []map[string]any {
	filtered := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		month := stringValue(row["month"])
		category := stringValue(row["category"])
		if options.FromMonth != "" && month != "" && month < options.FromMonth {
			continue
		}
		if options.ToMonth != "" && month != "" && month > options.ToMonth {
			continue
		}
		if options.Category != "" && category != "" && category != options.Category {
			continue
		}
		filtered = append(filtered, row)
	}
	if options.Limit > 0 && len(filtered) > options.Limit {
		return filtered[:options.Limit]
	}
	return filtered
}

func sortedKeys[T any](items map[string]T) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprintf("%v", typed)
	}
}

func numericValue(value any) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err == nil {
			return parsed
		}
	}
	return 0
}
