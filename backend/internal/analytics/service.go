// Package analytics provides a deliberately constrained analytics-serving
// surface. The service prefers DuckDB-backed curated datasets and falls back to
// materialized artifacts or sample data only when the analytical database is
// not yet ready.
package analytics

import (
	"bytes"
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
	FromMonth      string
	ToMonth        string
	Category       string
	Limit          int
	GroupBy        string
	DrillDimension string
	DrillValue     string
	SortBy         string
	SortDirection  string
}

// QueryResult is shaped for chart-friendly frontend consumption.
type QueryResult struct {
	Dataset             string           `json:"dataset"`
	Series              []map[string]any `json:"series"`
	AvailableDimensions []string         `json:"available_dimensions,omitempty"`
	AvailableMeasures   []string         `json:"available_measures,omitempty"`
	GroupBy             string           `json:"group_by,omitempty"`
	DrillDimension      string           `json:"drill_dimension,omitempty"`
	DrillValue          string           `json:"drill_value,omitempty"`
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
	case "mart_inventory_monthly_summary":
		return s.queryInventoryMonthlySummary(options)
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
	case "metrics_inventory_net_change":
		return s.queryInventoryNetChangeMetric(options)
	default:
		return QueryResult{}, fmt.Errorf("unknown metric %q", metricID)
	}
}

// SampleDashboard returns the default monthly cashflow dataset for the
// dashboard landing view.
func (s *Service) SampleDashboard() (QueryResult, error) {
	return s.queryMonthlyCashflow(QueryOptions{})
}

// ExportDatasetCSV returns one curated dataset serialized as CSV using the
// same constrained query path as the JSON analytics API.
func (s *Service) ExportDatasetCSV(dataset string, options QueryOptions) ([]byte, error) {
	result, err := s.QueryDataset(dataset, options)
	if err != nil {
		return nil, err
	}
	return encodeQueryResultCSV(result)
}

// ExportMetricCSV returns one curated metric query serialized as CSV.
func (s *Service) ExportMetricCSV(metricID string, options QueryOptions) ([]byte, error) {
	result, err := s.QueryMetric(metricID, options)
	if err != nil {
		return nil, err
	}
	return encodeQueryResultCSV(result)
}

func (s *Service) queryMonthlyCashflow(options QueryOptions) (QueryResult, error) {
	query := `
		select month, income, expenses, savings_rate
		from mart_monthly_cashflow
	`
	if rows, err := s.queryDuckDB(query, options, false, "month"); err == nil && len(rows) > 0 {
		return finalizeQueryResult("mart_monthly_cashflow", rows, options)
	}
	if rows, err := s.loadArtifactRows(filepath.Join("mart", "mart_monthly_cashflow.json"), options); err == nil && len(rows) > 0 {
		return finalizeQueryResult("mart_monthly_cashflow", rows, options)
	}

	rows, err := s.sampleMonthlyCashflow(options)
	if err != nil {
		return QueryResult{}, err
	}
	return finalizeQueryResult("mart_monthly_cashflow", rows, options)
}

func (s *Service) queryCategorySpend(options QueryOptions) (QueryResult, error) {
	query := `
		select month, category, actual_spend
		from mart_category_spend
	`
	if rows, err := s.queryDuckDB(query, options, true, "month, category"); err == nil && len(rows) > 0 {
		return finalizeQueryResult("mart_category_spend", rows, options)
	}
	if rows, err := s.loadArtifactRows(filepath.Join("mart", "mart_category_spend.json"), options); err == nil && len(rows) > 0 {
		return finalizeQueryResult("mart_category_spend", rows, options)
	}

	rows, err := s.sampleCategorySpend(options)
	if err != nil {
		return QueryResult{}, err
	}
	return finalizeQueryResult("mart_category_spend", rows, options)
}

func (s *Service) queryBudgetVariance(options QueryOptions) (QueryResult, error) {
	query := `
		select month, category, budget_amount, actual_spend, variance_amount
		from mart_budget_vs_actual
	`
	if rows, err := s.queryDuckDB(query, options, true, "month, category"); err == nil && len(rows) > 0 {
		return finalizeQueryResult("mart_budget_vs_actual", rows, options)
	}
	if rows, err := s.loadArtifactRows(filepath.Join("mart", "mart_budget_vs_actual.json"), options); err == nil && len(rows) > 0 {
		return finalizeQueryResult("mart_budget_vs_actual", rows, options)
	}

	rows, err := s.sampleBudgetVariance(options)
	if err != nil {
		return QueryResult{}, err
	}
	return finalizeQueryResult("mart_budget_vs_actual", rows, options)
}

