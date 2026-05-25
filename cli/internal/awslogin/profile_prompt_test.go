package awslogin

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestInteractiveProfileRunnerLogin(t *testing.T) {
	in := strings.NewReader("AKIATEST\nmy-secret-key\n")
	var out bytes.Buffer

	sess, err := InteractiveProfileRunner{In: in, Out: &out}.Obtain(context.Background(), "123456789012")
	if err != nil {
		t.Fatal(err)
	}
	if sess.AccessKeyID != "AKIATEST" || sess.SecretAccessKey != "my-secret-key" || sess.SessionToken != "" {
		t.Fatalf("session: %+v", sess)
	}
	if !strings.Contains(out.String(), "AWS Access Key ID") {
		t.Fatalf("output: %q", out.String())
	}
}

func TestInteractiveProfileRunnerLoginRejectsEmptyKey(t *testing.T) {
	in := strings.NewReader("\nsecret\n")
	_, err := InteractiveProfileRunner{In: in, Out: ioDiscard{}}.Obtain(context.Background(), "123456789012")
	if err == nil {
		t.Fatal("expected error")
	}
}

type ioDiscard struct{}

func (ioDiscard) Write(p []byte) (int, error) { return len(p), nil }
