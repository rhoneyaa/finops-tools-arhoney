package aws

import (
	"context"
	"errors"
)

// ResolveStatus describes stored credential lookup for an account.
type ResolveStatus int

const (
	// CredentialsValid means a profile validated successfully with STS.
	CredentialsValid ResolveStatus = iota
	// CredentialsAbsent means no usable profile was found.
	CredentialsAbsent
	// CredentialsInvalid means a profile exists but STS rejected it.
	CredentialsInvalid
)

// ResolveOptions configures ResolveCredentials.
type ResolveOptions struct {
	AccountName     string
	ProfileNames    []string
	CredentialsPath string
	Validator       CredentialValidator
}

// ResolveCredentials returns validated credentials from ~/.aws without obtaining new ones.
func ResolveCredentials(ctx context.Context, opts ResolveOptions) (Result, ResolveStatus, error) {
	if opts.AccountName == "" {
		return Result{}, CredentialsAbsent, errors.New("account name is required")
	}

	path := opts.CredentialsPath
	if path == "" {
		var err error
		path, err = DefaultCredentialsPath()
		if err != nil {
			return Result{}, CredentialsAbsent, err
		}
	}

	profiles := profileNamesForAccount(opts.AccountName, opts.ProfileNames)
	validator := opts.Validator
	if validator == nil {
		validator = STSValidator{}
	}

	return resolveFirstValidProfile(ctx, profiles, path, validator)
}

// profileNamesForAccount returns AWS profile names to try, in order.
func profileNamesForAccount(accountName string, profileNames []string) []string {
	if len(profileNames) > 0 {
		return profileNames
	}
	return []string{SanitizeProfileName(accountName)}
}
