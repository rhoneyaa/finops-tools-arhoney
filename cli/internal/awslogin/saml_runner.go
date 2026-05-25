package awslogin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	awsconfig "github.com/openshift-online/finops-tools/cli/internal/aws"
)

const (
	defaultLoginTimeout = 120 * time.Second
	samlLoginBinary     = "rh-aws-saml-login"

	// samlPromptHint explains which password rh-aws-saml-login may ask for.
	samlPromptHint = `If prompted for a password, enter your Red Hat Kerberos password` +
		` (the same one as kinit / RH laptop login — not your AWS console password).` +
		` Connect to Red Hat VPN first; run kinit if you have no ticket yet.`
)

// SAMLLoginRunner invokes the rh-aws-saml-login CLI.
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
	_, _ = fmt.Fprintf(w, "Running %s for %q...\n", samlLoginBinary, accountName)
	_, _ = fmt.Fprintln(w, samlPromptHint)
}

// Obtain runs rh-aws-saml-login --output env for the account.
func (r SAMLLoginRunner) Obtain(ctx context.Context, accountName string) (awsconfig.ProfileSession, error) {
	timeout := r.Timeout
	if timeout == 0 {
		timeout = defaultLoginTimeout
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	r.notice(accountName)

	cmd := exec.CommandContext(ctx, samlLoginBinary, "--output", "env", accountName)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return awsconfig.ProfileSession{}, fmt.Errorf("%w: %s timed out after %s (check VPN; run kinit if Kerberos ticket missing)", awsconfig.ErrObtainCredentials, samlLoginBinary, timeout)
		}
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			msg := string(exitErr.Stderr)
			if msg == "" {
				msg = string(out)
			}
			if msg == "" {
				msg = "non-zero exit"
			}
			return awsconfig.ProfileSession{}, fmt.Errorf("%w: %s (VPN + Kerberos required; password prompts are for your Red Hat account, not AWS)", awsconfig.ErrObtainCredentials, msg)
		}
		if errors.Is(err, exec.ErrNotFound) {
			return awsconfig.ProfileSession{}, fmt.Errorf("%w: install with: uv tool install %s", awsconfig.ErrObtainToolNotFound, samlLoginBinary)
		}
		return awsconfig.ProfileSession{}, fmt.Errorf("%w: %w", awsconfig.ErrObtainCredentials, err)
	}

	return awsconfig.ProfileSessionFromEnvOutput(string(out))
}
