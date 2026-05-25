package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

// LoadSharedConfigProfile loads an AWS API config for a named ~/.aws profile.
func LoadSharedConfigProfile(ctx context.Context, profile string) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
}
