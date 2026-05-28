package account

import (
	"context"
	"errors"
	"reflect"
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

func (f fakeOrganizations) ListTagsForResource(
	_ context.Context,
	_ *organizations.ListTagsForResourceInput,
	_ ...func(*organizations.Options),
) (*organizations.ListTagsForResourceOutput, error) {
	return &organizations.ListTagsForResourceOutput{}, nil
}

func (f fakeOrganizations) DescribeOrganization(
	_ context.Context,
	_ *organizations.DescribeOrganizationInput,
	_ ...func(*organizations.Options),
) (*organizations.DescribeOrganizationOutput, error) {
	return nil, errors.New("not implemented")
}

type fakeOrganizationsTags struct {
	pages []*organizations.ListTagsForResourceOutput
	err   error
	call  int
}

func (f *fakeOrganizationsTags) ListTagsForResource(
	_ context.Context,
	_ *organizations.ListTagsForResourceInput,
	_ ...func(*organizations.Options),
) (*organizations.ListTagsForResourceOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	if f.call >= len(f.pages) {
		return &organizations.ListTagsForResourceOutput{}, nil
	}
	out := f.pages[f.call]
	f.call++
	return out, nil
}

func (f *fakeOrganizationsTags) DescribeAccount(
	_ context.Context,
	_ *organizations.DescribeAccountInput,
	_ ...func(*organizations.Options),
) (*organizations.DescribeAccountOutput, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeOrganizationsTags) ListAccounts(
	_ context.Context,
	_ *organizations.ListAccountsInput,
	_ ...func(*organizations.Options),
) (*organizations.ListAccountsOutput, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeOrganizationsTags) DescribeOrganization(
	_ context.Context,
	_ *organizations.DescribeOrganizationInput,
	_ ...func(*organizations.Options),
) (*organizations.DescribeOrganizationOutput, error) {
	return nil, errors.New("not implemented")
}

type fakeDescribeOrganizationClient struct {
	output *organizations.DescribeOrganizationOutput
	err    error
}

func (f fakeDescribeOrganizationClient) DescribeOrganization(
	_ context.Context,
	_ *organizations.DescribeOrganizationInput,
	_ ...func(*organizations.Options),
) (*organizations.DescribeOrganizationOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.output, nil
}

func (f fakeDescribeOrganizationClient) DescribeAccount(
	_ context.Context,
	_ *organizations.DescribeAccountInput,
	_ ...func(*organizations.Options),
) (*organizations.DescribeAccountOutput, error) {
	return nil, errors.New("not implemented")
}

func (f fakeDescribeOrganizationClient) ListAccounts(
	_ context.Context,
	_ *organizations.ListAccountsInput,
	_ ...func(*organizations.Options),
) (*organizations.ListAccountsOutput, error) {
	return nil, errors.New("not implemented")
}

func (f fakeDescribeOrganizationClient) ListTagsForResource(
	_ context.Context,
	_ *organizations.ListTagsForResourceInput,
	_ ...func(*organizations.Options),
) (*organizations.ListTagsForResourceOutput, error) {
	return nil, errors.New("not implemented")
}

func TestAccountName(t *testing.T) {
	client := fakeOrganizations{accounts: map[string]string{
		"111111111111": "Quay Production",
	}}
	name, err := accountNameWithClient(context.Background(), client, "111111111111")
	if err != nil || name != "Quay Production" {
		t.Fatalf("name = %q err = %v", name, err)
	}
	_, err = accountNameWithClient(context.Background(), client, "999999999999")
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

func TestResolveAccountNamesDescribePerID(t *testing.T) {
	client := fakeOrganizations{accounts: map[string]string{
		"111111111111": "Quay Production",
		"222222222222": "Staging",
	}}
	names, err := resolveAccountNamesWithClient(context.Background(), client, []string{"111111111111"})
	if err != nil {
		t.Fatal(err)
	}
	if names["111111111111"] != "Quay Production" {
		t.Fatalf("names = %+v", names)
	}
}

func TestResolveAccountNamesUniqueIDs(t *testing.T) {
	ids := uniqueAccountIDs([]string{" 111 ", "111", ""})
	if len(ids) != 1 || ids[0] != "111" {
		t.Fatalf("ids = %v", ids)
	}
}

func TestListTagsWithClient(t *testing.T) {
	client := &fakeOrganizationsTags{
		pages: []*organizations.ListTagsForResourceOutput{
			{
				Tags: []types.Tag{
					{Key: aws.String("owner"), Value: aws.String("team-b")},
					{Key: aws.String("env"), Value: aws.String("prod")},
				},
				NextToken: aws.String("page-2"),
			},
			{
				Tags: []types.Tag{
					{Key: aws.String("owner"), Value: aws.String("team-a")},
				},
			},
		},
	}

	tags, err := listTagsWithClient(context.Background(), client, "123456789012")
	if err != nil {
		t.Fatal(err)
	}
	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tags))
	}
	if tags[0].Key != "env" || tags[0].Value != "prod" {
		t.Fatalf("unexpected first tag: %+v", tags[0])
	}
	if tags[1].Key != "owner" || tags[1].Value != "team-a" {
		t.Fatalf("unexpected second tag: %+v", tags[1])
	}
	if tags[2].Key != "owner" || tags[2].Value != "team-b" {
		t.Fatalf("unexpected third tag: %+v", tags[2])
	}
}

