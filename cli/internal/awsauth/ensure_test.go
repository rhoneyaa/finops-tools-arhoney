package awsauth

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	awsconfig "github.com/openshift-online/finops-tools/cli/internal/aws"
)

type fakeValidator struct {
	valid bool
	id    awsconfig.Identity
}

func (f fakeValidator) Validate(ctx context.Context, sess awsconfig.ProfileSession) (awsconfig.Identity, error) {
	if !f.valid {
		return awsconfig.Identity{}, awsconfig.ErrCredentialsInvalid
	}
	return f.id, nil
}

type fakeProvider struct {
	sess awsconfig.ProfileSession
	err  error
}

func (f fakeProvider) Obtain(ctx context.Context, accountName string) (awsconfig.ProfileSession, error) {
	if f.err != nil {
		return awsconfig.ProfileSession{}, f.err
	}
	return f.sess, nil
}

func TestEnsureAccountCredentialsUsesExistingProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	if err := awsconfig.WriteProfile(path, "rh-control", awsconfig.ProfileSession{
		AccessKeyID: "AK", SecretAccessKey: "SK", SessionToken: "ST", Region: "us-east-1",
	}); err != nil {
		t.Fatal(err)
	}

	res, err := EnsureAccountCredentials(context.Background(), EnsureOptions{
		AccountName:     "rh-control",
		Method:          MethodSAML,
		CredentialsPath: path,
		Validator: fakeValidator{
			valid: true,
			id:    awsconfig.Identity{AccountID: "123", ARN: "arn:x", UserID: "u"},
		},
		Provider: fakeProvider{err: errors.New("obtain should not run")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Refreshed || res.AccountID != "123" {
		t.Fatalf("result: %+v", res)
	}
}

func TestEnsureAccountCredentialsRefreshesWhenInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	if err := awsconfig.WriteProfile(path, "rh-control", awsconfig.ProfileSession{
		AccessKeyID: "OLD", SecretAccessKey: "S", SessionToken: "T", Region: "us-east-1",
	}); err != nil {
		t.Fatal(err)
	}

	res, err := EnsureAccountCredentials(context.Background(), EnsureOptions{
		AccountName:     "rh-control",
		Method:          MethodSAML,
		CredentialsPath: path,
		Validator: fakeValidatorFunc(func(_ context.Context, sess awsconfig.ProfileSession) (awsconfig.Identity, error) {
			if sess.AccessKeyID == "NEW" {
				return awsconfig.Identity{AccountID: "123", ARN: "arn:x", UserID: "u"}, nil
			}
			return awsconfig.Identity{}, awsconfig.ErrCredentialsInvalid
		}),
		Provider: fakeProvider{
			sess: awsconfig.ProfileSession{
				AccessKeyID: "NEW", SecretAccessKey: "SK", SessionToken: "ST", Region: "us-east-1",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Refreshed || res.AccountID != "123" {
		t.Fatalf("result: %+v", res)
	}
}

func TestEnsureAccountCredentialsForceSkipsValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	if err := awsconfig.WriteProfile(path, "rh-control", awsconfig.ProfileSession{
		AccessKeyID: "OLD", SecretAccessKey: "S", SessionToken: "T", Region: "us-east-1",
	}); err != nil {
		t.Fatal(err)
	}

	calls := 0
	_, err := EnsureAccountCredentials(context.Background(), EnsureOptions{
		AccountName:     "rh-control",
		Force:           true,
		Method:          MethodSAML,
		CredentialsPath: path,
		Validator: fakeValidator{
			valid: true,
			id:    awsconfig.Identity{AccountID: "999", ARN: "arn:x", UserID: "u"},
		},
		Provider: fakeProviderFunc(func(context.Context, string) (awsconfig.ProfileSession, error) {
			calls++
			return awsconfig.ProfileSession{
				AccessKeyID: "NEW", SecretAccessKey: "S", SessionToken: "T", Region: "us-east-1",
			}, nil
		}),
	})
	if err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("obtain calls = %d", calls)
	}
}

func TestEnsureAccountCredentialsProfileWithoutProviderFails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")

	_, err := EnsureAccountCredentials(context.Background(), EnsureOptions{
		AccountName:     "rh-control",
		Method:          MethodProfile,
		CredentialsPath: path,
	})
	if !errors.Is(err, awsconfig.ErrCredentialsNotFound) {
		t.Fatalf("err = %v", err)
	}
}

type fakeProviderFunc func(ctx context.Context, accountName string) (awsconfig.ProfileSession, error)

func (f fakeProviderFunc) Obtain(ctx context.Context, accountName string) (awsconfig.ProfileSession, error) {
	return f(ctx, accountName)
}

type fakeValidatorFunc func(ctx context.Context, sess awsconfig.ProfileSession) (awsconfig.Identity, error)

func (f fakeValidatorFunc) Validate(ctx context.Context, sess awsconfig.ProfileSession) (awsconfig.Identity, error) {
	return f(ctx, sess)
}
