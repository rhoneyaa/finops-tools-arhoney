// check.go verifies that stored or resolved credentials are valid for a given account profile.
package aws

import (
	"context"
	"fmt"
)

// CheckOptions configures CheckCredentials.
type CheckOptions struct {
	AccountName     string
	ProfileNames    []string
	CredentialsPath string
	Validator       CredentialValidator
}

// CheckCredentials validates stored profiles with STS without obtaining new credentials.
func CheckCredentials(ctx context.Context, opts CheckOptions) (Result, error) {
	res, status, err := ResolveCredentials(ctx, ResolveOptions{
		AccountName:     opts.AccountName,
		ProfileNames:    opts.ProfileNames,
		CredentialsPath: opts.CredentialsPath,
		Validator:       opts.Validator,
	})
	if err != nil {
		return Result{}, err
	}
	profiles := profileNamesForAccount(opts.AccountName, opts.ProfileNames)
	switch status {
	case CredentialsValid:
		return res, nil
	case CredentialsInvalid:
		return Result{}, fmt.Errorf("%w: profiles %v", ErrCredentialsInvalid, profiles)
	default:
		return Result{}, fmt.Errorf("%w: profiles %v", ErrCredentialsNotFound, profiles)
	}
}
