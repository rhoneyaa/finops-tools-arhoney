// provider_test.go tests provider/split-by parsing and the Fetch entry point.
package cost

import "testing"

func TestFetchRequiresAccount(t *testing.T) {
	_, err := Fetch(t.Context(), CostQuery{Provider: ProviderAWS})
	if err == nil || err.Error() != "at least one account is required" {
		t.Fatalf("got %v", err)
	}
}

func TestMergeResultsSinglePassthrough(t *testing.T) {
	in := CostResult{Amount: 42, Currency: "USD", StartDate: "2026-04-25", EndDate: "2026-05-24"}
	out, err := MergeResults([]CostResult{in})
	if err != nil || out.Amount != 42 {
		t.Fatalf("got %v %v", out, err)
	}
}

func TestParseSplitBy(t *testing.T) {
	s, err := ParseSplitBy("service")
	if err != nil || s != SplitByService {
		t.Fatalf("got %v %v", s, err)
	}
	s, err = ParseSplitBy("")
	if err != nil || s != SplitByNone {
		t.Fatalf("got %v %v", s, err)
	}
	s, err = ParseSplitBy("account")
	if err != nil || s != SplitByAccount {
		t.Fatalf("got %v %v", s, err)
	}
	_, err = ParseSplitBy("region")
	if err == nil {
		t.Fatal("expected error")
	}
}
