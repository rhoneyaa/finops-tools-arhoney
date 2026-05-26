// config.go registers the "finops config" noun command for local finops config file management.
package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage the finops configuration file and defaults",
}

func init() {
	configCmd.GroupID = "setup"
	rootCmd.AddCommand(configCmd)
}
