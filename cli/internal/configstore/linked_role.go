// linked_role.go resolves the default linked-account IAM role name and ARN from config.
package configstore

import (
	"github.com/openshift-online/finops-tools/cli/internal/awsrole"
)

// AWSLinkedRoleName returns the configured default linked-account role name.
func (f File) AWSLinkedRoleName() string {
	if v, ok := f.Default(DefaultFQNAWSLinkedRole); ok && v != "" {
		return v
	}
	return awsrole.DefaultLinkedRoleName
}

// LinkedRoleARNForAccount resolves a role name to an IAM role ARN in the given account.
func (f File) LinkedRoleARNForAccount(accountID, roleName string) (string, error) {
	if roleName == "" {
		roleName = f.AWSLinkedRoleName()
	}
	return awsrole.LinkedRoleARN(accountID, roleName)
}
