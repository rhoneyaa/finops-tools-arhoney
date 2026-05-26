// auth_method_resolve_test.go tests resolveAuthMethod flag vs config precedence.
package cmd

import (
	"path/filepath"
	"testing"

	"github.com/openshift-online/finops-tools/cli/internal/awsauth"
	"github.com/openshift-online/finops-tools/cli/internal/configstore"
	"github.com/spf13/cobra"
)

func TestResolveAuthMethodUsesConfigDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := configstore.SetDefault(path, "aws.auth-method", "profile"); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("auth-method", string(awsauth.MethodSAML), "")
	cmd.Flags().String("config", "", "")

	method, err := resolveAuthMethod(cmd, path, string(awsauth.MethodSAML))
	if err != nil {
		t.Fatal(err)
	}
	if method != awsauth.MethodProfile {
		t.Fatalf("got %q", method)
	}
}

func TestResolveAuthMethodFlagOverridesConfig(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := configstore.SetDefault(path, "aws.auth-method", "profile"); err != nil {
		t.Fatal(err)
	}

	cmd := &cobra.Command{}
	cmd.Flags().String("auth-method", string(awsauth.MethodSAML), "")
	if err := cmd.Flags().Set("auth-method", "saml"); err != nil {
		t.Fatal(err)
	}

	method, err := resolveAuthMethod(cmd, path, "saml")
	if err != nil {
		t.Fatal(err)
	}
	if method != awsauth.MethodSAML {
		t.Fatalf("got %q", method)
	}
}
