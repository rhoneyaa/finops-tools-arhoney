package cmd

import (
	"fmt"

	"github.com/openshift-online/finops-tools/core"
	"github.com/spf13/cobra"
)

var demoHelloCmd = &cobra.Command{
	Use:   "hello",
	Short: "Print a greeting from the core module",
	RunE: func(cmd *cobra.Command, args []string) error {
		msg, err := core.Hello(cmd.Context())
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(cmd.OutOrStdout(), msg)
		return err
	},
}

func init() {
	demoCmd.AddCommand(demoHelloCmd)
}
