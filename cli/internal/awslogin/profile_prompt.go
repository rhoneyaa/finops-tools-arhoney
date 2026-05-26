// Package awslogin runs external and interactive AWS login flows (SAML CLI, access-key prompt).
package awslogin

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	awsconfig "github.com/openshift-online/finops-tools/cli/internal/aws"
	"golang.org/x/term"
)

// InteractiveProfileRunner prompts for long-lived AWS access keys when a profile is missing.
type InteractiveProfileRunner struct {
	In  io.Reader
	Out io.Writer
	// PasswordIn is used for hidden secret entry; defaults to os.Stdin when nil.
	PasswordIn *os.File
}

// Obtain reads AWS access key ID and secret access key from the terminal.
func (r InteractiveProfileRunner) Obtain(ctx context.Context, _ string) (awsconfig.ProfileSession, error) {
	_ = ctx
	in := r.In
	if in == nil {
		in = os.Stdin
	}
	out := r.Out
	if out == nil {
		out = os.Stdout
	}

	br := bufio.NewReader(in)
	accessKeyID, err := readLineFrom(br, out, "AWS Access Key ID: ")
	if err != nil {
		return awsconfig.ProfileSession{}, err
	}
	accessKeyID = strings.TrimSpace(accessKeyID)
	if accessKeyID == "" {
		return awsconfig.ProfileSession{}, fmt.Errorf("%w: access key ID is required", awsconfig.ErrObtainCredentials)
	}

	secret, err := readSecret(br, in, out, r.PasswordIn)
	if err != nil {
		return awsconfig.ProfileSession{}, err
	}
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return awsconfig.ProfileSession{}, fmt.Errorf("%w: secret access key is required", awsconfig.ErrObtainCredentials)
	}

	return awsconfig.ProfileSession{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secret,
		Region:          "us-east-1",
	}, nil
}

func readLineFrom(br *bufio.Reader, out io.Writer, prompt string) (string, error) {
	if _, err := fmt.Fprint(out, prompt); err != nil {
		return "", err
	}
	line, err := br.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(line, "\n"), nil
}

func readSecret(br *bufio.Reader, in io.Reader, out io.Writer, passwordIn *os.File) (string, error) {
	if _, err := fmt.Fprint(out, "AWS Secret Access Key: "); err != nil {
		return "", err
	}
	fdIn := passwordIn
	if fdIn == nil {
		if f, ok := in.(*os.File); ok {
			fdIn = f
		}
	}
	if fdIn != nil && term.IsTerminal(int(fdIn.Fd())) {
		secret, err := term.ReadPassword(int(fdIn.Fd()))
		if _, werr := fmt.Fprintln(out); werr != nil {
			return "", werr
		}
		if err != nil {
			return "", err
		}
		return string(secret), nil
	}
	line, err := br.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(line, "\n"), nil
}
