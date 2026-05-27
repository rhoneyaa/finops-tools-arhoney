// credential_provider.go defines how the CLI obtains credentials when no valid stored profile exists.
package aws

import (
	"context"
	"fmt"
)

// CredentialProvider supplies AWS credentials when stored profiles cannot be used.
type CredentialProvider interface {
	Obtain(ctx context.Context, accountName string) (ProfileSession, error)
}

// CredentialLookup carries provider-specific account lookup hints.
type CredentialLookup struct {
	AccountID string
	Names     []string
	Role      string
}

// LookupCredentialProvider is implemented by providers that support richer lookup hints.
type LookupCredentialProvider interface {
	ObtainWithLookup(ctx context.Context, lookup CredentialLookup) (ProfileSession, error)
}

// ProfileSessionFromEnvOutput builds a session from shell-style AWS_* env export text.
func ProfileSessionFromEnvOutput(stdout string) (ProfileSession, error) {
	sess := SessionFromEnvMap(ParseEnvOutput(stdout))
	if !sess.complete() {
		return ProfileSession{}, fmt.Errorf("%w: output missing AWS credential env vars", ErrObtainCredentials)
	}
	return sess, nil
}
