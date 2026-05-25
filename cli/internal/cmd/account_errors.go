package cmd

import "fmt"

func errUnknownPayerAlias(alias string) error {
	return fmt.Errorf("unknown payer alias %q (register payer first: finops account add aws <12-digit-id> --alias %s)", alias, alias)
}
