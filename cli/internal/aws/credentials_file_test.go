// credentials_file_test.go tests reading and writing the shared credentials file.
package aws

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteProfilePreservesOtherSections(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")

	initial := `[other]
aws_access_key_id = OLD
aws_secret_access_key = OLDSECRET
aws_session_token = OLDTOKEN
region = eu-west-1
`
	if err := os.WriteFile(path, []byte(initial), 0o600); err != nil {
		t.Fatal(err)
	}

	sess := ProfileSession{
		AccessKeyID:     "NEW",
		SecretAccessKey: "NEWSECRET",
		SessionToken:    "NEWTOKEN",
		Region:          "us-east-1",
	}
	if err := WriteProfile(path, "rh-control", sess); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "[other]") || !strings.Contains(content, "OLD") {
		t.Fatalf("other profile lost:\n%s", content)
	}
	if !strings.Contains(content, "[rh-control]") || !strings.Contains(content, "NEW") {
		t.Fatalf("new profile missing:\n%s", content)
	}

	got, ok, err := ReadProfile(path, "rh-control")
	if err != nil || !ok || got.AccessKeyID != "NEW" {
		t.Fatalf("read rh-control: ok=%v err=%v got=%+v", ok, err, got)
	}
}

func TestWriteProfileStaticCredentials(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	if err := WriteProfile(path, "payer", ProfileSession{
		AccessKeyID: "AK", SecretAccessKey: "SK", SessionToken: "ST", Region: "us-east-1",
	}); err != nil {
		t.Fatal(err)
	}

	if err := WriteProfile(path, "payer", ProfileSession{
		AccessKeyID: "AK2", SecretAccessKey: "SK2", Region: "us-east-1",
	}); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "aws_session_token") {
		t.Fatalf("static profile should not keep session token:\n%s", data)
	}

	got, ok, err := ReadProfile(path, "payer")
	if err != nil || !ok || got.AccessKeyID != "AK2" || got.SessionToken != "" {
		t.Fatalf("profile: ok=%v err=%v got=%+v", ok, err, got)
	}
}

func TestReadProfileMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	_, ok, err := ReadProfile(path, "missing")
	if err != nil || ok {
		t.Fatalf("ok=%v err=%v", ok, err)
	}
}
