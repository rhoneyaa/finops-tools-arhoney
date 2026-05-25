package account

import (
	"context"
	"strings"
	"testing"

	"github.com/openshift-online/finops-tools/cli/internal/awsauth"
	awsconfig "github.com/openshift-online/finops-tools/cli/internal/aws"
)

func TestAddAWSLinked(t *testing.T) {
	res, err := AddAWSLinked(context.Background(), AddAWSLinkedOptions{
		LinkedAccountID: "111111111111",
		PayerAccountID:  "123456789012",
		PayerAlias:      "rh-control",
		RoleARN:         "arn:aws:iam::111:role/X",
		EnsurePayer: func(context.Context, awsauth.EnsureOptions) (awsconfig.Result, error) {
			return awsconfig.Result{AccountID: "123456789012"}, nil
		},
		EnsureLinked: func(context.Context, awsconfig.EnsureLinkedOptions) (awsconfig.Result, error) {
			return awsconfig.Result{
				AccountID: "111111111111",
				ARN:       "arn:linked",
				Profile:   "111111111111",
				Refreshed: true,
			}, nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.AccountID != "111111111111" || res.Profile != "111111111111" {
		t.Fatalf("got %+v", res)
	}
}

func TestAddAWSLinkedRejectsMismatchedAccount(t *testing.T) {
	_, err := AddAWSLinked(context.Background(), AddAWSLinkedOptions{
		LinkedAccountID: "111111111111",
		PayerAccountID:  "123456789012",
		RoleARN:         "arn:aws:iam::111:role/X",
		EnsurePayer: func(context.Context, awsauth.EnsureOptions) (awsconfig.Result, error) {
			return awsconfig.Result{AccountID: "123456789012"}, nil
		},
		EnsureLinked: func(context.Context, awsconfig.EnsureLinkedOptions) (awsconfig.Result, error) {
			return awsconfig.Result{AccountID: "999999999999"}, nil
		},
	})
	if err == nil || !strings.Contains(err.Error(), "expected") {
		t.Fatalf("got %v", err)
	}
}