func (s *Service) querySavingsRateMetric(options QueryOptions) (QueryResult, error) {
	query := `
		select month, savings_rate
		from metrics_savings_rate
	`
	if rows, err := s.queryDuckDB(query, options, false, "month"); err == nil && len(rows) > 0 {
		return finalizeQueryResult("metrics_savings_rate", rows, options)
	}
	if rows, err := s.loadMetricArtifact("metrics_savings_rate", options); err == nil && len(rows) > 0 {
		return finalizeQueryResult("metrics_savings_rate", rows, options)
	}

	rows, err := s.sampleSavingsRate(options)
	if err != nil {
		return QueryResult{}, err
	}
	return finalizeQueryResult("metrics_savings_rate", rows, options)
}

func (s *Service) queryCategoryVarianceMetric(options QueryOptions) (QueryResult, error) {
	query := `
		select month, category, variance_amount
		from metrics_category_variance
	`
	if rows, err := s.queryDuckDB(query, options, true, "month, category"); err == nil && len(rows) > 0 {
		return finalizeQueryResult("metrics_category_variance", rows, options)
	}
	if rows, err := s.loadMetricArtifact("metrics_category_variance", options); err == nil && len(rows) > 0 {
		return finalizeQueryResult("metrics_category_variance", rows, options)
	}

	rows, err := s.sampleCategoryVariance(options)
	if err != nil {
		return QueryResult{}, err
	}
	return finalizeQueryResult("metrics_category_variance", rows, options)
}

func (s *Service) queryInventoryMonthlySummary(options QueryOptions) (QueryResult, error) {
	query := `
		select month, sku, warehouse, receipts, issues, net_quantity, closing_quantity, movement_count
		from mart_inventory_monthly_summary
	`
	if rows, err := s.queryDuckDB(query, options, false, "month, sku, warehouse"); err == nil && len(rows) > 0 {
		return finalizeQueryResult("mart_inventory_monthly_summary", rows, options)
	}
	if rows, err := s.loadArtifactRows(filepath.Join("mart", "mart_inventory_monthly_summary.json"), options); err == nil && len(rows) > 0 {
		return finalizeQueryResult("mart_inventory_monthly_summary", rows, options)
	}

	rows, err := s.sampleInventoryMonthlySummary(options)
	if err != nil {
		return QueryResult{}, err
	}
	return finalizeQueryResult("mart_inventory_monthly_summary", rows, options)
}

func (s *Service) queryInventoryNetChangeMetric(options QueryOptions) (QueryResult, error) {
	query := `
		select month, warehouse, net_quantity
		from metrics_inventory_net_change
	`
	if rows, err := s.queryDuckDB(query, options, false, "month, warehouse"); err == nil && len(rows) > 0 {
		return finalizeQueryResult("metrics_inventory_net_change", rows, options)
	}
	if rows, err := s.loadMetricArtifact("metrics_inventory_net_change", options); err == nil && len(rows) > 0 {
		return finalizeQueryResult("metrics_inventory_net_change", rows, options)
	}

	rows, err := s.sampleInventoryNetChange(options)
	if err != nil {
		return QueryResult{}, err
	}
	return finalizeQueryResult("metrics_inventory_net_change", rows, options)
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

type sampleInventoryMovement struct {
	Month     string
	SKU       string
	Warehouse string
	Quantity  float64
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

func (s *Service) sampleInventoryMovements() ([]sampleInventoryMovement, error) {
	path := filepath.Join(s.sampleDataRoot, "inventory_operations", "stock_movements.csv")
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open inventory movements sample: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("read inventory movements sample: %w", err)
	}

	movements := make([]sampleInventoryMovement, 0, len(rows))
	for index, row := range rows {
		if index == 0 || len(row) < 5 {
			continue
		}
		quantity, err := strconv.ParseFloat(strings.TrimSpace(row[4]), 64)
		if err != nil {
			return nil, fmt.Errorf("parse inventory quantity on row %d: %w", index+1, err)
		}
		movements = append(movements, sampleInventoryMovement{
			Month:     strings.TrimSpace(row[1])[:7],
			SKU:       strings.TrimSpace(row[0]),
			Warehouse: strings.TrimSpace(row[3]),
			Quantity:  quantity,
		})
	}
	return movements, nil
}

func (s *Service) sampleInventoryMonthlySummary(options QueryOptions) ([]map[string]any, error) {
	movements, err := s.sampleInventoryMovements()
	if err != nil {
		return nil, err
	}
	type summary struct {
		receipts      float64
		issues        float64
		netQuantity   float64
		movementCount int
	}
	summaries := map[string]summary{}
	for _, movement := range movements {
		key := strings.Join([]string{movement.Month, movement.SKU, movement.Warehouse}, "|")
		current := summaries[key]
		if movement.Quantity >= 0 {
			current.receipts += movement.Quantity
		} else {
			current.issues += -movement.Quantity
		}
		current.netQuantity += movement.Quantity
		current.movementCount++
		summaries[key] = current
	}

	keys := sortedKeys(summaries)
	closingBySKU := map[string]float64{}
	rows := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		parts := strings.Split(key, "|")
		summary := summaries[key]
		skuWarehouse := parts[1] + "|" + parts[2]
		closingBySKU[skuWarehouse] += summary.netQuantity
		rows = append(rows, map[string]any{
			"month":            parts[0],
			"sku":              parts[1],
			"warehouse":        parts[2],
			"receipts":         summary.receipts,
			"issues":           summary.issues,
			"net_quantity":     summary.netQuantity,
			"closing_quantity": closingBySKU[skuWarehouse],
			"movement_count":   summary.movementCount,
		})
	}
	return filterRows(rows, options), nil
}

