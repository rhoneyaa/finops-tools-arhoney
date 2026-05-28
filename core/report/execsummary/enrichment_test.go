package execsummary

import (
	"testing"
	"time"
)

func TestEnrichWithMapping(t *testing.T) {
	costData := &CostData{
		Records: []CostRecord{
			{Month: "2026-01", AccountID: "111111111111", Cost: 1500.50, Service: "EC2"},
			{Month: "2026-01", AccountID: "222222222222", Cost: 2300.75, Service: "S3"},
			{Month: "2026-01", AccountID: "333333333333", Cost: 850.00, Service: "RDS"},
		},
	}

	mappings := []AccountMapping{
		{
			AccountID:       "111111111111",
			RefinedCategory: "HCP",
			SubType:         "Production",
			OwnerTeam:       "Team A",
			AccountName:     "HCP Prod",
		},
		{
			AccountID:       "222222222222",
			RefinedCategory: "Tooling",
			SubType:         "",
			OwnerTeam:       "Team B",
			AccountName:     "Tooling Account",
		},
		// 333333333333 deliberately unmapped
	}

	enriched := EnrichWithMapping(costData, mappings)

	if len(enriched) != 3 {
		t.Fatalf("Expected 3 enriched records, got %d", len(enriched))
	}

	// Test mapped account
	rec0 := enriched[0]
	if rec0.RefinedCategory != "HCP" {
		t.Errorf("Record[0].RefinedCategory = %q, want %q", rec0.RefinedCategory, "HCP")
	}
	if rec0.SubType != "Production" {
		t.Errorf("Record[0].SubType = %q, want %q", rec0.SubType, "Production")
	}
	if rec0.OwnerTeam != "Team A" {
		t.Errorf("Record[0].OwnerTeam = %q, want %q", rec0.OwnerTeam, "Team A")
	}
	if rec0.AccountName != "HCP Prod" {
		t.Errorf("Record[0].AccountName = %q, want %q", rec0.AccountName, "HCP Prod")
	}

	// Test unmapped account
	rec2 := enriched[2]
	if rec2.RefinedCategory != "Unmapped" {
		t.Errorf("Record[2].RefinedCategory = %q, want %q", rec2.RefinedCategory, "Unmapped")
	}
	if rec2.AccountName != "333333333333" {
		t.Errorf("Record[2].AccountName = %q, want %q (account ID)", rec2.AccountName, "333333333333")
	}
	if rec2.OwnerTeam != "" {
		t.Errorf("Record[2].OwnerTeam = %q, want empty string", rec2.OwnerTeam)
	}
}

func TestEnrichWithMappingEmptyCategory(t *testing.T) {
	costData := &CostData{
		Records: []CostRecord{
			{Month: "2026-01", AccountID: "111111111111", Cost: 100.0},
		},
	}

	mappings := []AccountMapping{
		{
			AccountID:       "111111111111",
			RefinedCategory: "", // Empty category
			AccountName:     "Test Account",
		},
	}

	enriched := EnrichWithMapping(costData, mappings)

	if enriched[0].RefinedCategory != "Unmapped" {
		t.Errorf("Empty category should normalize to 'Unmapped', got %q", enriched[0].RefinedCategory)
	}
}

func TestEnrichWithMappingEmptyData(t *testing.T) {
	costData := &CostData{Records: []CostRecord{}}
	mappings := []AccountMapping{{AccountID: "111111111111", RefinedCategory: "HCP"}}

	enriched := EnrichWithMapping(costData, mappings)

	if enriched != nil {
		t.Errorf("Expected nil for empty data, got %d records", len(enriched))
	}
}

func TestCategoryMonthly(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Month: "2026-01", Cost: 1000}, RefinedCategory: "HCP"},
		{CostRecord: CostRecord{Month: "2026-01", Cost: 500}, RefinedCategory: "HCP"},
		{CostRecord: CostRecord{Month: "2026-01", Cost: 800}, RefinedCategory: "Tooling"},
		{CostRecord: CostRecord{Month: "2026-02", Cost: 1200}, RefinedCategory: "HCP"},
		{CostRecord: CostRecord{Month: "2026-03", Cost: 300}, RefinedCategory: "HCP"}, // Outside window
	}

	months := []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	results := CategoryMonthly(enriched, months)

	// Should have 3 records: HCP-2026-01, Tooling-2026-01, HCP-2026-02
	if len(results) != 3 {
		t.Fatalf("Expected 3 category-month records, got %d", len(results))
	}

	// Find HCP 2026-01
	var hcp0126 *CategoryMonthRecord
	for i := range results {
		if results[i].RefinedCategory == "HCP" && results[i].Month == "2026-01" {
			hcp0126 = &results[i]
			break
		}
	}

	if hcp0126 == nil {
		t.Fatal("Expected HCP 2026-01 record not found")
	}

	if hcp0126.Cost != 1500 {
		t.Errorf("HCP 2026-01 cost = %f, want 1500", hcp0126.Cost)
	}

	// Verify 2026-03 is excluded
	for _, rec := range results {
		if rec.Month == "2026-03" {
			t.Error("2026-03 should be excluded from results")
		}
	}
}

