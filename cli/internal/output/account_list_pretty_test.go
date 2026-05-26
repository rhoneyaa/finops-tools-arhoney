// account_list_pretty_test.go tests account list tables and styling.
package output

import (
	"strings"
	"testing"
)

func TestWriteAWSAccountListPretty(t *testing.T) {
	entries := []AccountListRow{
		{Alias: "rh-control", AccountID: "123456789012", Kind: "payer"},
		{
			Alias: "osd-tenant-1", AccountID: "111111111111", Kind: "linked",
			PayerAlias: "rh-control", Role: "OrganizationAccountAccessRole",
		},
	}

	var buf strings.Builder
	if err := WriteAWSAccountList(&buf, entries); err != nil {
		t.Fatal(err)
	}
	out := stripANSI(buf.String())
	if strings.Contains(out, "\033") {
		t.Fatal("buffer output should not contain escape codes")
	}
	for _, want := range []string{
		"Registered AWS accounts:",
		"2 (1 payer, 1 linked)",
		"AWS accounts",
		"ALIAS",
		"rh-control",
		"123456789012",
		"payer",
		"osd-tenant-1",
		"linked",
		"OrganizationAccountAccessRole",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestWriteAWSAccountListEmpty(t *testing.T) {
	var buf strings.Builder
	if err := WriteAWSAccountList(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), "No AWS accounts registered.") {
		t.Fatalf("got %q", buf.String())
	}
}

func TestFormatAccountKindColored(t *testing.T) {
	s := styler{enabled: true}
	if !strings.Contains(formatAccountKind(s, "payer"), "\033") {
		t.Fatal("expected ANSI for payer")
	}
	if !strings.Contains(formatAccountKind(s, "linked"), "\033") {
		t.Fatal("expected ANSI for linked")
	}
}
