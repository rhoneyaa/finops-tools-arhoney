// saml_runner.go performs Red Hat Kerberos + SAML login and returns temporary AWS credentials.
package awslogin

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	awsconfig "github.com/openshift-online/finops-tools/cli/internal/aws"
	"github.com/openshift-online/finops-tools/cli/internal/awslogin/rhsaml"
)

const (
	defaultLoginTimeout = 120 * time.Second

	// samlPromptHint explains SAML login prerequisites shown before login.
	samlPromptHint = `Connect to Red Hat VPN first; run kinit if you have no valid Kerberos ticket yet.`
)

// SAMLLoginRunner performs Red Hat SAML login and retrieves AWS credentials.
type SAMLLoginRunner struct {
	Timeout time.Duration
	// NoticeWriter receives pre-login hints (default stderr). Set to io.Discard in tests.
	NoticeWriter io.Writer
}

func (r SAMLLoginRunner) notice(accountName string) {
	w := r.NoticeWriter
	if w == nil {
		w = os.Stderr
	}
	_, _ = fmt.Fprintf(w, "Running Red Hat SAML login for %q...\n", accountName)
	_, _ = fmt.Fprintln(w, samlPromptHint)
}

// Obtain runs Red Hat SAML login using accountName as the lookup key.
func (r SAMLLoginRunner) Obtain(ctx context.Context, accountName string) (awsconfig.ProfileSession, error) {
	return r.ObtainWithLookup(ctx, awsconfig.CredentialLookup{
		AccountID: accountName,
		Names:     []string{accountName},
	})
}

// ObtainWithLookup runs Red Hat SAML login with richer account lookup keys.
func (r SAMLLoginRunner) ObtainWithLookup(ctx context.Context, lookup awsconfig.CredentialLookup) (awsconfig.ProfileSession, error) {
	timeout := r.Timeout
	if timeout == 0 {
		timeout = defaultLoginTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	accountLabel := lookup.AccountID
	if accountLabel == "" && len(lookup.Names) > 0 {
		accountLabel = lookup.Names[0]
	}
	if accountLabel == "" {
		accountLabel = "unknown-account"
	}
	r.notice(accountLabel)

	client := rhsaml.Client{}
	sess, err := client.Obtain(ctx, rhsaml.ObtainRequest{
		LookupKeys: lookup.Names,
		Role:       lookup.Role,
	})
	if err != nil {
		return awsconfig.ProfileSession{}, fmt.Errorf("%w: %v (VPN + valid Kerberos ticket required; run kinit if needed)", awsconfig.ErrObtainCredentials, err)
	}
	return sess, nil
}
