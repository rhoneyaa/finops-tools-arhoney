// configuration_default_set.go implements "finops configuration default set" to persist a default value.
package cmd

import (
	"fmt"

	"github.com/openshift-online/finops-tools/cli/internal/configstore"
	"github.com/spf13/cobra"
)

var (
	configurationDefaultSetName  string
	configurationDefaultSetValue string
	configurationDefaultSetConfig string
)

var configurationDefaultSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a configuration default",
	Long: `Store a default value in the finops config file (fully qualified names).

Supported names:
  aws.auth-method — AWS authentication for account add (saml or profile)
  aws.linked_role — IAM role name for linked account add (default: OrganizationAccountAccessRole)
  gcp.auth-method — reserved for future GCP support

Example:
  finops configuration default set --name aws.auth-method --value profile
  finops configuration default set --name aws.linked_role --value OrganizationAccountAccessRole`,
	RunE: runConfigurationDefaultSet,
}

func init() {
	configurationDefaultCmd.AddCommand(configurationDefaultSetCmd)
	configurationDefaultSetCmd.Flags().StringVar(&configurationDefaultSetName, "name", "", "Default name (required)")
	configurationDefaultSetCmd.Flags().StringVar(&configurationDefaultSetValue, "value", "", "Default value (required)")
	configurationDefaultSetCmd.Flags().StringVar(&configurationDefaultSetConfig, "config", "", "Path to finops config file (default: OS-specific config dir)")
	_ = configurationDefaultSetCmd.MarkFlagRequired("name")
	_ = configurationDefaultSetCmd.MarkFlagRequired("value")
}

func runConfigurationDefaultSet(cmd *cobra.Command, _ []string) error {
	path, err := configstore.ResolvePath(configurationDefaultSetConfig)
	if err != nil {
		return err
	}
	if err := configstore.SetDefault(path, configurationDefaultSetName, configurationDefaultSetValue); err != nil {
		return err
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "Default %s set to %q in %s\n",
		configurationDefaultSetName, configurationDefaultSetValue, path)
	return err
}
