// resolve.go parses --account flags and builds core/cost.AccountTarget slices from config aliases.
package configstore

import (
	"errors"
	"fmt"

	"github.com/openshift-online/finops-tools/cli/internal/account"
	"github.com/openshift-online/finops-tools/core/cost"
)

// ParseAWSAccountIDs parses --account (12-digit IDs only).
func ParseAWSAccountIDs(s string) ([]string, error) {
	ids, err := account.ParseCommaSeparated(s)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return nil, errors.New("at least one AWS account ID is required")
	}
	for _, id := range ids {
		if err := account.ValidateAWSAccountID(id); err != nil {
			return nil, err
		}
	}
	return ids, nil
}

// ParseAccountAliases parses --account-alias.
func ParseAccountAliases(s string) ([]string, error) {
	aliases, err := account.ParseCommaSeparated(s)
	if err != nil {
		return nil, err
	}
	if len(aliases) == 0 {
		return nil, errors.New("at least one account alias is required")
	}
	return aliases, nil
}

// ResolveCostTargets builds cost.AccountTarget values from account IDs and/or aliases.
func ResolveCostTargets(cfg File, accountIDs, aliases []string) ([]cost.AccountTarget, error) {
	if len(accountIDs) == 0 && len(aliases) == 0 {
		return nil, errors.New("at least one of --account or --account-alias is required")
	}

	var out []cost.AccountTarget
	seen := make(map[string]struct{})

	add := func(target cost.AccountTarget) error {
		accountID := target.AccountID
		if _, ok := seen[accountID]; ok {
			return nil
		}
		if !cfg.HasAWSAccount(accountID) {
			return fmt.Errorf("account %s is not registered (run: finops account add aws %s)", accountID, accountID)
		}
		seen[accountID] = struct{}{}
		out = append(out, target)
		return nil
	}

	for _, alias := range aliases {
		if linked, ok := cfg.LinkedAccountForAlias(alias); ok {
			payerID, ok := cfg.PayerAccountIDForAlias(linked.PayerAlias)
			if !ok {
				return nil, fmt.Errorf("unknown payer alias %q for linked account %q", linked.PayerAlias, alias)
			}
			if err := add(cost.AccountTarget{
				AccountID:      linked.AccountID,
				PayerAccountID: payerID,
				DisplayAlias:   alias,
			}); err != nil {
				return nil, err
			}
			continue
		}
		id, ok := cfg.PayerAccountIDForAlias(alias)
		if !ok {
			return nil, fmt.Errorf("unknown account alias %q (add with: finops account add aws <12-digit-id> --alias %s)", alias, alias)
		}
		if err := add(cost.AccountTarget{
			AccountID:    id,
			DisplayAlias: alias,
		}); err != nil {
			return nil, err
		}
	}

	for _, id := range accountIDs {
		displayAlias := cfg.AliasForAccountID(id)
		if displayAlias == id {
			displayAlias = ""
		}
		target := cost.AccountTarget{
			AccountID:    id,
			DisplayAlias: displayAlias,
		}
		if payerID, ok := cfg.PayerAccountIDForLinkedAccountID(id); ok {
			target.PayerAccountID = payerID
		}
		if err := add(target); err != nil {
			return nil, err
		}
	}

	return out, nil
}
