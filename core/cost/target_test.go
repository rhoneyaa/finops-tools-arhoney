// target_test.go tests AccountTarget helpers (linked detection, credentials account ID).
package cost

import "testing"

func TestAccountTargetCredentialsAccountID(t *testing.T) {
	payer := AccountTarget{AccountID: "123456789012"}
	if got := payer.CredentialsAccountID(); got != "123456789012" {
		t.Fatalf("payer creds = %q", got)
	}
	if payer.IsLinked() {
		t.Fatal("payer should not be linked")
	}

	linked := AccountTarget{AccountID: "111111111111", PayerAccountID: "123456789012"}
	if got := linked.CredentialsAccountID(); got != "123456789012" {
		t.Fatalf("linked creds = %q", got)
	}
	if !linked.IsLinked() {
		t.Fatal("expected linked target")
	}
}

func TestFilterOverlappingTargets(t *testing.T) {
	independent := FilterOverlappingTargets([]AccountTarget{
		{AccountID: "123456789012"},
		{AccountID: "987654321098"},
	})
	if len(independent) != 2 {
		t.Fatalf("got %d targets, want 2", len(independent))
	}

	siblings := FilterOverlappingTargets([]AccountTarget{
		{AccountID: "111111111111", PayerAccountID: "123456789012"},
		{AccountID: "222222222222", PayerAccountID: "123456789012"},
	})
	if len(siblings) != 2 {
		t.Fatalf("got %d targets, want 2", len(siblings))
	}

	overlap := FilterOverlappingTargets([]AccountTarget{
		{AccountID: "123456789012"},
		{AccountID: "111111111111", PayerAccountID: "123456789012"},
	})
	if len(overlap) != 1 || overlap[0].AccountID != "123456789012" {
		t.Fatalf("got %+v, want payer only", overlap)
	}
}
