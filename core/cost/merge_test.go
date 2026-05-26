// merge_test.go tests merging multiple CostResult values and breakdown aggregation.
package cost

import "testing"

func TestMergeResultsCombinesTotalsAndServices(t *testing.T) {
	results := []CostResult{
		{
			Provider: ProviderAWS, AccountName: "a", AccountID: "1", Currency: "USD",
			StartDate: "2026-04-25", EndDate: "2026-05-24", Amount: 100,
			Breakdown: []CostBreakdownItem{{Service: "Amazon EC2", Amount: 80}, {Service: "Amazon S3", Amount: 20}},
		},
		{
			Provider: ProviderAWS, AccountName: "b", AccountID: "2", Currency: "USD",
			StartDate: "2026-04-25", EndDate: "2026-05-24", Amount: 50,
			Breakdown: []CostBreakdownItem{{Service: "Amazon EC2", Amount: 30}, {Service: "Amazon RDS", Amount: 20}},
		},
	}

	merged, err := MergeResults(results)
	if err != nil {
		t.Fatal(err)
	}
	if merged.Amount != 150 {
		t.Errorf("Amount = %v, want 150", merged.Amount)
	}
	if merged.AccountName != "a, b" {
		t.Errorf("AccountName = %q", merged.AccountName)
	}
	if len(merged.Breakdown) != 3 {
		t.Fatalf("Breakdown = %+v", merged.Breakdown)
	}
	if merged.Breakdown[0].Service != "Amazon EC2" || merged.Breakdown[0].Amount != 110 {
		t.Errorf("EC2 = %+v", merged.Breakdown[0])
	}
}

func TestMergeResultsCombinesLinkedAccounts(t *testing.T) {
	results := []CostResult{
		{
			Provider: ProviderAWS, AccountName: "payer-a", Currency: "USD",
			StartDate: "2026-04-25", EndDate: "2026-05-24", Amount: 60, SplitBy: SplitByAccount,
			Breakdown: []CostBreakdownItem{{Account: "111111111111", Amount: 60}},
		},
		{
			Provider: ProviderAWS, AccountName: "payer-b", Currency: "USD",
			StartDate: "2026-04-25", EndDate: "2026-05-24", Amount: 40, SplitBy: SplitByAccount,
			Breakdown: []CostBreakdownItem{{Account: "111111111111", Amount: 10}, {Account: "222222222222", Amount: 30}},
		},
	}
	merged, err := MergeResults(results)
	if err != nil {
		t.Fatal(err)
	}
	if merged.Amount != 100 {
		t.Errorf("Amount = %v", merged.Amount)
	}
	if len(merged.Breakdown) != 2 || merged.Breakdown[0].Account != "111111111111" || merged.Breakdown[0].Amount != 70 {
		t.Fatalf("Breakdown = %+v", merged.Breakdown)
	}
}

func TestMergeResultsRejectsMixedCurrency(t *testing.T) {
	_, err := MergeResults([]CostResult{
		{Currency: "USD", StartDate: "2026-04-25", EndDate: "2026-05-24", Amount: 1},
		{Currency: "EUR", StartDate: "2026-04-25", EndDate: "2026-05-24", Amount: 1},
	})
	if err == nil {
		t.Fatal("expected error")
	}
}
