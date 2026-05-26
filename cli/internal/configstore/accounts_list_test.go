// accounts_list_test.go tests listing registered AWS accounts from config.
package configstore

import (
	"testing"
)

func TestListAWSAccounts(t *testing.T) {
	cfg := Default()
	var err error
	cfg, err = cfg.SetAWSAlias("rh-control", "123456789012")
	if err != nil {
		t.Fatal(err)
	}
	cfg, err = cfg.SetLinkedAccount("osd-tenant-1", LinkedAccount{
		AccountID:  "111111111111",
		PayerAlias: "rh-control",
		Role:       "OrganizationAccountAccessRole",
	})
	if err != nil {
		t.Fatal(err)
	}

	entries := cfg.ListAWSAccounts()
	if len(entries) != 2 {
		t.Fatalf("len = %d, want 2", len(entries))
	}
	if entries[0].Alias != "osd-tenant-1" || entries[0].Kind != "linked" {
		t.Fatalf("first entry = %+v", entries[0])
	}
	if entries[0].PayerAlias != "rh-control" || entries[0].Role != "OrganizationAccountAccessRole" {
		t.Fatalf("linked fields = %+v", entries[0])
	}
	if entries[1].Alias != "rh-control" || entries[1].Kind != "payer" {
		t.Fatalf("second entry = %+v", entries[1])
	}
}

func TestListAWSAccountsEmpty(t *testing.T) {
	if entries := Default().ListAWSAccounts(); entries != nil {
		t.Fatalf("got %v, want nil", entries)
	}
}
