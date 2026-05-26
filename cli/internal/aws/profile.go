// profile.go sanitizes account names into safe AWS credentials profile names.
package aws

import (
	"regexp"
	"strings"
)

var nonProfileChars = regexp.MustCompile(`[^a-zA-Z0-9_-]`)

// SanitizeProfileName turns an account name into a safe AWS credentials profile name.
func SanitizeProfileName(accountName string) string {
	s := nonProfileChars.ReplaceAllString(accountName, "-")
	return strings.ToLower(strings.Trim(s, "-"))
}
