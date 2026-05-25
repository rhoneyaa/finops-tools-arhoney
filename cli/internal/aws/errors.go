package aws

import "errors"

var (
	// ErrCredentialsInvalid means stored credentials are missing or rejected by STS.
	ErrCredentialsInvalid = errors.New("aws credentials invalid or expired")
	// ErrCredentialsNotFound means the profile is absent from the credentials file.
	ErrCredentialsNotFound = errors.New("aws credentials profile not found")
	// ErrObtainCredentials means a CredentialProvider did not return usable credentials.
	ErrObtainCredentials = errors.New("aws credentials obtain failed")
	// ErrObtainToolNotFound means an external tool required by the caller is not installed.
	ErrObtainToolNotFound = errors.New("aws credentials tool not found")
)
