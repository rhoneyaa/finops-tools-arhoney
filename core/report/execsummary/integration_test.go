package execsummary

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestIntegrationEndToEnd tests the full workflow from loading to enrichment.
// This simulates a real executive summary report generation.
func TestIntegrationEndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tmpDir := t.TempDir()

	// Create sample cost data CSV
	costCSV := filepath.Join(tmpDir, "costs.csv")
	costData := `month,account_id,cost,service,payer
2026-01,111111111111,1500.50,EC2,payer1
2026-01,111111111111,800.25,S3,payer1
2026-01,222222222222,2300.75,RDS,payer1
2026-01,333333333333,950.00,Lambda,payer1
2026-02,111111111111,1650.00,EC2,payer1
2026-02,111111111111,850.50,S3,payer1
2026-02,222222222222,2500.00,RDS,payer1
2026-02,333333333333,1050.00,Lambda,payer1
2026-03,111111111111,1800.00,EC2,payer1`

	if err := os.WriteFile(costCSV, []byte(costData), 0644); err != nil {
		t.Fatal(err)
	}

	// Create account mapping CSV
	mappingCSV := filepath.Join(tmpDir, "mappings.csv")
	mappingData := `account_id,refined_category,sub_type,owner_team,account_name
111111111111,HCP,Production,Team A,HCP Production Account
222222222222,Tooling,Shared,Team B,Shared Tooling
333333333333,Development,Sandbox,Team C,Dev Sandbox`

	if err := os.WriteFile(mappingCSV, []byte(mappingData), 0644); err != nil {
		t.Fatal(err)
	}

	// Step 1: Load cost transport
	t.Log("Step 1: Loading cost data...")
	costDataLoaded, err := LoadCostTransport(costCSV)
	if err != nil {
		t.Fatalf("LoadCostTransport failed: %v", err)
	}

	if len(costDataLoaded.Records) != 9 {
		t.Errorf("Expected 9 cost records, got %d", len(costDataLoaded.Records))
	}

	t.Logf("  ✓ Loaded %d cost records", len(costDataLoaded.Records))
	t.Logf("  ✓ Indexed by %d months", len(costDataLoaded.ByMonth))
	t.Logf("  ✓ Indexed by %d accounts", len(costDataLoaded.ByAccount))

	// Step 2: Load account mappings
	t.Log("Step 2: Loading account mappings...")
	mappings, err := loadMappingsFromCSV(mappingCSV)
	if err != nil {
		t.Fatalf("loadMappingsFromCSV failed: %v", err)
	}

	if len(mappings) != 3 {
		t.Errorf("Expected 3 mappings, got %d", len(mappings))
	}

	t.Logf("  ✓ Loaded %d account mappings", len(mappings))

	// Step 3: Enrich cost data
	t.Log("Step 3: Enriching cost data with mappings...")
	enriched := EnrichWithMapping(costDataLoaded, mappings)

	if len(enriched) != 9 {
		t.Errorf("Expected 9 enriched records, got %d", len(enriched))
	}

	// Verify enrichment worked
	var hcpCount, toolingCount, devCount int
	for _, rec := range enriched {
		switch rec.RefinedCategory {
		case "HCP":
			hcpCount++
		case "Tooling":
			toolingCount++
		case "Development":
			devCount++
		}
	}

	t.Logf("  ✓ HCP: %d records", hcpCount)
	t.Logf("  ✓ Tooling: %d records", toolingCount)
	t.Logf("  ✓ Development: %d records", devCount)

	if hcpCount != 5 {
		t.Errorf("Expected 5 HCP records, got %d", hcpCount)
	}

	// Step 4: Compute time window
	t.Log("Step 4: Computing 3-month window...")
	today := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	window, err := ComputeWindow(3, &today)
	if err != nil {
		t.Fatalf("ComputeWindow failed: %v", err)
	}

	t.Logf("  ✓ Window: %s to %s", MonthLabel(window.Start), MonthLabel(window.End))
	t.Logf("  ✓ Months in window: %d", len(window.Months))

	// Step 5: Category monthly aggregation
	t.Log("Step 5: Aggregating by category and month...")
	categoryMonthly := CategoryMonthly(enriched, window.Months)

	t.Logf("  ✓ Generated %d category-month records", len(categoryMonthly))

	for _, rec := range categoryMonthly {
		t.Logf("    - %s %s: $%.2f", rec.RefinedCategory, rec.Month, rec.Cost)
	}

	// Step 6: Account monthly totals
	t.Log("Step 6: Computing account monthly totals...")
	accountTotals := AccountMonthlyTotal(enriched)

	for accountID, months := range accountTotals {
		for month, cost := range months {
			t.Logf("    - Account %s, %s: $%.2f", accountID, month, cost)
		}
	}

	// Step 7: Top services for an account
	t.Log("Step 7: Finding top services for HCP account in Jan 2026...")
	topServices := ServiceCostsForAccount(enriched, "111111111111", "2026-01", 3)

	t.Logf("  ✓ Top services:")
	for i, svc := range topServices {
		t.Logf("    %d. %s: $%.2f", i+1, svc.Service, svc.Cost)
	}

	// Step 8: Total cost
	t.Log("Step 8: Computing total cost...")
	total := TotalCost(enriched)
	t.Logf("  ✓ Total cost across all records: $%.2f", total)

	// Verify total
	expectedTotal := 1500.50 + 800.25 + 2300.75 + 950.00 + // Jan
		1650.00 + 850.50 + 2500.00 + 1050.00 + // Feb
		1800.00 // Mar
	if total != expectedTotal {
		t.Errorf("Total cost = $%.2f, want $%.2f", total, expectedTotal)
	}

	t.Log("\n✅ Integration test complete - all components working together!")
}

// Helper to load mappings from CSV
func loadMappingsFromCSV(path string) ([]AccountMapping, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	records, err := readCSVRecords(file)
	if err != nil {
		return nil, err
	}

	if len(records) < 2 {
		return nil, nil
	}

	header := records[0]
	data := records[1:]

	var mappings []AccountMapping
	for _, row := range data {
		mapping := AccountMapping{
			AccountID:       getColumnValue(row, header, "account_id"),
			RefinedCategory: getColumnValue(row, header, "refined_category"),
			SubType:         getColumnValue(row, header, "sub_type"),
			OwnerTeam:       getColumnValue(row, header, "owner_team"),
			AccountName:     getColumnValue(row, header, "account_name"),
		}
		mappings = append(mappings, mapping)
	}

	return mappings, nil
}

func readCSVRecords(file *os.File) ([][]string, error) {
	// This is a simplified CSV reader - in production we'd use encoding/csv
	content, err := os.ReadFile(file.Name())
	if err != nil {
		return nil, err
	}

	lines := splitLines(string(content))
	var records [][]string
	for _, line := range lines {
		if line == "" {
			continue
		}
		records = append(records, splitCSVLine(line))
	}

	return records, nil
}

func splitLines(s string) []string {
	var lines []string
	current := ""
	for _, ch := range s {
		if ch == '\n' || ch == '\r' {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
		} else {
			current += string(ch)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitCSVLine(line string) []string {
	var fields []string
	current := ""
	for _, ch := range line {
		if ch == ',' {
			fields = append(fields, current)
			current = ""
		} else {
			current += string(ch)
		}
	}
	fields = append(fields, current)
	return fields
}

func getColumnValue(row []string, header []string, column string) string {
	for i, h := range header {
		if h == column && i < len(row) {
			return row[i]
		}
	}
	return ""
}
