package cmd

import (
	"github.com/openshift-online/finops-tools/cli/internal/configstore"
)

func registerAWSAccount(configPath, accountID, alias string) error {
	path, err := configstore.ResolvePath(configPath)
	if err != nil {
		return err
	}
	return configstore.RegisterAWSAccount(path, accountID, alias)
}

func registerAWSLinkedAccount(configPath, linkedAccountID, alias, payerAlias, roleARN string) error {
	path, err := configstore.ResolvePath(configPath)
	if err != nil {
		return err
	}
	return configstore.RegisterAWSLinkedAccount(path, linkedAccountID, alias, payerAlias, roleARN)
}

func payerAccountIDFromConfig(configPath, payerAlias string) (string, error) {
	path, err := configstore.ResolvePath(configPath)
	if err != nil {
		return "", err
	}
	cfg, err := configstore.Load(path)
	if err != nil {
		return "", err
	}
	id, ok := cfg.PayerAccountIDForAlias(payerAlias)
	if !ok {
		return "", errUnknownPayerAlias(payerAlias)
	}
	return id, nil
}
