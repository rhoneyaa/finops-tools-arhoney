// aws_ensure.go builds awsauth.EnsureOptions from Cobra flags and TTY detection for account commands.
package cmd

import (
	"os"

	"github.com/mattn/go-isatty"
	"github.com/openshift-online/finops-tools/cli/internal/awsauth"
	"github.com/openshift-online/finops-tools/cli/internal/awslogin"
	"github.com/spf13/cobra"
)

type awsEnsureConfig struct {
	configPath      string
	authMethodFlag  string
	force           bool
	credentialsFile string
}

func newAWSEnsureOptions(cmd *cobra.Command, cfg awsEnsureConfig) (awsauth.EnsureOptions, error) {
	method, err := resolveAuthMethod(cmd, cfg.configPath, cfg.authMethodFlag)
	if err != nil {
		return awsauth.EnsureOptions{}, err
	}
	ensureOpts := awsauth.EnsureOptions{
		Force:           cfg.force,
		Method:          method,
		CredentialsPath: cfg.credentialsFile,
	}
	switch method {
	case awsauth.MethodSAML:
		ensureOpts.Provider = awslogin.SAMLLoginRunner{}
	case awsauth.MethodProfile:
		if isatty.IsTerminal(os.Stdin.Fd()) {
			ensureOpts.Provider = awslogin.InteractiveProfileRunner{
				Out: cmd.OutOrStdout(),
			}
		}
	}
	return ensureOpts, nil
}
