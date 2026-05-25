package cmd

import (
	"github.com/spf13/cobra"
)

var costCmd = &cobra.Command{
	Use:   "cost",
	Short: "Cloud cost queries",
	Long:  "Fetch cost and usage summaries from cloud providers (AWS today; GCP planned).",
}

func init() {
	costCmd.GroupID = "core"
	rootCmd.AddCommand(costCmd)
}
