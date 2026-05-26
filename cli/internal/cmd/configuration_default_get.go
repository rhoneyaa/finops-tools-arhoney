// configuration_default_get.go implements "finops configuration default get" to print a stored default value.
package cmd

import (
	"fmt"

	"github.com/openshift-online/finops-tools/cli/internal/configstore"
	"github.com/spf13/cobra"
)

var (
	configurationDefaultGetName   string
	configurationDefaultGetConfig string
)

var configurationDefaultGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Print a configuration default",
	Long: `Print a default value from the finops config file.

Example:
  finops configuration default get --name aws.auth-method`,
	RunE: runConfigurationDefaultGet,
}

func init() {
	configurationDefaultCmd.AddCommand(configurationDefaultGetCmd)
	configurationDefaultGetCmd.Flags().StringVar(&configurationDefaultGetName, "name", "", "Default name (required)")
	configurationDefaultGetCmd.Flags().StringVar(&configurationDefaultGetConfig, "config", "", "Path to finops config file (default: OS-specific config dir)")
	_ = configurationDefaultGetCmd.MarkFlagRequired("name")
}

func runConfigurationDefaultGet(cmd *cobra.Command, _ []string) error {
	path, err := configstore.ResolvePath(configurationDefaultGetConfig)
	if err != nil {
		return err
	}
	cfg, err := configstore.Load(path)
	if err != nil {
		return err
	}
	v, ok := cfg.Default(configurationDefaultGetName)
	if !ok {
		return fmt.Errorf("default %q is not set in %s", configurationDefaultGetName, path)
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), v)
	return err
}
