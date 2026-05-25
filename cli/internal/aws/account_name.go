package aws

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

const organizationsRegion = "us-east-1"

// OrganizationsAccountsAPI is the subset of Organizations used for account names.
type OrganizationsAccountsAPI interface {
	DescribeAccount(
		ctx context.Context,
		params *organizations.DescribeAccountInput,
		optFns ...func(*organizations.Options),
	) (*organizations.DescribeAccountOutput, error)
	ListAccounts(
		ctx context.Context,
		params *organizations.ListAccountsInput,
		optFns ...func(*organizations.Options),
	) (*organizations.ListAccountsOutput, error)
}

func newOrganizationsClient(cfg aws.Config) *organizations.Client {
	return organizations.NewFromConfig(cfg, func(o *organizations.Options) {
		o.Region = organizationsRegion
	})
}

// AccountName returns the AWS Organizations account name for accountID.
func AccountName(ctx context.Context, cfg aws.Config, accountID string) (string, error) {
	return accountName(ctx, newOrganizationsClient(cfg), accountID)
}

func accountName(ctx context.Context, client OrganizationsAccountsAPI, accountID string) (string, error) {
	accountID = strings.TrimSpace(accountID)
	if accountID == "" {
		return "", fmt.Errorf("account ID is required")
	}
	out, err := client.DescribeAccount(ctx, &organizations.DescribeAccountInput{
		AccountId: aws.String(accountID),
	})
	if err != nil {
		return "", err
	}
	return accountNameFromOrgAccount(out.Account, accountID)
}

// ListAccountNames returns a map of account ID to AWS Organizations account name.
func ListAccountNames(ctx context.Context, cfg aws.Config) (map[string]string, error) {
	return listAccountNamesWithClient(ctx, newOrganizationsClient(cfg))
}

func listAccountNamesWithClient(ctx context.Context, client OrganizationsAccountsAPI) (map[string]string, error) {
	names := make(map[string]string)
	var token *string
	for {
		out, err := client.ListAccounts(ctx, &organizations.ListAccountsInput{NextToken: token})
		if err != nil {
			return nil, err
		}
		for _, acct := range out.Accounts {
			if name, err := accountNameFromOrgAccount(&acct, aws.ToString(acct.Id)); err == nil {
				names[aws.ToString(acct.Id)] = name
			}
		}
		if out.NextToken == nil || *out.NextToken == "" {
			break
		}
		token = out.NextToken
	}
	return names, nil
}

func accountNameFromOrgAccount(acct *types.Account, accountID string) (string, error) {
	if acct == nil {
		return "", fmt.Errorf("account %s not found", accountID)
	}
	name := strings.TrimSpace(aws.ToString(acct.Name))
	if name == "" {
		return "", fmt.Errorf("account %s has no name", accountID)
	}
	return name, nil
}