func TestListTagsWithClientValidationAndErrors(t *testing.T) {
	client := &fakeOrganizationsTags{}
	if _, err := listTagsWithClient(context.Background(), client, " "); err == nil {
		t.Fatal("expected account ID validation error")
	}

	wantErr := errors.New("boom")
	client.err = wantErr
	_, err := listTagsWithClient(context.Background(), client, "123456789012")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected wrapped error %v, got %v", wantErr, err)
	}
}

func TestDetectAccountKindWithClientPayer(t *testing.T) {
	client := fakeDescribeOrganizationClient{
		output: &organizations.DescribeOrganizationOutput{
			Organization: organizationWithManagementAccountID("123456789012"),
		},
	}
	kind, err := detectAccountKindWithClient(context.Background(), client, "123456789012")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kind != AccountKindPayer {
		t.Fatalf("kind = %q, want %q", kind, AccountKindPayer)
	}
}

func TestDetectAccountKindWithClientLinked(t *testing.T) {
	client := fakeDescribeOrganizationClient{
		output: &organizations.DescribeOrganizationOutput{
			Organization: organizationWithManagementAccountID("999999999999"),
		},
	}
	kind, err := detectAccountKindWithClient(context.Background(), client, "123456789012")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kind != AccountKindLinked {
		t.Fatalf("kind = %q, want %q", kind, AccountKindLinked)
	}
}

func TestDetectAccountKindWithClientUnknownOnDescribeFailure(t *testing.T) {
	client := fakeDescribeOrganizationClient{err: errors.New("access denied")}
	kind, err := detectAccountKindWithClient(context.Background(), client, "123456789012")
	if err == nil {
		t.Fatal("expected error")
	}
	if kind != AccountKindUnknown {
		t.Fatalf("kind = %q, want %q", kind, AccountKindUnknown)
	}
}

func TestOrganizationManagementAccountIDFallsBackToOrganizationID(t *testing.T) {
	org := &types.Organization{Id: aws.String("123456789012")}
	if got := organizationManagementAccountID(org); got != "123456789012" {
		t.Fatalf("got %q want %q", got, "123456789012")
	}
}

func organizationWithManagementAccountID(accountID string) *types.Organization {
	org := &types.Organization{}
	v := reflect.ValueOf(org).Elem()
	for _, fieldName := range []string{"ManagementAccountId", "MasterAccountId"} {
		field := v.FieldByName(fieldName)
		if !field.IsValid() || !field.CanSet() || field.Kind() != reflect.Ptr || field.Type().Elem().Kind() != reflect.String {
			continue
		}
		id := accountID
		field.Set(reflect.ValueOf(&id))
		return org
	}
	org.Id = aws.String(accountID)
	return org
}
