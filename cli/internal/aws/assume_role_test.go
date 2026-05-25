package aws

import (
	"context"
	"testing"
)

func TestAssumeRoleRequiresRoleARN(t *testing.T) {
	_, err := AssumeRole(context.Background(), ProfileSession{
		AccessKeyID: "A", SecretAccessKey: "S", SessionToken: "T",
	}, "", "sess")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAssumeRoleRequiresPayerKeys(t *testing.T) {
	_, err := AssumeRole(context.Background(), ProfileSession{}, "arn:aws:iam::111:role/X", "sess")
	if err == nil {
		t.Fatal("expected error")
	}
}
