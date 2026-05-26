// profile_test.go tests SanitizeProfileName.
package aws

import "testing"

func TestSanitizeProfileName(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"rh-control", "rh-control"},
		{"RH-Control", "rh-control"},
		{"osd staging 1", "osd-staging-1"},
		{"  spaced  ", "spaced"},
		{"a@b#c", "a-b-c"},
	}
	for _, tc := range tests {
		if got := SanitizeProfileName(tc.in); got != tc.want {
			t.Errorf("SanitizeProfileName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
