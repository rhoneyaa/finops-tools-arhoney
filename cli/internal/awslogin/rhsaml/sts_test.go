package rhsaml

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
)

type fakeSTSClient struct {
	out *sts.AssumeRoleWithSAMLOutput
	err error
}

func (f fakeSTSClient) AssumeRoleWithSAML(ctx context.Context, params *sts.AssumeRoleWithSAMLInput, optFns ...func(*sts.Options)) (*sts.AssumeRoleWithSAMLOutput, error) {
	return f.out, f.err
}

func TestAssumeRoleWithSAMLClient(t *testing.T) {
	client := fakeSTSClient{
		out: &sts.AssumeRoleWithSAMLOutput{
			Credentials: &types.Credentials{
				AccessKeyId:     aws.String("AKIA"),
				SecretAccessKey: aws.String("SECRET"),
				SessionToken:    aws.String("TOKEN"),
			},
		},
	}
	sess, err := assumeRoleWithSAMLClient(context.Background(), client, "us-east-1", awsAccount{
		UID:      "123456789012",
		RoleARN:  "arn:aws:iam::123456789012:role/ReadOnlyAccess",
		RoleName: "ReadOnlyAccess",
	}, "assertion", 3600)
	if err != nil {
		t.Fatalf("assumeRoleWithSAMLClient: %v", err)
	}
	if sess.AccessKeyID != "AKIA" || sess.Region != "us-east-1" {
		t.Fatalf("session: %+v", sess)
	}
}
