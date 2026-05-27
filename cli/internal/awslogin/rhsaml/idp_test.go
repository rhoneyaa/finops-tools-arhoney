package rhsaml

import "testing"

func TestParseSAMLForm(t *testing.T) {
	html := `<html><body><form action="/saml/login"><input type="hidden" value="abc123"></form></body></html>`
	form, err := parseSAMLForm("https://auth.redhat.com/realm", html)
	if err != nil {
		t.Fatalf("parseSAMLForm: %v", err)
	}
	if form.AWSURL != "https://auth.redhat.com/saml/login" {
		t.Fatalf("aws url = %q", form.AWSURL)
	}
	if form.SAMLAssertion != "abc123" {
		t.Fatalf("assertion = %q", form.SAMLAssertion)
	}
}
