package rhsaml

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestParseSingleAccountFromAssertion(t *testing.T) {
	raw := `<Response xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">
<saml:Assertion>
  <saml:AttributeStatement>
    <saml:Attribute Name="https://aws.amazon.com/SAML/Attributes/Role">
      <saml:AttributeValue>arn:aws:iam::123456789012:role/AdministratorAccess,arn:aws:iam::123456789012:saml-provider/RedHatInternal</saml:AttributeValue>
    </saml:Attribute>
  </saml:AttributeStatement>
</saml:Assertion>
</Response>`
	assertion := base64.StdEncoding.EncodeToString([]byte(raw))
	got, ok, err := parseSingleAccountFromAssertion(assertion)
	if err != nil {
		t.Fatalf("parseSingleAccountFromAssertion: %v", err)
	}
	if !ok {
		t.Fatalf("expected single account")
	}
	if got.UID != "123456789012" || got.Name != "123456789012" {
		t.Fatalf("unexpected account identity: %+v", got)
	}
	if got.RoleName != "AdministratorAccess" || got.RoleARN != "arn:aws:iam::123456789012:role/AdministratorAccess" {
		t.Fatalf("unexpected role details: %+v", got)
	}
}

func TestParseAccountsFromHTML(t *testing.T) {
	page := `
<div class="saml-account">
  <div class="saml-account-name">Account: rh-control (123456789012)</div>
  <div class="saml-role">
    <label for="arn:aws:iam::123456789012:role/ReadOnlyAccess">ReadOnlyAccess</label>
  </div>
</div>`
	accounts, err := parseAccountsFromHTML(page)
	if err != nil {
		t.Fatalf("parseAccountsFromHTML: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("accounts len = %d", len(accounts))
	}
	if accounts[0].Name != "rh-control" || accounts[0].UID != "123456789012" {
		t.Fatalf("unexpected account: %+v", accounts[0])
	}
}

func TestSelectAccountMatchesUIDAndRole(t *testing.T) {
	accounts := []awsAccount{
		{Name: "rh-control", UID: "123456789012", RoleName: "ReadOnlyAccess", RoleARN: "arn:aws:iam::123456789012:role/ReadOnlyAccess"},
		{Name: "rh-control", UID: "123456789012", RoleName: "AdministratorAccess", RoleARN: "arn:aws:iam::123456789012:role/AdministratorAccess"},
	}
	got, err := selectAccount(accounts, []string{"123456789012"}, "AdministratorAccess")
	if err != nil {
		t.Fatalf("selectAccount: %v", err)
	}
	if got.RoleName != "AdministratorAccess" {
		t.Fatalf("selected role = %q", got.RoleName)
	}
}

func TestSelectAccountShowsAvailableEntries(t *testing.T) {
	_, err := selectAccount([]awsAccount{
		{Name: "rh-control", UID: "123456789012", RoleName: "ReadOnlyAccess"},
		{Name: "osd-staging", UID: "210987654321", RoleName: "ReadOnlyAccess"},
	}, []string{"missing"}, "")
	if err == nil || !strings.Contains(err.Error(), "available") {
		t.Fatalf("expected detailed not found error, got %v", err)
	}
}