func (s *Service) sampleInventoryNetChange(options QueryOptions) ([]map[string]any, error) {
	rows, err := s.sampleInventoryMonthlySummary(options)
	if err != nil {
		return nil, err
	}
	type bucket struct {
		netQuantity float64
	}
	buckets := map[string]bucket{}
	for _, row := range rows {
		key := stringValue(row["month"]) + "|" + stringValue(row["warehouse"])
		current := buckets[key]
		current.netQuantity += numericValue(row["net_quantity"])
		buckets[key] = current
	}
	keys := sortedKeys(buckets)
	series := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		parts := strings.Split(key, "|")
		series = append(series, map[string]any{
			"month":        parts[0],
			"warehouse":    parts[1],
			"net_quantity": buckets[key].netQuantity,
		})
	}
	return series, nil
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
		if options.DrillDimension != "" && options.DrillValue != "" {
			if stringValue(row[options.DrillDimension]) != options.DrillValue {
				continue
			}
		}
		filtered = append(filtered, row)
	}
	return filtered
}

type datasetSchema struct {
	dimensions []string
	measures   []string
}

func finalizeQueryResult(dataset string, rows []map[string]any, options QueryOptions) (QueryResult, error) {
	schema, err := schemaForDataset(dataset)
	if err != nil {
		return QueryResult{}, err
	}
	filtered := filterRows(rows, options)
	grouped, err := applyGrouping(filtered, schema, options)
	if err != nil {
		return QueryResult{}, err
	}
	sorted := sortRows(grouped, options)
	if options.Limit > 0 && len(sorted) > options.Limit {
		sorted = sorted[:options.Limit]
	}
	return QueryResult{
		Dataset:             dataset,
		Series:              sorted,
		AvailableDimensions: schema.dimensions,
		AvailableMeasures:   schema.measures,
		GroupBy:             options.GroupBy,
		DrillDimension:      options.DrillDimension,
		DrillValue:          options.DrillValue,
	}, nil
}

func schemaForDataset(dataset string) (datasetSchema, error) {
	switch dataset {
	case "mart_monthly_cashflow":
		return datasetSchema{
			dimensions: []string{"month"},
			measures:   []string{"income", "expenses", "savings_rate"},
		}, nil
	case "mart_category_spend":
		return datasetSchema{
			dimensions: []string{"month", "category"},
			measures:   []string{"actual_spend"},
		}, nil
	case "mart_budget_vs_actual":
		return datasetSchema{
			dimensions: []string{"month", "category"},
			measures:   []string{"budget_amount", "actual_spend", "variance_amount"},
		}, nil
	case "mart_inventory_monthly_summary":
		return datasetSchema{
			dimensions: []string{"month", "sku", "warehouse"},
			measures:   []string{"receipts", "issues", "net_quantity", "closing_quantity", "movement_count"},
		}, nil
	case "metrics_savings_rate":
		return datasetSchema{
			dimensions: []string{"month"},
			measures:   []string{"savings_rate"},
		}, nil
	case "metrics_category_variance":
		return datasetSchema{
			dimensions: []string{"month", "category"},
			measures:   []string{"variance_amount"},
		}, nil
	case "metrics_inventory_net_change":
		return datasetSchema{
			dimensions: []string{"month", "warehouse"},
			measures:   []string{"net_quantity"},
		}, nil
	default:
		return datasetSchema{}, fmt.Errorf("unknown schema for dataset %q", dataset)
	}
}

