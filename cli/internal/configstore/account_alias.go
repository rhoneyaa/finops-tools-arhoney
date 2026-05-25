package configstore

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// AWSAccountAlias is a payer account ID (YAML scalar) or linked account metadata (YAML mapping).
type AWSAccountAlias struct {
	accountID  string
	PayerAlias string
	Role       string
	RoleARN    string // deprecated; loaded for backward compatibility only
}

// AccountID returns the 12-digit AWS account ID for this alias.
func (a AWSAccountAlias) AccountID() string {
	return strings.TrimSpace(a.accountID)
}

// IsLinked reports whether this alias refers to a member account accessed via role assumption.
func (a AWSAccountAlias) IsLinked() bool {
	return strings.TrimSpace(a.PayerAlias) != ""
}

// LinkedAccount returns linked account metadata when this alias is linked.
func (a AWSAccountAlias) LinkedAccount() (LinkedAccount, bool) {
	if !a.IsLinked() {
		return LinkedAccount{}, false
	}
	return LinkedAccount{
		AccountID:  a.accountID,
		PayerAlias: a.PayerAlias,
		Role:       a.Role,
		RoleARN:    a.RoleARN,
	}, true
}

func awsAccountAliasFromPayer(accountID string) AWSAccountAlias {
	return AWSAccountAlias{accountID: strings.TrimSpace(accountID)}
}

func awsAccountAliasFromLinked(linked LinkedAccount) AWSAccountAlias {
	return AWSAccountAlias{
		accountID:  linked.AccountID,
		PayerAlias: linked.PayerAlias,
		Role:       linked.Role,
		RoleARN:    linked.RoleARN,
	}
}

func (a *AWSAccountAlias) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		a.accountID = strings.TrimSpace(value.Value)
		a.PayerAlias = ""
		a.Role = ""
		a.RoleARN = ""
		return nil
	case yaml.MappingNode:
		var linked LinkedAccount
		if err := value.Decode(&linked); err != nil {
			return err
		}
		*a = awsAccountAliasFromLinked(linked)
		return nil
	default:
		return fmt.Errorf("account alias must be a string or mapping")
	}
}

func (a AWSAccountAlias) MarshalYAML() (interface{}, error) {
	if a.IsLinked() {
		out := map[string]string{
			"account_id":  a.accountID,
			"payer_alias": a.PayerAlias,
		}
		if strings.TrimSpace(a.Role) != "" {
			out["role"] = a.Role
		}
		return out, nil
	}
	return a.accountID, nil
}
