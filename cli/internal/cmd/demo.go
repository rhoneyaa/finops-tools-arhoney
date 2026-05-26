// demo.go registers the "finops demo" noun command for development and smoke-test utilities.
package cmd

import (
	"github.com/spf13/cobra"
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Demo commands for development and smoke tests",
}

func init() {
	demoCmd.GroupID = "setup"
	rootCmd.AddCommand(demoCmd)
}