func applyGrouping(rows []map[string]any, schema datasetSchema, options QueryOptions) ([]map[string]any, error) {
	groupByFields := parseGroupByFields(options.GroupBy)
	if len(groupByFields) == 0 {
		return rows, nil
	}
	for _, field := range groupByFields {
		if !containsString(schema.dimensions, field) {
			return nil, fmt.Errorf("group_by %q is not supported for this dataset", field)
		}
	}

	grouped := map[string]map[string]any{}
	for _, row := range rows {
		keyParts := make([]string, 0, len(groupByFields))
		for _, field := range groupByFields {
			value := stringValue(row[field])
			if value == "" {
				value = "(empty)"
			}
			keyParts = append(keyParts, value)
		}
		key := strings.Join(keyParts, "\x1f")
		current, ok := grouped[key]
		if !ok {
			current = map[string]any{}
			for index, field := range groupByFields {
				current[field] = keyParts[index]
			}
			for _, measure := range schema.measures {
				current[measure] = 0.0
			}
		}
		for _, measure := range schema.measures {
			current[measure] = numericValue(current[measure]) + numericValue(row[measure])
		}
		grouped[key] = current
	}

	keys := sortedKeys(grouped)
	out := make([]map[string]any, 0, len(keys))
	for _, key := range keys {
		out = append(out, grouped[key])
	}
	return out, nil
}

func parseGroupByFields(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	fields := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, part := range parts {
		field := strings.TrimSpace(part)
		if field == "" {
			continue
		}
		if _, exists := seen[field]; exists {
			continue
		}
		seen[field] = struct{}{}
		fields = append(fields, field)
	}
	return fields
}

func encodeQueryResultCSV(result QueryResult) ([]byte, error) {
	var buffer bytes.Buffer
	writer := csv.NewWriter(&buffer)

	columns := orderedCSVColumns(result)
	if len(columns) == 0 {
		return []byte{}, nil
	}
	if err := writer.Write(columns); err != nil {
		return nil, fmt.Errorf("write csv header: %w", err)
	}
	for _, row := range result.Series {
		record := make([]string, 0, len(columns))
		for _, column := range columns {
			record = append(record, stringValue(row[column]))
		}
		if err := writer.Write(record); err != nil {
			return nil, fmt.Errorf("write csv row: %w", err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, fmt.Errorf("flush csv writer: %w", err)
	}
	return buffer.Bytes(), nil
}

func orderedCSVColumns(result QueryResult) []string {
	present := map[string]struct{}{}
	for _, row := range result.Series {
		for key := range row {
			present[key] = struct{}{}
		}
	}
	if len(present) == 0 {
		groupByFields := parseGroupByFields(result.GroupBy)
		if len(groupByFields) > 0 {
			return append(groupByFields, result.AvailableMeasures...)
		}
		return append(append([]string{}, result.AvailableDimensions...), result.AvailableMeasures...)
	}

	columns := []string{}
	appendIfPresent := func(items []string) {
		for _, item := range items {
			if _, ok := present[item]; !ok {
				continue
			}
			if containsString(columns, item) {
				continue
			}
			columns = append(columns, item)
		}
	}

	groupByFields := parseGroupByFields(result.GroupBy)
	appendIfPresent(groupByFields)
	appendIfPresent(result.AvailableDimensions)
	appendIfPresent(result.AvailableMeasures)

	extra := make([]string, 0, len(present))
	for key := range present {
		if containsString(columns, key) {
			continue
		}
		extra = append(extra, key)
	}
	sort.Strings(extra)
	columns = append(columns, extra...)
	return columns
}

func sortRows(rows []map[string]any, options QueryOptions) []map[string]any {
	out := make([]map[string]any, len(rows))
	copy(out, rows)
	field := options.SortBy
	if field == "" {
		if options.GroupBy != "" {
			field = options.GroupBy
		} else {
			field = "month"
		}
	}
	descending := strings.EqualFold(options.SortDirection, "desc")
	sort.SliceStable(out, func(i, j int) bool {
		left := out[i][field]
		right := out[j][field]
		leftString := stringValue(left)
		rightString := stringValue(right)
		leftNumber, leftIsNumber := parseNumeric(left)
		rightNumber, rightIsNumber := parseNumeric(right)

		comparison := 0
		switch {
		case leftIsNumber && rightIsNumber:
			switch {
			case leftNumber < rightNumber:
				comparison = -1
			case leftNumber > rightNumber:
				comparison = 1
			}
		default:
			switch {
			case leftString < rightString:
				comparison = -1
			case leftString > rightString:
				comparison = 1
			}
		}
		if descending {
			return comparison > 0
		}
		return comparison < 0
	})
	return out
}

func containsString(values []string, candidate string) bool {
	for _, value := range values {
		if value == candidate {
			return true
		}
	}
	return false
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
	parsed, _ := parseNumeric(value)
	return parsed
}

func parseNumeric(value any) (float64, bool) {
	switch typed := value.(type) {
	case float64:
		return typed, true
	case float32:
		return float64(typed), true
	case int:
		return float64(typed), true
	case int64:
		return float64(typed), true
	case string:
		parsed, err := strconv.ParseFloat(strings.TrimSpace(typed), 64)
		if err == nil {
			return parsed, true
		}
	}
	return 0, false
}
