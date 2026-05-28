package account

import "github.com/aws/aws-sdk-go-v2/aws"

const (
	organizationsRegion      = "us-east-1"
	accountNameListThreshold = 50
)

// Tag is one AWS Organizations account tag.
type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// OrganizationAccount is one AWS Organizations account directory entry.
type OrganizationAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// AccountKind describes whether a validated account session is payer or linked.
type AccountKind string

const (
	AccountKindPayer   AccountKind = "payer"
	AccountKindLinked  AccountKind = "linked"
	AccountKindUnknown AccountKind = "unknown"
)

// Query identifies a target account and the credentials used to query it.
type Query struct {
	AccountID string
	AWSConfig aws.Config
}
