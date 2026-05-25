package awsrole

import "testing"

func TestLinkedRoleARN(t *testing.T) {
	arn, err := LinkedRoleARN("111111111111", "OrganizationAccountAccessRole")
	if err != nil {
		t.Fatal(err)
	}
	want := "arn:aws:iam::111111111111:role/OrganizationAccountAccessRole"
	if arn != want {
		t.Fatalf("got %q want %q", arn, want)
	}
}

func TestLinkedRoleARNPassthrough(t *testing.T) {
	full := "arn:aws:iam::111111111111:role/CustomRole"
	got, err := LinkedRoleARN("111111111111", full)
	if err != nil || got != full {
		t.Fatalf("got %q err=%v", got, err)
	}
}

func TestNameFromARN(t *testing.T) {
	if got := NameFromARN("arn:aws:iam::111:role/OrganizationAccountAccessRole"); got != "OrganizationAccountAccessRole" {
		t.Fatalf("got %q", got)
	}
}

func TestLinkedRoleARNRejectsInvalidName(t *testing.T) {
	_, err := LinkedRoleARN("111111111111", "invalid/role")
	if err == nil {
		t.Fatal("expected error")
	}
}
