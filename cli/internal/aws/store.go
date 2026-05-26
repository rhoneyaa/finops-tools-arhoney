// store.go persists fresh credentials to the credentials file and returns session metadata (Result).
package aws

import (
	"context"
	"errors"
	"fmt"
)

// Result is returned when credentials are resolved or stored.
type Result struct {
	AccountID string
	ARN       string
	UserID    string
	Profile   string
	Refreshed bool
}

// StoreOptions configures StoreCredentials.
type StoreOptions struct {
	AccountName     string
	ProfileNames    []string
	CredentialsPath string
	Validator       CredentialValidator
}

// StoreCredentials writes a session to the account profile and validates it with STS.
func StoreCredentials(ctx context.Context, opts StoreOptions, sess ProfileSession) (Result, error) {
	if opts.AccountName == "" {
		return Result{}, errors.New("account name is required")
	}

	path := opts.CredentialsPath
	if path == "" {
		var err error
		path, err = DefaultCredentialsPath()
		if err != nil {
			return Result{}, err
		}
	}

	writeProfile := profileNamesForAccount(opts.AccountName, opts.ProfileNames)[0]
	validator := opts.Validator
	if validator == nil {
		validator = STSValidator{}
	}

	if err := WriteProfile(path, writeProfile, sess); err != nil {
		return Result{}, fmt.Errorf("write credentials profile: %w", err)
	}

	id, err := validator.Validate(ctx, sess)
	if err != nil {
		return Result{}, fmt.Errorf("verify credentials: %w", err)
	}

	return Result{
		AccountID: id.AccountID,
		ARN:       id.ARN,
		UserID:    id.UserID,
		Profile:   writeProfile,
		Refreshed: true,
	}, nil
}
