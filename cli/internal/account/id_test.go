package account

import "testing"

func TestValidateAWSAccountID(t *testing.T) {
	if err := ValidateAWSAccountID("123456789012"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateAWSAccountID("rh-control"); err == nil {
		t.Fatal("expected error for non-numeric account")
	}
}
