package aws

import (
	"context"
	"errors"
)

// resolveStoredProfile checks ~/.aws/credentials then the shared config profile.
func resolveStoredProfile(
	ctx context.Context,
	profile string,
	credentialsPath string,
	validator CredentialValidator,
) (Result, ResolveStatus, error) {
	if validator == nil {
		validator = STSValidator{}
	}

	sess, found, err := ReadProfile(credentialsPath, profile)
	if err != nil {
		return Result{}, CredentialsAbsent, err
	}
	var invalidInCredentialsFile bool
	if found {
		id, err := validator.Validate(ctx, sess)
		if err == nil {
			return Result{
				AccountID: id.AccountID,
				ARN:       id.ARN,
				UserID:    id.UserID,
				Profile:   profile,
				Refreshed: false,
			}, CredentialsValid, nil
		}
		if errors.Is(err, ErrCredentialsInvalid) {
			invalidInCredentialsFile = true
		} else {
			return Result{}, CredentialsAbsent, err
		}
	}

	id, err := SharedConfigValidator{Profile: profile}.Validate(ctx, ProfileSession{})
	if err == nil {
		return Result{
			AccountID: id.AccountID,
			ARN:       id.ARN,
			UserID:    id.UserID,
			Profile:   profile,
			Refreshed: false,
		}, CredentialsValid, nil
	}
	if errors.Is(err, ErrCredentialsInvalid) || invalidInCredentialsFile {
		return Result{}, CredentialsInvalid, nil
	}

	return Result{}, CredentialsAbsent, nil
}

// resolveFirstValidProfile tries each profile name until one validates with STS.
func resolveFirstValidProfile(
	ctx context.Context,
	profiles []string,
	credentialsPath string,
	validator CredentialValidator,
) (Result, ResolveStatus, error) {
	var sawInvalid bool
	for _, profile := range profiles {
		res, status, err := resolveStoredProfile(ctx, profile, credentialsPath, validator)
		if err != nil {
			return Result{}, CredentialsAbsent, err
		}
		switch status {
		case CredentialsValid:
			return res, CredentialsValid, nil
		case CredentialsInvalid:
			sawInvalid = true
		}
	}
	if sawInvalid {
		return Result{}, CredentialsInvalid, nil
	}
	return Result{}, CredentialsAbsent, nil
}
