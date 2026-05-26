// Package awsauth orchestrates AWS credential ensure: SAML login, profile prompt, or reuse of stored profiles.
package awsauth

import (
	"fmt"
	"strings"
)

// Method selects how the CLI obtains AWS credentials when stored credentials are unavailable.
type Method string

const (
	// MethodSAML runs rh-aws-saml-login when stored credentials are missing or invalid.
	MethodSAML Method = "saml"
	// MethodProfile uses existing ~/.aws profiles; may prompt for keys when interactive.
	MethodProfile Method = "profile"
)

// ParseMethod parses an --auth-method flag value.
func ParseMethod(s string) (Method, error) {
	switch Method(strings.ToLower(strings.TrimSpace(s))) {
	case "", MethodSAML:
		return MethodSAML, nil
	case MethodProfile:
		return MethodProfile, nil
	default:
		return "", fmt.Errorf("unknown auth method %q (use saml or profile)", s)
	}
}
