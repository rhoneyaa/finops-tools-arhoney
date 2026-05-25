package awsauth

import "testing"

func TestParseMethod(t *testing.T) {
	tests := []struct {
		in   string
		want Method
	}{
		{"", MethodSAML},
		{"saml", MethodSAML},
		{"SAML", MethodSAML},
		{"profile", MethodProfile},
	}
	for _, tc := range tests {
		got, err := ParseMethod(tc.in)
		if err != nil {
			t.Fatalf("ParseMethod(%q): %v", tc.in, err)
		}
		if got != tc.want {
			t.Fatalf("ParseMethod(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
	if _, err := ParseMethod("kerberos"); err == nil {
		t.Fatal("expected error")
	}
}
