// config_default_get.go implements "finops config default get" to print a stored default value.
package cmd

import (
	"fmt"

	"github.com/openshift-online/finops-tools/cli/internal/configstore"
	"github.com/spf13/cobra"
)

var (
	configDefaultGetName   string
	configDefaultGetConfig string
)

var configDefaultGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Print a configuration default",
	Long: `Print a default value from the finops config file.

Example:
  finops config default get --name aws.auth-method`,
	RunE: runConfigDefaultGet,
}

func init() {
	configDefaultCmd.AddCommand(configDefaultGetCmd)
	configDefaultGetCmd.Flags().StringVar(&configDefaultGetName, "name", "", "Default name (required)")
	configDefaultGetCmd.Flags().StringVar(&configDefaultGetConfig, "config", "", "Path to finops config file (default: OS-specific config dir)")
	_ = configDefaultGetCmd.MarkFlagRequired("name")
}

func runConfigDefaultGet(cmd *cobra.Command, _ []string) error {
	path, err := configstore.ResolvePath(configDefaultGetConfig)
	if err != nil {
		return err
	}
	cfg, err := configstore.Load(path)
	if err != nil {
		return err
	}
	v, ok := cfg.Default(configDefaultGetName)
	if !ok {
		return fmt.Errorf("default %q is not set in %s", configDefaultGetName, path)
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), v)
	return err
}
