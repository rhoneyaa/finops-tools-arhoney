// Command finops is the entry point for the FinOps CLI binary.
// It delegates to cli/internal/cmd for Cobra command wiring and exits non-zero on error.
package main

import (
	"os"

	"github.com/openshift-online/finops-tools/cli/internal/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
