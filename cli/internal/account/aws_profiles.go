package account

import (
	"strings"

	awsconfig "github.com/openshift-online/finops-tools/cli/internal/aws"
)

// AWSProfileNames returns AWS profile names to try: alias first (when set), then account ID.
func AWSProfileNames(accountID, alias string, existing []string) []string {
	return awsProfileNames(accountID, alias, existing)
}

// awsProfileNames returns AWS profile names to try: alias first (when set), then account ID.
func awsProfileNames(accountID, alias string, existing []string) []string {
	if len(existing) > 0 {
		return existing
	}
	accountProfile := awsconfig.SanitizeProfileName(accountID)
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return []string{accountProfile}
	}
	aliasProfile := awsconfig.SanitizeProfileName(alias)
	if aliasProfile == accountProfile {
		return []string{accountProfile}
	}
	return []string{aliasProfile, accountProfile}
}
