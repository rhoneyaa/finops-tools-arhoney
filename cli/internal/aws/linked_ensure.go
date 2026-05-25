package aws

import (
	"context"
	"fmt"
)

// AssumeRoleFunc obtains linked-account credentials from a payer session (injectable for tests).
type AssumeRoleFunc func(ctx context.Context, payerSession ProfileSession, roleARN, sessionName string) (ProfileSession, error)

// EnsureLinkedOptions configures EnsureLinkedCredentials.
type EnsureLinkedOptions struct {
	PayerAccountID     string
	PayerProfileNames  []string
	LinkedAccountID    string
	LinkedProfileNames []string
	RoleARN            string
	CredentialsPath    string
	Validator          CredentialValidator
	AssumeRoleFn       AssumeRoleFunc
}

// EnsureLinkedCredentials assumes a role into the linked account using payer credentials
// already stored under the payer profile, writes linked credentials, and validates the linked account ID.
// The caller must ensure payer credentials are valid before calling this function.
func EnsureLinkedCredentials(ctx context.Context, opts EnsureLinkedOptions) (Result, error) {
	if opts.PayerAccountID == "" {
		return Result{}, fmt.Errorf("payer account ID is required")
	}
	if opts.LinkedAccountID == "" {
		return Result{}, fmt.Errorf("linked account ID is required")
	}
	if opts.RoleARN == "" {
		return Result{}, fmt.Errorf("role ARN is required")
	}

	path := opts.CredentialsPath
	if path == "" {
		var err error
		path, err = DefaultCredentialsPath()
		if err != nil {
			return Result{}, err
		}
	}

	payerProfiles := profileNamesForAccount(opts.PayerAccountID, opts.PayerProfileNames)
	validator := opts.Validator
	if validator == nil {
		validator = STSValidator{}
	}

	payerRes, status, err := resolveFirstValidProfile(ctx, payerProfiles, path, validator)
	if err != nil {
		return Result{}, err
	}
	if status != CredentialsValid {
		return Result{}, fmt.Errorf("payer credentials: %w", errPayerCredentialsUnavailable(status, payerProfiles))
	}
	if payerRes.AccountID != opts.PayerAccountID {
		return Result{}, fmt.Errorf("payer session is account %s, expected %s", payerRes.AccountID, opts.PayerAccountID)
	}

	payerProfile := payerRes.Profile
	if payerProfile == "" {
		payerProfile = SanitizeProfileName(opts.PayerAccountID)
	}
	payerSess, ok, err := ReadProfile(path, payerProfile)
	if err != nil {
		return Result{}, err
	}
	if !ok {
		return Result{}, fmt.Errorf("%w: payer profile %q", ErrCredentialsNotFound, payerProfile)
	}

	assume := opts.AssumeRoleFn
	if assume == nil {
		assume = AssumeRole
	}
	linkedSess, err := assume(ctx, payerSess, opts.RoleARN, "finops-"+SanitizeProfileName(opts.LinkedAccountID))
	if err != nil {
		return Result{}, err
	}

	linkedProfiles := opts.LinkedProfileNames
	if len(linkedProfiles) == 0 {
		linkedProfiles = []string{SanitizeProfileName(opts.LinkedAccountID)}
	}
	writeProfile := linkedProfiles[0]

	if err := WriteProfile(path, writeProfile, linkedSess); err != nil {
		return Result{}, fmt.Errorf("write linked credentials profile: %w", err)
	}

	id, err := validator.Validate(ctx, linkedSess)
	if err != nil {
		return Result{}, fmt.Errorf("verify linked credentials: %w", err)
	}
	if id.AccountID != opts.LinkedAccountID {
		return Result{}, fmt.Errorf("linked role session is account %s, expected %s", id.AccountID, opts.LinkedAccountID)
	}

	return Result{
		AccountID: id.AccountID,
		ARN:       id.ARN,
		UserID:    id.UserID,
		Profile:   writeProfile,
		Refreshed: true,
	}, nil
}

func errPayerCredentialsUnavailable(status ResolveStatus, profiles []string) error {
	switch status {
	case CredentialsInvalid:
		return fmt.Errorf("%w: profiles %v", ErrCredentialsInvalid, profiles)
	default:
		return fmt.Errorf("%w: profiles %v", ErrCredentialsNotFound, profiles)
	}
}
