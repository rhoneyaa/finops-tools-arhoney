// config_default.go registers the "finops config default" subcommand group.
package cmd

import (
	"github.com/spf13/cobra"
)

var configDefaultCmd = &cobra.Command{
	Use:   "default",
	Short: "Manage finops configuration defaults",
	Long:  "Set or read default values stored in the finops config file (fully qualified names under defaults).",
}

func init() {
	configCmd.AddCommand(configDefaultCmd)
}
