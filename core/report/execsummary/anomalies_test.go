package execsummary

import (
	"math"
	"testing"
	"time"
)

func TestTopGrowingAccounts(t *testing.T) {
	enriched := []EnrichedCostRecord{
		// Account 111 - growing
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Cost: 1000, Service: "EC2"}, AccountName: "Account A", RefinedCategory: "HCP"},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "111111111111", Cost: 1800, Service: "EC2"}, AccountName: "Account A", RefinedCategory: "HCP"},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "111111111111", Cost: 200, Service: "S3"}, AccountName: "Account A", RefinedCategory: "HCP"},
		// Account 222 - slight growth
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "222222222222", Cost: 500, Service: "RDS"}, AccountName: "Account B", RefinedCategory: "Tooling"},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "222222222222", Cost: 600, Service: "RDS"}, AccountName: "Account B", RefinedCategory: "Tooling"},
		// Account 333 - decline
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "333333333333", Cost: 800, Service: "Lambda"}, AccountName: "Account C", RefinedCategory: "Dev"},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "333333333333", Cost: 400, Service: "Lambda"}, AccountName: "Account C", RefinedCategory: "Dev"},
	}

	months := []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	results := TopGrowingAccounts(enriched, months, 3, 2)

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Results should be sorted by delta descending
	// Account 111: 2000 - 1000 = +1000 (largest growth)
	// Account 222: 600 - 500 = +100
	// Account 333: 400 - 800 = -400 (decline)

	if results[0].AccountID != "111111111111" {
		t.Errorf("Top growing account should be 111111111111, got %s", results[0].AccountID)
	}

	if results[0].Delta != 1000 {
		t.Errorf("Account 111 delta = %.2f, want 1000", results[0].Delta)
	}

	if results[0].LastMonthCost != 2000 {
		t.Errorf("Account 111 last month = %.2f, want 2000", results[0].LastMonthCost)
	}

	if results[0].PrevMonthCost != 1000 {
		t.Errorf("Account 111 prev month = %.2f, want 1000", results[0].PrevMonthCost)
	}

	if results[0].AccountName != "Account A" {
		t.Errorf("Account 111 name = %q, want %q", results[0].AccountName, "Account A")
	}

	// Check top services
	if len(results[0].TopServices) != 2 {
		t.Errorf("Expected 2 top services, got %d", len(results[0].TopServices))
	}

	if len(results[0].TopServices) > 0 && results[0].TopServices[0].Service != "EC2" {
		t.Errorf("Top service should be EC2, got %s", results[0].TopServices[0].Service)
	}
}

func TestTopGrowingAccountsInsufficientData(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Cost: 1000}},
	}

	months := []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	results := TopGrowingAccounts(enriched, months, 5, 3)

	if results != nil {
		t.Errorf("Expected nil for insufficient months, got %d results", len(results))
	}
}

func TestStatisticalAnomalies(t *testing.T) {
	enriched := []EnrichedCostRecord{
		// Account 111 - stable then spike
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Cost: 1000}, AccountName: "Spiky Account", RefinedCategory: "HCP"},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "111111111111", Cost: 1100}, AccountName: "Spiky Account", RefinedCategory: "HCP"},
		{CostRecord: CostRecord{Month: "2026-03", AccountID: "111111111111", Cost: 1050}, AccountName: "Spiky Account", RefinedCategory: "HCP"},
		{CostRecord: CostRecord{Month: "2026-04", AccountID: "111111111111", Cost: 5000}, AccountName: "Spiky Account", RefinedCategory: "HCP"}, // Spike!
		// Account 222 - stable (no anomaly)
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "222222222222", Cost: 500}, AccountName: "Stable Account"},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "222222222222", Cost: 510}, AccountName: "Stable Account"},
		{CostRecord: CostRecord{Month: "2026-03", AccountID: "222222222222", Cost: 495}, AccountName: "Stable Account"},
		{CostRecord: CostRecord{Month: "2026-04", AccountID: "222222222222", Cost: 505}, AccountName: "Stable Account"},
		// Account 333 - drop
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "333333333333", Cost: 2000}, AccountName: "Dropping Account"},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "333333333333", Cost: 2100}, AccountName: "Dropping Account"},
		{CostRecord: CostRecord{Month: "2026-03", AccountID: "333333333333", Cost: 1950}, AccountName: "Dropping Account"},
		{CostRecord: CostRecord{Month: "2026-04", AccountID: "333333333333", Cost: 100}, AccountName: "Dropping Account"}, // Drop!
	}

	months := []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
	}

	results := StatisticalAnomalies(enriched, months, 2.0)

	// Should detect account 111 (spike) and 333 (drop)
	if len(results) < 2 {
		t.Fatalf("Expected at least 2 anomalies, got %d", len(results))
	}

	// Results sorted by absolute z-score
	// Check that anomalies were detected
	foundSpike := false
	foundDrop := false

	for _, anomaly := range results {
		if anomaly.AccountID == "111111111111" {
			foundSpike = true
			if anomaly.Direction != "spike" {
				t.Errorf("Account 111 direction = %s, want spike", anomaly.Direction)
			}
			if anomaly.CurrentCost != 5000 {
				t.Errorf("Account 111 current cost = %.2f, want 5000", anomaly.CurrentCost)
			}
			if math.Abs(anomaly.ZScore) < 2.0 {
				t.Errorf("Account 111 z-score = %.2f, should be >= 2.0", anomaly.ZScore)
			}
		}

		if anomaly.AccountID == "333333333333" {
			foundDrop = true
			if anomaly.Direction != "drop" {
				t.Errorf("Account 333 direction = %s, want drop", anomaly.Direction)
			}
			if anomaly.CurrentCost != 100 {
				t.Errorf("Account 333 current cost = %.2f, want 100", anomaly.CurrentCost)
			}
			if anomaly.ZScore >= 0 {
				t.Errorf("Account 333 z-score = %.2f, should be negative", anomaly.ZScore)
			}
		}
	}

	if !foundSpike {
		t.Error("Did not find spike anomaly for account 111111111111")
	}

	if !foundDrop {
		t.Error("Did not find drop anomaly for account 333333333333")
	}

	// Account 222 should not appear (stable)
	for _, anomaly := range results {
		if anomaly.AccountID == "222222222222" {
			t.Error("Stable account 222 should not be flagged as anomaly")
		}
	}
}