func TestCategoryMonthlySorting(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Month: "2026-02", Cost: 100}, RefinedCategory: "Tooling"},
		{CostRecord: CostRecord{Month: "2026-01", Cost: 200}, RefinedCategory: "Tooling"},
		{CostRecord: CostRecord{Month: "2026-02", Cost: 300}, RefinedCategory: "HCP"},
		{CostRecord: CostRecord{Month: "2026-01", Cost: 400}, RefinedCategory: "HCP"},
	}

	months := []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	results := CategoryMonthly(enriched, months)

	// Results should be sorted by category, then month
	// Expected order: HCP-2026-01, HCP-2026-02, Tooling-2026-01, Tooling-2026-02
	expected := []struct {
		category string
		month    string
	}{
		{"HCP", "2026-01"},
		{"HCP", "2026-02"},
		{"Tooling", "2026-01"},
		{"Tooling", "2026-02"},
	}

	if len(results) != len(expected) {
		t.Fatalf("Expected %d results, got %d", len(expected), len(results))
	}

	for i, exp := range expected {
		if results[i].RefinedCategory != exp.category || results[i].Month != exp.month {
			t.Errorf("Result[%d] = %s-%s, want %s-%s",
				i, results[i].RefinedCategory, results[i].Month, exp.category, exp.month)
		}
	}
}

func TestAccountMonthlyTotal(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Cost: 500}},
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Cost: 300}},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "111111111111", Cost: 700}},
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "222222222222", Cost: 900}},
	}

	totals := AccountMonthlyTotal(enriched)

	if len(totals) != 2 {
		t.Fatalf("Expected 2 accounts, got %d", len(totals))
	}

	if totals["111111111111"]["2026-01"] != 800 {
		t.Errorf("Account 111111111111 2026-01 = %f, want 800", totals["111111111111"]["2026-01"])
	}

	if totals["111111111111"]["2026-02"] != 700 {
		t.Errorf("Account 111111111111 2026-02 = %f, want 700", totals["111111111111"]["2026-02"])
	}

	if totals["222222222222"]["2026-01"] != 900 {
		t.Errorf("Account 222222222222 2026-01 = %f, want 900", totals["222222222222"]["2026-01"])
	}
}

func TestServiceCostsForAccount(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Service: "EC2", Cost: 1000}},
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Service: "S3", Cost: 500}},
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Service: "RDS", Cost: 300}},
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Service: "Lambda", Cost: 100}},
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "222222222222", Service: "EC2", Cost: 9999}}, // Different account
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "111111111111", Service: "EC2", Cost: 9999}}, // Different month
	}

	services := ServiceCostsForAccount(enriched, "111111111111", "2026-01", 3)

	if len(services) != 3 {
		t.Fatalf("Expected 3 services, got %d", len(services))
	}

	// Should be sorted descending by cost: EC2, S3, RDS
	if services[0].Service != "EC2" || services[0].Cost != 1000 {
		t.Errorf("Top service = %s ($%.2f), want EC2 ($1000)", services[0].Service, services[0].Cost)
	}

	if services[1].Service != "S3" || services[1].Cost != 500 {
		t.Errorf("2nd service = %s ($%.2f), want S3 ($500)", services[1].Service, services[1].Cost)
	}

	if services[2].Service != "RDS" || services[2].Cost != 300 {
		t.Errorf("3rd service = %s ($%.2f), want RDS ($300)", services[2].Service, services[2].Cost)
	}
}

func TestFilterByMonth(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Month: "2026-01"}},
		{CostRecord: CostRecord{Month: "2026-02"}},
		{CostRecord: CostRecord{Month: "2026-01"}},
	}

	filtered := FilterByMonth(enriched, "2026-01")

	if len(filtered) != 2 {
		t.Errorf("Expected 2 records for 2026-01, got %d", len(filtered))
	}
}

func TestFilterByPayer(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Payer: "payer1"}},
		{CostRecord: CostRecord{Payer: "payer2"}},
		{CostRecord: CostRecord{Payer: "payer1"}},
	}

	filtered := FilterByPayer(enriched, "payer1")

	if len(filtered) != 2 {
		t.Errorf("Expected 2 records for payer1, got %d", len(filtered))
	}

	// Test "all" returns everything
	filteredAll := FilterByPayer(enriched, "all")
	if len(filteredAll) != 3 {
		t.Errorf("Expected all 3 records for 'all', got %d", len(filteredAll))
	}
}

func TestFilterByAccounts(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{AccountID: "111111111111"}},
		{CostRecord: CostRecord{AccountID: "222222222222"}},
		{CostRecord: CostRecord{AccountID: "333333333333"}},
	}

	accounts := map[string]bool{
		"111111111111": true,
		"333333333333": true,
	}

	filtered := FilterByAccounts(enriched, accounts)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered records, got %d", len(filtered))
	}

	// Verify correct accounts
	for _, rec := range filtered {
		if rec.AccountID == "222222222222" {
			t.Error("Account 222222222222 should be filtered out")
		}
	}
}

func TestTotalCost(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Cost: 1000.50}},
		{CostRecord: CostRecord{Cost: 2500.25}},
		{CostRecord: CostRecord{Cost: 750.00}},
	}

	total := TotalCost(enriched)

	expected := 4250.75
	if total != expected {
		t.Errorf("TotalCost = %f, want %f", total, expected)
	}
}

func TestTotalCostEmpty(t *testing.T) {
	total := TotalCost(nil)
	if total != 0 {
		t.Errorf("TotalCost(nil) = %f, want 0", total)
	}
}
