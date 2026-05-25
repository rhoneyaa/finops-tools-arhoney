package cmd

import (
	"github.com/spf13/cobra"
)

var configurationCmd = &cobra.Command{
	Use:   "configuration",
	Short: "Manage the finops configuration file and defaults",
}

func init() {
	configurationCmd.GroupID = "setup"
	rootCmd.AddCommand(configurationCmd)
}
