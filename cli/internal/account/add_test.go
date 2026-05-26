// add_test.go tests AddAWS payer login with mocked credential ensure.
package account

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/openshift-online/finops-tools/cli/internal/awsauth"
	awsconfig "github.com/openshift-online/finops-tools/cli/internal/aws"
)

func TestAddAWSRegistersAccount(t *testing.T) {
	res, err := AddAWS(context.Background(), AddAWSOptions{
		AccountID: "123456789012",
		Ensure: func(context.Context, awsauth.EnsureOptions) (awsconfig.Result, error) {
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
