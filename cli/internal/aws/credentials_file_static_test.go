// credentials_file_static_test.go tests static credential file edge cases.
package aws

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadProfileStaticCredentials(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "credentials")
	content := `[staging1]
aws_access_key_id = AKIATEST
aws_secret_access_key = secret
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	sess, ok, err := ReadProfile(path, "staging1")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected profile to be found")
	}
	if sess.AccessKeyID != "AKIATEST" || sess.SecretAccessKey != "secret" || sess.SessionToken != "" {
		t.Fatalf("session: %+v", sess)
	}
}
