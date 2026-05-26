// account_name_test.go tests Organizations account name lookup with a mocked API.
package aws

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
)

type fakeOrganizations struct {
	accounts map[string]string
}

func (f fakeOrganizations) DescribeAccount(
	_ context.Context,
	params *organizations.DescribeAccountInput,
	_ ...func(*organizations.Options),
) (*organizations.DescribeAccountOutput, error) {
	id := aws.ToString(params.AccountId)
	name, ok := f.accounts[id]
	if !ok {
		return nil, errors.New("AccountNotFoundException")
	}
	return &organizations.DescribeAccountOutput{
		Account: &types.Account{Id: params.AccountId, Name: aws.String(name)},
	}, nil
}

func (f fakeOrganizations) ListAccounts(
	_ context.Context,
	_ *organizations.ListAccountsInput,
	_ ...func(*organizations.Options),
) (*organizations.ListAccountsOutput, error) {
	var accounts []types.Account
	for id, name := range f.accounts {
		accounts = append(accounts, types.Account{
			Id:   aws.String(id),
			Name: aws.String(name),
		})
	}
	return &organizations.ListAccountsOutput{Accounts: accounts}, nil
}

func TestAccountName(t *testing.T) {
	client := fakeOrganizations{accounts: map[string]string{
		"111111111111": "Quay Production",
	}}
	name, err := accountName(context.Background(), client, "111111111111")
	if err != nil || name != "Quay Production" {
		t.Fatalf("name = %q err = %v", name, err)
	}
	_, err = accountName(context.Background(), client, "999999999999")
	if err == nil {
		t.Fatal("expected error for missing account")
	}
}

func TestListAccountNames(t *testing.T) {
	client := fakeOrganizations{accounts: map[string]string{
		"111111111111": "Quay Production",
		"222222222222": "Staging",
	}}
	names, err := listAccountNamesWithClient(context.Background(), client)
	if err != nil {
		t.Fatal(err)
	}
	if names["111111111111"] != "Quay Production" || names["222222222222"] != "Staging" {
		t.Fatalf("names = %+v", names)
	}
}
