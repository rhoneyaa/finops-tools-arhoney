package cmd

import (
	"github.com/spf13/cobra"
)

var accountCmd = &cobra.Command{
	Use:   "account",
	Short: "Manage cloud payer accounts and aliases",
}

func init() {
	accountCmd.GroupID = "setup"
	rootCmd.AddCommand(accountCmd)
}
