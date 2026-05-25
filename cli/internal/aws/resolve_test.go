package aws

import (
	"context"
	"path/filepath"
	"testing"
)

type fakeValidator struct {
	valid bool
	id    Identity
}

func (f fakeValidator) Validate(ctx context.Context, sess ProfileSession) (Identity, error) {
	if !f.valid {
		return Identity{}, ErrCredentialsInvalid
	}
	return f.id, nil
}

func TestResolveCredentialsUsesExistingProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	if err := WriteProfile(path, "rh-control", ProfileSession{
		AccessKeyID: "AK", SecretAccessKey: "SK", SessionToken: "ST", Region: "us-east-1",
	}); err != nil {
		t.Fatal(err)
	}

	res, status, err := ResolveCredentials(context.Background(), ResolveOptions{
		AccountName:     "rh-control",
		CredentialsPath: path,
		Validator: fakeValidator{
			valid: true,
			id:    Identity{AccountID: "123", ARN: "arn:x", UserID: "u"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if status != CredentialsValid || res.Refreshed || res.AccountID != "123" || res.Profile != "rh-control" {
		t.Fatalf("status=%v result: %+v", status, res)
	}
}

func TestResolveCredentialsInvalidProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	if err := WriteProfile(path, "rh-control", ProfileSession{
		AccessKeyID: "OLD", SecretAccessKey: "S", SessionToken: "T", Region: "us-east-1",
	}); err != nil {
		t.Fatal(err)
	}

	_, status, err := ResolveCredentials(context.Background(), ResolveOptions{
		AccountName:     "rh-control",
		CredentialsPath: path,
		Validator:       fakeValidator{valid: false},
	})
	if err != nil {
		t.Fatal(err)
	}
	if status != CredentialsInvalid {
		t.Fatalf("status = %v", status)
	}
}

func TestResolveCredentialsAbsentProfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")

	_, status, err := ResolveCredentials(context.Background(), ResolveOptions{
		AccountName:     "rh-control",
		CredentialsPath: path,
	})
	if err != nil {
		t.Fatal(err)
	}
	if status != CredentialsAbsent {
		t.Fatalf("status = %v", status)
	}
}

func TestResolveCredentialsUsesAliasProfileBeforeAccountID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	if err := WriteProfile(path, "rh-control", ProfileSession{
		AccessKeyID: "AK", SecretAccessKey: "SK", SessionToken: "ST", Region: "us-east-1",
	}); err != nil {
		t.Fatal(err)
	}

	res, status, err := ResolveCredentials(context.Background(), ResolveOptions{
		AccountName:     "123456789012",
		ProfileNames:    []string{"rh-control", "123456789012"},
		CredentialsPath: path,
		Validator: fakeValidator{
			valid: true,
			id:    Identity{AccountID: "123456789012", ARN: "arn:x", UserID: "u"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if status != CredentialsValid || res.Profile != "rh-control" {
		t.Fatalf("status=%v result: %+v", status, res)
	}
}

func TestStoreCredentialsWritesAndValidates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")

	res, err := StoreCredentials(context.Background(), StoreOptions{
		AccountName:     "123456789012",
		CredentialsPath: path,
		Validator: fakeValidator{
			valid: true,
			id:    Identity{AccountID: "123456789012", ARN: "arn:x", UserID: "u"},
		},
	}, ProfileSession{
		AccessKeyID: "AKIATEST", SecretAccessKey: "secret", Region: "us-east-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !res.Refreshed || res.AccountID != "123456789012" {
		t.Fatalf("result: %+v", res)
	}

	got, ok, err := ReadProfile(path, "123456789012")
	if err != nil || !ok || got.AccessKeyID != "AKIATEST" || got.SessionToken != "" {
		t.Fatalf("profile: ok=%v err=%v got=%+v", ok, err, got)
	}
}
