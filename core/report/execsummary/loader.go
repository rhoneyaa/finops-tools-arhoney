package execsummary

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// LoadCostTransport reads cost data from CSV or JSON and returns indexed CostData.
// Supports both array of records and {"costs": [...]} envelope formats.
func LoadCostTransport(path string) (*CostData, error) {
	ext := strings.ToLower(filepath.Ext(path))

	var records []CostRecord
	var err error

	if ext == ".csv" {
		records, err = loadCostCSV(path)
	} else {
		records, err = loadCostJSON(path)
	}

	if err != nil {
		return nil, fmt.Errorf("cost transport parse failed (%s): %w", filepath.Base(path), err)
	}

	if len(records) == 0 {
		return nil, fmt.Errorf("cost transport produced zero rows: %s", path)
	}

	return indexCostData(records), nil
}

// LoadSavingsPlansTransport reads savings plan coverage and utilization from JSON.
func LoadSavingsPlansTransport(path string) (*SPData, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("SP transport: %w", err)
	}
	defer file.Close()

	var raw map[string]interface{}
	if err := json.NewDecoder(file).Decode(&raw); err != nil {
		return nil, fmt.Errorf("SP transport parse failed (%s): %w", filepath.Base(path), err)
	}

	result := &SPData{
		Coverage:    parseSPMetrics(raw["coverage"]),
		Utilization: parseSPMetrics(raw["utilization"]),
		SPByPayer:   make(map[string]PayerSPData),
	}

	// Parse sp_by_payer if present
	if spByPayer, ok := raw["sp_by_payer"].(map[string]interface{}); ok {
		for payer, data := range spByPayer {
			if payerData, ok := data.(map[string]interface{}); ok {
				result.SPByPayer[payer] = PayerSPData{
					Coverage:    parseSPMetrics(payerData["coverage"]),
					Utilization: parseSPMetrics(payerData["utilization"]),
				}
			}
		}
	}

	if len(result.Coverage) == 0 && len(result.Utilization) == 0 {
		return nil, fmt.Errorf("SP transport has no coverage or utilization records: %s", path)
	}

	return result, nil
}

// LoadClusterCounts loads month-to-cluster count mapping from CSV.
// Tries multiple column strategies: (month, cluster_count), (month, total_count), or created_date grouping.
func LoadClusterCounts(path string) (ClusterCounts, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("cluster counts CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("cluster counts CSV parse failed: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("cluster counts CSV has no data rows")
	}

	header := records[0]
	data := records[1:]

	// Try month + cluster_count/total_count columns
	if counts := clusterCountsFromTotalColumns(header, data); counts != nil {
		return counts, nil
	}

	// Fallback: aggregate by created_date
	if counts := clusterCountsFromCreatedDate(header, data); counts != nil {
		return counts, nil
	}

	return nil, fmt.Errorf("cluster counts CSV has no usable month/count columns: %s", path)
}

// LoadEnvClusterCounts loads latest per-environment cluster totals from CSV.
func LoadEnvClusterCounts(path string) (EnvCounts, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("env cluster counts CSV: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("env cluster counts CSV parse failed: %w", err)
	}

	if len(records) < 2 {
		return EnvCounts{}, nil
	}

	header := records[0]
	prodIdx := indexOf(header, "prod_count")
	if prodIdx == -1 {
		return EnvCounts{}, nil
	}

	// Select live_ocm row if available, else last row
	var selectedRow []string
	prodSourceIdx := indexOf(header, "prod_source")
	if prodSourceIdx != -1 {
		for i := len(records) - 1; i >= 1; i-- {
			if records[i][prodSourceIdx] == "live_ocm" {
				selectedRow = records[i]
				break
			}
		}
	}
	if selectedRow == nil {
		selectedRow = records[len(records)-1]
	}

	return EnvCounts{
		"prod":  safeInt(getColumn(selectedRow, header, "prod_count")),
		"stg":   safeInt(getColumn(selectedRow, header, "stg_count")),
		"int":   safeInt(getColumn(selectedRow, header, "int_count")),
		"total": safeInt(getColumn(selectedRow, header, "total_count")),
		"month": getColumn(selectedRow, header, "month"),
	}, nil
}

// loadCostCSV reads cost records from CSV file.
func loadCostCSV(path string) ([]CostRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	header, err := reader.Read()
	if err != nil {
		return nil, err
	}

	monthIdx := indexOf(header, "month")
	accountIdx := indexOf(header, "account_id")
	costIdx := indexOf(header, "cost")
	serviceIdx := indexOf(header, "service")
	payerIdx := indexOf(header, "payer")

	if monthIdx == -1 || accountIdx == -1 || costIdx == -1 {
		return nil, fmt.Errorf("missing required columns: month, account_id, cost")
	}

	var records []CostRecord
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		cost, _ := strconv.ParseFloat(row[costIdx], 64)
		record := CostRecord{
			Month:     row[monthIdx],
			AccountID: row[accountIdx],
			Cost:      cost,
		}

		if serviceIdx != -1 && serviceIdx < len(row) {
			record.Service = row[serviceIdx]
		}
		if payerIdx != -1 && payerIdx < len(row) {
			record.Payer = row[payerIdx]
		}

		records = append(records, record)
	}

	return records, nil
}

