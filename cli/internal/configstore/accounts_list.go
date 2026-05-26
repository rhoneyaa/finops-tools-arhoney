// accounts_list.go lists registered AWS accounts from the finops config.
package configstore

import (
	"slices"
	"strings"
)

// AWSAccountListEntry is one registered AWS account alias.
type AWSAccountListEntry struct {
	Alias      string
	AccountID  string
	Kind       string // "payer" or "linked"
	PayerAlias string
	Role       string
}

// ListAWSAccounts returns registered AWS account aliases, sorted by alias.
func (f File) ListAWSAccounts() []AWSAccountListEntry {
	if len(f.AWS.AccountAliases) == 0 {
		return nil
	}
	out := make([]AWSAccountListEntry, 0, len(f.AWS.AccountAliases))
	for alias, entry := range f.AWS.AccountAliases {
		item := AWSAccountListEntry{
			Alias:     alias,
			AccountID: entry.AccountID(),
		}
		if entry.IsLinked() {
			item.Kind = "linked"
			item.PayerAlias = strings.TrimSpace(entry.PayerAlias)
			if linked, ok := entry.LinkedAccount(); ok {
				item.Role = linked.RoleName()
			}
		} else {
			item.Kind = "payer"
		}
		out = append(out, item)
	}
	slices.SortFunc(out, func(a, b AWSAccountListEntry) int {
		return strings.Compare(a.Alias, b.Alias)
	})
	return out
}
