package configstore

// RegisterAWSAccount ensures the config file exists and records alias → accountID.
// When alias is empty, the account ID is used as the alias key.
func RegisterAWSAccount(path, accountID, alias string) error {
	cfg, err := Ensure(path)
	if err != nil {
		return err
	}
	if alias == "" {
		alias = accountID
	}
	cfg, err = cfg.SetAWSAlias(alias, accountID)
	if err != nil {
		return err
	}
	return Save(path, cfg)
}