func TestStatisticalAnomaliesInsufficientHistory(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Cost: 1000}},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "111111111111", Cost: 5000}},
	}

	months := []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	results := StatisticalAnomalies(enriched, months, 2.0)

	// Need at least 3 months
	if results != nil {
		t.Errorf("Expected nil for insufficient months, got %d results", len(results))
	}
}

func TestStatisticalAnomaliesLowVariance(t *testing.T) {
	// Account with very stable costs (stddev < $10)
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Cost: 5.00}},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "111111111111", Cost: 5.50}},
		{CostRecord: CostRecord{Month: "2026-03", AccountID: "111111111111", Cost: 5.25}},
		{CostRecord: CostRecord{Month: "2026-04", AccountID: "111111111111", Cost: 10.00}}, // Spike but too low variance
	}

	months := []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
	}

	results := StatisticalAnomalies(enriched, months, 2.0)

	// Should be filtered out due to stddev < $10
	if len(results) > 0 {
		t.Errorf("Expected no anomalies for low variance account, got %d", len(results))
	}
}

func TestMean(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
	}{
		{"simple", []float64{1, 2, 3, 4, 5}, 3.0},
		{"single", []float64{10}, 10.0},
		{"empty", []float64{}, 0.0},
		{"decimals", []float64{1.5, 2.5, 3.5}, 2.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mean(tt.values)
			if result != tt.expected {
				t.Errorf("mean(%v) = %f, want %f", tt.values, result, tt.expected)
			}
		})
	}
}

func TestStddev(t *testing.T) {
	tests := []struct {
		name     string
		values   []float64
		expected float64
		epsilon  float64
	}{
		{"simple", []float64{2, 4, 4, 4, 5, 5, 7, 9}, 2.138, 0.01},
		{"single", []float64{10}, 0.0, 0.0},
		{"two values", []float64{10, 20}, 7.071, 0.01},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mean(tt.values)
			result := stddev(tt.values, m)
			if math.Abs(result-tt.expected) > tt.epsilon {
				t.Errorf("stddev(%v) = %f, want %f (±%f)", tt.values, result, tt.expected, tt.epsilon)
			}
		})
	}
}

func TestAccountCostDeltas(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Cost: 500}},
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Cost: 500}},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "111111111111", Cost: 800}},
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "222222222222", Cost: 300}},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "222222222222", Cost: 250}},
	}

	deltas, last, prev := accountCostDeltas(enriched, "2026-02", "2026-01")

	// Account 111: 800 - 1000 = -200
	if deltas["111111111111"] != -200 {
		t.Errorf("Account 111 delta = %.2f, want -200", deltas["111111111111"])
	}

	// Account 222: 250 - 300 = -50
	if deltas["222222222222"] != -50 {
		t.Errorf("Account 222 delta = %.2f, want -50", deltas["222222222222"])
	}

	// Check last month totals
	if last["111111111111"] != 800 {
		t.Errorf("Account 111 last = %.2f, want 800", last["111111111111"])
	}

	// Check prev month totals
	if prev["111111111111"] != 1000 {
		t.Errorf("Account 111 prev = %.2f, want 1000", prev["111111111111"])
	}
}

func TestGetAccountMetadata(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{
			CostRecord:      CostRecord{AccountID: "111111111111"},
			AccountName:     "Test Account",
			RefinedCategory: "HCP",
			OwnerTeam:       "Platform Team",
		},
	}

	meta := getAccountMetadata(enriched, "111111111111")

	if meta.AccountName != "Test Account" {
		t.Errorf("AccountName = %q, want %q", meta.AccountName, "Test Account")
	}
	if meta.Category != "HCP" {
		t.Errorf("Category = %q, want %q", meta.Category, "HCP")
	}
	if meta.Owner != "Platform Team" {
		t.Errorf("Owner = %q, want %q", meta.Owner, "Platform Team")
	}
}

func TestGetAccountMetadataNotFound(t *testing.T) {
	enriched := []EnrichedCostRecord{}

	meta := getAccountMetadata(enriched, "999999999999")

	if meta.AccountName != "999999999999" {
		t.Errorf("AccountName = %q, want account ID", meta.AccountName)
	}
	if meta.Category != "Unmapped" {
		t.Errorf("Category = %q, want Unmapped", meta.Category)
	}
	if meta.Owner != "" {
		t.Errorf("Owner = %q, want empty string", meta.Owner)
	}
}
