package cmd

import (
	"strings"

	"github.com/openshift-online/finops-tools/cli/internal/awsrole"
	"github.com/openshift-online/finops-tools/cli/internal/configstore"
	"github.com/spf13/cobra"
)

// resolveLinkedRoleName returns the IAM role name: --role when set, else config default, else built-in default.
func resolveLinkedRoleName(cmd *cobra.Command, configPath, flagValue string) (string, error) {
	if cmd.Flags().Changed("role") {
		name := strings.TrimSpace(flagValue)
		if name == "" {
			return "", nil
		}
		if strings.HasPrefix(name, "arn:") {
			return awsrole.NameFromARN(name), nil
		}
		if err := awsrole.ValidateName(name); err != nil {
			return "", err
		}
		return name, nil
	}
	path, err := configstore.ResolvePath(configPath)
	if err != nil {
		return "", err
	}
	cfg, err := configstore.Load(path)
	if err != nil {
		return "", err
	}
	return cfg.AWSLinkedRoleName(), nil
}

// resolveLinkedRoleARN builds the role ARN for a linked account.
func resolveLinkedRoleARN(cmd *cobra.Command, configPath, linkedAccountID, flagRole string) (string, error) {
	roleName, err := resolveLinkedRoleName(cmd, configPath, flagRole)
	if err != nil {
		return "", err
	}
	return awsrole.LinkedRoleARN(linkedAccountID, roleName)
}
