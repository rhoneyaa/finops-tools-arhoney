package configstore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegisterAWSLinkedAccount(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := RegisterAWSAccount(path, "123456789012", "rh-control"); err != nil {
		t.Fatal(err)
	}
	if err := RegisterAWSLinkedAccount(path, "111111111111", "osd-tenant-1", "rh-control",
		"OrganizationAccountAccessRole"); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	linked, ok := cfg.LinkedAccountForAlias("osd-tenant-1")
	if !ok || linked.AccountID != "111111111111" || linked.PayerAlias != "rh-control" {
		t.Fatalf("linked: %+v %v", linked, ok)
	}
	if linked.Role != "OrganizationAccountAccessRole" {
		t.Fatalf("role = %q", linked.Role)
	}
	arn, err := cfg.LinkedRoleARNForAccount("111111111111", linked.Role)
	if err != nil {
		t.Fatal(err)
	}
	if arn != "arn:aws:iam::111111111111:role/OrganizationAccountAccessRole" {
		t.Fatalf("arn = %q", arn)
	}
}

func TestSetLinkedAccountRequiresPayer(t *testing.T) {
	cfg := Default()
	_, err := cfg.SetLinkedAccount("linked-1", LinkedAccount{
		AccountID: "111111111111", PayerAlias: "missing", Role: "OrganizationAccountAccessRole",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSetLinkedAccountRejectsLinkedPayerAlias(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := RegisterAWSAccount(path, "123456789012", "rh-control"); err != nil {
		t.Fatal(err)
	}
	if err := RegisterAWSLinkedAccount(path, "111111111111", "osd-tenant-1", "rh-control",
		"OrganizationAccountAccessRole"); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	_, err = cfg.SetLinkedAccount("other", LinkedAccount{
		AccountID: "222222222222", PayerAlias: "osd-tenant-1", Role: "OrganizationAccountAccessRole",
	})
	if err == nil {
		t.Fatal("expected error when payer alias is a linked account")
	}
}

func TestMigrateLinkedAccountRoleARN(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	legacy := []byte(`aws:
  account_aliases:
    rh-control: "123456789012"
    osd-tenant-1:
      account_id: "111111111111"
      payer_alias: rh-control
      role_arn: "arn:aws:iam::111111111111:role/OrganizationAccountAccessRole"
`)
	if err := os.WriteFile(path, legacy, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	linked, ok := cfg.LinkedAccountForAlias("osd-tenant-1")
	if !ok || linked.Role != "OrganizationAccountAccessRole" || linked.RoleARN != "" {
		t.Fatalf("linked: %+v", linked)
	}
}

func TestMigrateLegacyLinkedAccountsSection(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	legacy := []byte(`aws:
  account_aliases:
    rh-control: "123456789012"
  linked_accounts:
    osd-tenant-1:
      account_id: "111111111111"
      payer_alias: rh-control
      role: OrganizationAccountAccessRole
`)
	if err := os.WriteFile(path, legacy, 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	linked, ok := cfg.LinkedAccountForAlias("osd-tenant-1")
	if !ok || linked.AccountID != "111111111111" || linked.PayerAlias != "rh-control" {
		t.Fatalf("linked: %+v %v", linked, ok)
	}
}

func TestAWSLinkedRoleNameDefault(t *testing.T) {
	if got := Default().AWSLinkedRoleName(); got != "OrganizationAccountAccessRole" {
		t.Fatalf("got %q", got)
	}
}
