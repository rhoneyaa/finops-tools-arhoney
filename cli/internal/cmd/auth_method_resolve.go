// auth_method_resolve.go picks the AWS auth method from --auth-method or config defaults.
package cmd

import (
	"github.com/openshift-online/finops-tools/cli/internal/awsauth"
	"github.com/openshift-online/finops-tools/cli/internal/configstore"
	"github.com/spf13/cobra"
)

// resolveAuthMethod uses --auth-method when explicitly set; otherwise the config default, then saml.
func resolveAuthMethod(cmd *cobra.Command, configPath, flagValue string) (awsauth.Method, error) {
	if cmd.Flags().Changed("auth-method") {
		return awsauth.ParseMethod(flagValue)
	}
	path, err := configstore.ResolvePath(configPath)
	if err != nil {
		return "", err
	}
	cfg, err := configstore.Load(path)
	if err != nil {
		return "", err
	}
	if v, ok := cfg.Default(configstore.DefaultFQNAWSAuthMethod); ok {
		return awsauth.ParseMethod(v)
	}
	return awsauth.MethodSAML, nil
}
