package aws

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func TestCheckCredentialsValid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	if err := WriteProfile(path, "rh-control", ProfileSession{
		AccessKeyID: "AK", SecretAccessKey: "SK", SessionToken: "ST", Region: "us-east-1",
	}); err != nil {
		t.Fatal(err)
	}

	res, err := CheckCredentials(context.Background(), CheckOptions{
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
	if res.Profile != "rh-control" || res.AccountID != "123" {
		t.Fatalf("result: %+v", res)
	}
}

func TestCheckCredentialsNotFound(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")

	_, err := CheckCredentials(context.Background(), CheckOptions{
		AccountName:     "rh-control",
		CredentialsPath: path,
	})
	if !errors.Is(err, ErrCredentialsNotFound) {
		t.Fatalf("err = %v", err)
	}
}

func TestCheckCredentialsInvalid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	if err := WriteProfile(path, "rh-control", ProfileSession{
		AccessKeyID: "AK", SecretAccessKey: "SK", SessionToken: "ST", Region: "us-east-1",
	}); err != nil {
		t.Fatal(err)
	}

	_, err := CheckCredentials(context.Background(), CheckOptions{
		AccountName:     "rh-control",
		CredentialsPath: path,
		Validator:       fakeValidator{valid: false},
	})
	if !errors.Is(err, ErrCredentialsInvalid) {
		t.Fatalf("err = %v", err)
	}
}
