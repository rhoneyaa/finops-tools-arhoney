// defaults_test.go tests get/set of fully qualified configuration defaults and legacy migration.
package configstore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetDefaultAWSAuthMethod(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	if err := SetDefault(path, "aws.auth-method", "profile"); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := cfg.Default("aws.auth-method")
	if !ok || v != "profile" {
		t.Fatalf("got %q %v", v, ok)
	}
	if _, ok := cfg.Defaults[DefaultFQNAWSAuthMethod]; !ok {
		t.Fatalf("defaults map: %v", cfg.Defaults)
	}
}

func TestSetDefaultRequiresFQN(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cfg := Default()
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	if err := SetDefault(path, "auth-method", "profile"); err == nil {
		t.Fatal("expected error for non-FQN name")
	}
}

func TestSetDefaultRejectsUnknownFQN(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cfg := Default()
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	if err := SetDefault(path, "azure.auth-method", "profile"); err == nil {
		t.Fatal("expected error")
	}
}

func TestSetDefaultRejectsInvalidAuthMethod(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cfg := Default()
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	if err := SetDefault(path, "aws.auth-method", "kerberos"); err == nil {
		t.Fatal("expected error")
	}
}

func TestSetDefaultAWSLinkedRole(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := SetDefault(path, "aws.linked_role", "CustomRole"); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := cfg.AWSLinkedRoleName(); got != "CustomRole" {
		t.Fatalf("got %q", got)
	}
}

func TestSetDefaultRejectsLinkedRoleARN(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cfg := Default()
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	if err := SetDefault(path, "aws.linked_role", "arn:aws:iam::111:role/X"); err == nil {
		t.Fatal("expected error")
	}
}

func TestSetDefaultCostExcludeRecentDays(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := SetDefault(path, "cost.exclude_recent_days", "2"); err != nil {
		t.Fatal(err)
	}
	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := cfg.Default(DefaultFQNCostExcludeRecentDays)
	if !ok || v != "2" {
		t.Fatalf("got %q %v", v, ok)
	}
}

func TestSetDefaultCostDays(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := SetDefault(path, "cost.days", "7"); err != nil {
		t.Fatal(err)
	}
}

func TestSetDefaultRejectsInvalidCostDays(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	cfg := Default()
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	if err := SetDefault(path, "cost.days", "0"); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidateCostPeriodDefaultsConflict(t *testing.T) {
	cfg := Default()
	cfg.Defaults = map[string]string{
		DefaultFQNCostDays:   "30",
		DefaultFQNCostMonths: "3",
	}
	if err := cfg.ValidateCostPeriodDefaults(); err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestMigrateLegacyAWSDefaults(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	legacy := []byte(`aws:
  defaults:
    auth_method: profile
  account_aliases: {}
`)
	if err := os.WriteFile(path, legacy, 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	v, ok := cfg.Default(DefaultFQNAWSAuthMethod)
	if !ok || v != "profile" {
		t.Fatalf("got %q %v", v, ok)
	}
}
