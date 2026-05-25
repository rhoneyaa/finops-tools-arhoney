package awsrole

import (
	"fmt"
	"strings"
)

// DefaultLinkedRoleName is the IAM role name assumed into linked accounts when none is configured.
const DefaultLinkedRoleName = "OrganizationAccountAccessRole"

// LinkedRoleARN builds arn:aws:iam::<account-id>:role/<role-name>.
// If roleName is already an IAM role ARN, it is returned unchanged.
func LinkedRoleARN(accountID, roleName string) (string, error) {
	accountID = strings.TrimSpace(accountID)
	roleName = strings.TrimSpace(roleName)
	if accountID == "" {
		return "", fmt.Errorf("account ID is required")
	}
	if roleName == "" {
		return "", fmt.Errorf("role name is required")
	}
	if strings.HasPrefix(roleName, "arn:") {
		return roleName, nil
	}
	if err := ValidateName(roleName); err != nil {
		return "", err
	}
	return fmt.Sprintf("arn:aws:iam::%s:role/%s", accountID, roleName), nil
}

// NameFromARN extracts the role name from a role ARN, or returns the input if not an ARN.
func NameFromARN(arn string) string {
	arn = strings.TrimSpace(arn)
	const suffix = ":role/"
	if idx := strings.LastIndex(arn, suffix); idx >= 0 {
		return arn[idx+len(suffix):]
	}
	return arn
}

// ValidateName reports whether name is a valid IAM role name (not a full ARN).
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if strings.HasPrefix(name, "arn:") {
		return fmt.Errorf("pass the role name only (e.g. %s), not a full ARN", DefaultLinkedRoleName)
	}
	if name == "" {
		return fmt.Errorf("role name is required")
	}
	if strings.Contains(name, "/") || strings.Contains(name, ":") || strings.Contains(name, " ") {
		return fmt.Errorf("invalid role name %q (pass the role name only, e.g. %s)", name, DefaultLinkedRoleName)
	}
	return nil
}
