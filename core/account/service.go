package account

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
)

// OrganizationsAPI is the subset of Organizations used by core account operations.
type OrganizationsAPI interface {
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
	ListTagsForResource(
		ctx context.Context,
		params *organizations.ListTagsForResourceInput,
		optFns ...func(*organizations.Options),
	) (*organizations.ListTagsForResourceOutput, error)
	DescribeOrganization(
		ctx context.Context,
		params *organizations.DescribeOrganizationInput,
		optFns ...func(*organizations.Options),
	) (*organizations.DescribeOrganizationOutput, error)
}

type organizationsClientFactory func(aws.Config) OrganizationsAPI

func newOrganizationsClient(cfg aws.Config) OrganizationsAPI {
	return organizations.NewFromConfig(cfg, func(o *organizations.Options) {
		o.Region = organizationsRegion
	})
}
