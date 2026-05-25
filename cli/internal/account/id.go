package account

import (
	"fmt"
	"regexp"
	"strings"
)

var awsAccountIDPattern = regexp.MustCompile(`^\d{12}$`)

// ValidateAWSAccountID reports whether s is a 12-digit AWS account ID.
func ValidateAWSAccountID(s string) error {
	s = strings.TrimSpace(s)
	if !awsAccountIDPattern.MatchString(s) {
		return fmt.Errorf("invalid AWS account ID %q (expected 12 digits)", s)
	}
	return nil
}

// ParseCommaSeparated splits a comma-separated value into unique trimmed entries.
func ParseCommaSeparated(s string) ([]string, error) {
	var out []string
	seen := make(map[string]struct{})
	for _, part := range strings.Split(s, ",") {
		v := strings.TrimSpace(part)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out, nil
}
