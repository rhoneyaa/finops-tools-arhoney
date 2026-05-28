package execsummary

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCostTransportCSV(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "costs.csv")

	csvContent := `month,account_id,cost,service,payer
2026-01,123456789012,1500.50,EC2,payer1
2026-01,987654321098,2300.75,S3,payer1
2026-02,123456789012,1600.00,EC2,payer1`

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := LoadCostTransport(csvPath)
	if err != nil {
		t.Fatalf("LoadCostTransport() error = %v", err)
	}

	if len(data.Records) != 3 {
		t.Errorf("Expected 3 records, got %d", len(data.Records))
	}

	// Verify first record
	rec := data.Records[0]
	if rec.Month != "2026-01" {
		t.Errorf("Record[0].Month = %q, want %q", rec.Month, "2026-01")
	}
	if rec.AccountID != "123456789012" {
		t.Errorf("Record[0].AccountID = %q, want %q", rec.AccountID, "123456789012")
	}
	if rec.Cost != 1500.50 {
		t.Errorf("Record[0].Cost = %f, want %f", rec.Cost, 1500.50)
	}
	if rec.Service != "EC2" {
		t.Errorf("Record[0].Service = %q, want %q", rec.Service, "EC2")
	}

	// Verify indexes
	if len(data.ByMonth["2026-01"]) != 2 {
		t.Errorf("ByMonth[2026-01] has %d records, want 2", len(data.ByMonth["2026-01"]))
	}

	if len(data.ByAccount["123456789012"]) != 2 {
		t.Errorf("ByAccount[123456789012] has %d records, want 2", len(data.ByAccount["123456789012"]))
	}

	if cost := data.ByMonthAccount["2026-01"]["123456789012"]; cost != 1500.50 {
		t.Errorf("ByMonthAccount[2026-01][123456789012] = %f, want %f", cost, 1500.50)
	}
}

func TestLoadCostTransportJSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "costs.json")

	jsonContent := `[
		{"month": "2026-01", "account_id": "123456789012", "cost": 1500.50, "service": "EC2"},
		{"month": "2026-01", "account_id": "987654321098", "cost": 2300.75, "service": "S3"}
	]`

	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := LoadCostTransport(jsonPath)
	if err != nil {
		t.Fatalf("LoadCostTransport() error = %v", err)
	}

	if len(data.Records) != 2 {
		t.Errorf("Expected 2 records, got %d", len(data.Records))
	}
}

func TestLoadCostTransportJSONEnvelope(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "costs-envelope.json")

	jsonContent := `{
		"costs": [
			{"month": "2026-01", "account_id": "123456789012", "cost": 1500.50}
		]
	}`

	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := LoadCostTransport(jsonPath)
	if err != nil {
		t.Fatalf("LoadCostTransport() error = %v", err)
	}

	if len(data.Records) != 1 {
		t.Errorf("Expected 1 record, got %d", len(data.Records))
	}
}

func TestLoadCostTransportCSVMissingColumns(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "invalid.csv")

	csvContent := `month,account_id
2026-01,123456789012`

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadCostTransport(csvPath)
	if err == nil {
		t.Error("Expected error for missing cost column, got nil")
	}
}

func TestLoadCostTransportEmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "empty.csv")

	csvContent := `month,account_id,cost
`

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadCostTransport(csvPath)
	if err == nil {
		t.Error("Expected error for empty data, got nil")
	}
}

func TestLoadSavingsPlansTransport(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "sp.json")

	jsonContent := `{
		"coverage": [
			{"month": "2026-01", "coverage": 85.5, "payer": "payer1"}
		],
		"utilization": [
			{"month": "2026-01", "utilization": 92.3, "payer": "payer1"}
		],
		"sp_by_payer": {
			"payer1": {
				"coverage": [{"month": "2026-01", "coverage": 85.5}],
				"utilization": [{"month": "2026-01", "utilization": 92.3}]
			}
		}
	}`

	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	data, err := LoadSavingsPlansTransport(jsonPath)
	if err != nil {
		t.Fatalf("LoadSavingsPlansTransport() error = %v", err)
	}

	if len(data.Coverage) != 1 {
		t.Errorf("Expected 1 coverage metric, got %d", len(data.Coverage))
	}

	if data.Coverage[0].Coverage != 85.5 {
		t.Errorf("Coverage = %f, want %f", data.Coverage[0].Coverage, 85.5)
	}

	if len(data.Utilization) != 1 {
		t.Errorf("Expected 1 utilization metric, got %d", len(data.Utilization))
	}

	if _, ok := data.SPByPayer["payer1"]; !ok {
		t.Error("Expected payer1 in SPByPayer")
	}
}

func TestLoadSavingsPlansTransportEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	jsonPath := filepath.Join(tmpDir, "sp-empty.json")

	jsonContent := `{
		"coverage": [],
		"utilization": []
	}`

	if err := os.WriteFile(jsonPath, []byte(jsonContent), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadSavingsPlansTransport(jsonPath)
	if err == nil {
		t.Error("Expected error for empty SP data, got nil")
	}
}

func TestLoadClusterCounts(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "clusters.csv")

	csvContent := `month,cluster_count
2026-01,150
2026-02,155
2026-03,160`

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	counts, err := LoadClusterCounts(csvPath)
	if err != nil {
		t.Fatalf("LoadClusterCounts() error = %v", err)
	}

	if len(counts) != 3 {
		t.Errorf("Expected 3 months, got %d", len(counts))
	}

	if counts["2026-01"] != 150 {
		t.Errorf("counts[2026-01] = %d, want 150", counts["2026-01"])
	}
	if counts["2026-02"] != 155 {
		t.Errorf("counts[2026-02] = %d, want 155", counts["2026-02"])
	}
}

func TestLoadClusterCountsTotalColumn(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "clusters-total.csv")

	csvContent := `month,total_count,prod_count
2026-01,200,100
2026-02,210,105`

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	counts, err := LoadClusterCounts(csvPath)
	if err != nil {
		t.Fatalf("LoadClusterCounts() error = %v", err)
	}

	if counts["2026-01"] != 200 {
		t.Errorf("counts[2026-01] = %d, want 200", counts["2026-01"])
	}
}

func TestLoadClusterCountsFromCreatedDate(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "clusters-created.csv")

	csvContent := `created_date,cluster_id
2026-01-05,cluster1
2026-01-15,cluster2
2026-02-03,cluster3`

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	counts, err := LoadClusterCounts(csvPath)
	if err != nil {
		t.Fatalf("LoadClusterCounts() error = %v", err)
	}

	if counts["2026-01"] != 2 {
		t.Errorf("counts[2026-01] = %d, want 2", counts["2026-01"])
	}
	if counts["2026-02"] != 1 {
		t.Errorf("counts[2026-02] = %d, want 1", counts["2026-02"])
	}
}

func TestLoadEnvClusterCounts(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "env-clusters.csv")

	csvContent := `month,prod_count,stg_count,int_count,total_count,prod_source
2026-01,100,50,30,180,historical
2026-02,105,52,31,188,live_ocm`

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	counts, err := LoadEnvClusterCounts(csvPath)
	if err != nil {
		t.Fatalf("LoadEnvClusterCounts() error = %v", err)
	}

	// Should select live_ocm row (last one with that source)
	if prod := counts["prod"].(int); prod != 105 {
		t.Errorf("prod = %d, want 105", prod)
	}
	if stg := counts["stg"].(int); stg != 52 {
		t.Errorf("stg = %d, want 52", stg)
	}
	if total := counts["total"].(int); total != 188 {
		t.Errorf("total = %d, want 188", total)
	}
	if month := counts["month"].(string); month != "2026-02" {
		t.Errorf("month = %q, want %q", month, "2026-02")
	}
}

func TestLoadEnvClusterCountsNoLiveOCM(t *testing.T) {
	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "env-clusters-historical.csv")

	csvContent := `month,prod_count,stg_count,total_count
2026-01,100,50,150
2026-02,105,52,157`

	if err := os.WriteFile(csvPath, []byte(csvContent), 0644); err != nil {
		t.Fatal(err)
	}

	counts, err := LoadEnvClusterCounts(csvPath)
	if err != nil {
		t.Fatalf("LoadEnvClusterCounts() error = %v", err)
	}

	// Should select last row
	if prod := counts["prod"].(int); prod != 105 {
		t.Errorf("prod = %d, want 105", prod)
	}
}

func TestExtractMonth(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"2026-01-15", "2026-01"},
		{"2026-12-31", "2026-12"},
		{"2026-05-01T12:00:00", "2026-05"},
		{"2026-03-22T18:30:45Z", "2026-03"},
		{"invalid", ""},
		{"2026", ""},
	}

	for _, tt := range tests {
		result := extractMonth(tt.input)
		if result != tt.expected {
			t.Errorf("extractMonth(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestSafeInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"123", 123},
		{"0", 0},
		{"  456  ", 456},
		{"", 0},
		{"invalid", 0},
		{"12.5", 0},
	}

	for _, tt := range tests {
		result := safeInt(tt.input)
		if result != tt.expected {
			t.Errorf("safeInt(%q) = %d, want %d", tt.input, result, tt.expected)
		}
	}
}
