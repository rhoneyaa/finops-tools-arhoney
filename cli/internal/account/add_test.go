// add_test.go tests AddAWS payer login with mocked credential ensure.
package account

import (
	"context"
	"errors"
	"strings"
	"testing"

	awsconfig "github.com/openshift-online/finops-tools/cli/internal/aws"
	"github.com/openshift-online/finops-tools/cli/internal/awsauth"
)

func TestAddAWSRegistersAccount(t *testing.T) {
	var gotEnsureOpts awsauth.EnsureOptions
	res, err := AddAWS(context.Background(), AddAWSOptions{
		AccountID: "123456789012",
		Alias:     "rh-control",
		Ensure: func(_ context.Context, opts awsauth.EnsureOptions) (awsconfig.Result, error) {
			gotEnsureOpts = opts
			return awsconfig.Result{
				AccountID: "123456789012",
				ARN:       "arn:aws:sts::123456789012:assumed-role/x",
				Profile:   "123456789012",
				Refreshed: true,
			}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.AccountID != "123456789012" || res.Profile != "123456789012" {
		t.Fatalf("got %+v", res)
	}
	if gotEnsureOpts.Lookup.AccountID != "123456789012" {
		t.Fatalf("lookup accountID = %q", gotEnsureOpts.Lookup.AccountID)
	}
	if len(gotEnsureOpts.Lookup.Names) != 2 || gotEnsureOpts.Lookup.Names[0] != "rh-control" {
		t.Fatalf("lookup names = %v", gotEnsureOpts.Lookup.Names)
	}
}

func TestAddAWSRejectsMismatchedAccount(t *testing.T) {
	_, err := AddAWS(context.Background(), AddAWSOptions{
		AccountID: "123456789012",
		Ensure: func(context.Context, awsauth.EnsureOptions) (awsconfig.Result, error) {
			return awsconfig.Result{AccountID: "999999999999"}, nil
		},
	})
	if err == nil || !strings.Contains(err.Error(), "expected") {
		t.Fatalf("got %v", err)
	}
}

func TestAddAWSRequiresAccountID(t *testing.T) {
	_, err := AddAWS(context.Background(), AddAWSOptions{})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAddGCPNotImplemented(t *testing.T) {
	_, err := Add(context.Background(), AddOptions{Provider: ProviderGCP, AccountID: "123456789012"})
	if err == nil || !errors.Is(err, errProviderNotImplemented) {
		t.Fatalf("got %v", err)
	}
}