// loadCostJSON reads cost records from JSON file.
// Supports array format or {"costs": [...]} envelope.
func loadCostJSON(path string) ([]CostRecord, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var raw interface{}
	if err := json.NewDecoder(file).Decode(&raw); err != nil {
		return nil, err
	}

	var items []interface{}
	switch v := raw.(type) {
	case []interface{}:
		items = v
	case map[string]interface{}:
		if costs, ok := v["costs"].([]interface{}); ok {
			items = costs
		} else {
			return nil, fmt.Errorf("JSON envelope missing 'costs' array")
		}
	default:
		return nil, fmt.Errorf("unsupported JSON structure")
	}

	var records []CostRecord
	for _, item := range items {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		record := CostRecord{
			Month:     getString(obj, "month"),
			AccountID: getString(obj, "account_id"),
			Cost:      getFloat(obj, "cost"),
			Service:   getString(obj, "service"),
			Payer:     getString(obj, "payer"),
		}
		records = append(records, record)
	}

	return records, nil
}

// indexCostData creates indexed lookups for efficient aggregation.
func indexCostData(records []CostRecord) *CostData {
	data := &CostData{
		Records:        records,
		ByMonth:        make(map[string][]CostRecord),
		ByAccount:      make(map[string][]CostRecord),
		ByMonthAccount: make(map[string]map[string]float64),
	}

	for _, rec := range records {
		data.ByMonth[rec.Month] = append(data.ByMonth[rec.Month], rec)
		data.ByAccount[rec.AccountID] = append(data.ByAccount[rec.AccountID], rec)

		if data.ByMonthAccount[rec.Month] == nil {
			data.ByMonthAccount[rec.Month] = make(map[string]float64)
		}
		data.ByMonthAccount[rec.Month][rec.AccountID] += rec.Cost
	}

	return data
}

// parseSPMetrics converts interface{} to []SavingsPlanMetric.
func parseSPMetrics(raw interface{}) []SavingsPlanMetric {
	arr, ok := raw.([]interface{})
	if !ok {
		return nil
	}

	var metrics []SavingsPlanMetric
	for _, item := range arr {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		metric := SavingsPlanMetric{
			Month:       getString(obj, "month"),
			Coverage:    getFloat(obj, "coverage"),
			Utilization: getFloat(obj, "utilization"),
			Payer:       getString(obj, "payer"),
		}
		metrics = append(metrics, metric)
	}

	return metrics
}

// clusterCountsFromTotalColumns extracts counts from month + cluster_count/total_count columns.
func clusterCountsFromTotalColumns(header []string, data [][]string) ClusterCounts {
	monthIdx := indexOf(header, "month")
	clusterIdx := indexOf(header, "cluster_count")
	totalIdx := indexOf(header, "total_count")

	if monthIdx == -1 || (clusterIdx == -1 && totalIdx == -1) {
		return nil
	}

	countIdx := totalIdx
	if countIdx == -1 {
		countIdx = clusterIdx
	}

	counts := make(ClusterCounts)
	for _, row := range data {
		if len(row) <= monthIdx || len(row) <= countIdx {
			continue
		}

		month := strings.TrimSpace(row[monthIdx])
		countStr := strings.TrimSpace(row[countIdx])
		if month == "" || countStr == "" {
			continue
		}

		if count, err := strconv.Atoi(countStr); err == nil {
			counts[month] = count
		}
	}

	if len(counts) == 0 {
		return nil
	}
	return counts
}

// clusterCountsFromCreatedDate aggregates counts by month from created_date column.
func clusterCountsFromCreatedDate(header []string, data [][]string) ClusterCounts {
	createdIdx := indexOf(header, "created_date")
	if createdIdx == -1 {
		return nil
	}

	counts := make(ClusterCounts)
	for _, row := range data {
		if len(row) <= createdIdx {
			continue
		}

		dateStr := strings.TrimSpace(row[createdIdx])
		if dateStr == "" {
			continue
		}

		// Extract YYYY-MM from date string
		month := extractMonth(dateStr)
		if month != "" {
			counts[month]++
		}
	}

	if len(counts) == 0 {
		return nil
	}
	return counts
}

// Utility functions

func indexOf(slice []string, value string) int {
	for i, v := range slice {
		if v == value {
			return i
		}
	}
	return -1
}

func getColumn(row []string, header []string, column string) string {
	idx := indexOf(header, column)
	if idx == -1 || idx >= len(row) {
		return ""
	}
	return row[idx]
}

func getString(obj map[string]interface{}, key string) string {
	if v, ok := obj[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFloat(obj map[string]interface{}, key string) float64 {
	if v, ok := obj[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case int:
			return float64(val)
		case string:
			if f, err := strconv.ParseFloat(val, 64); err == nil {
				return f
			}
		}
	}
	return 0
}

func safeInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	i, _ := strconv.Atoi(s)
	return i
}

func extractMonth(dateStr string) string {
	// Try to extract YYYY-MM from various date formats
	// Supports: YYYY-MM-DD, YYYY-MM-DDTHH:MM:SS, etc.
	parts := strings.Split(dateStr, "T")
	datePart := parts[0]

	segments := strings.Split(datePart, "-")
	if len(segments) >= 2 {
		return segments[0] + "-" + segments[1]
	}

	return ""
}
