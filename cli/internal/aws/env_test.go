// env_test.go tests parsing of shell-style AWS_* export lines.
package aws

import "testing"

func TestParseEnvOutput(t *testing.T) {
	stdout := `export AWS_ACCESS_KEY_ID=AKIAEXAMPLE
export AWS_SECRET_ACCESS_KEY='secret'
AWS_SESSION_TOKEN="token"
AWS_REGION=us-west-2
`
	env := ParseEnvOutput(stdout)
	if env["AWS_ACCESS_KEY_ID"] != "AKIAEXAMPLE" {
		t.Fatalf("access key: %q", env["AWS_ACCESS_KEY_ID"])
	}
	if env["AWS_SECRET_ACCESS_KEY"] != "secret" {
		t.Fatalf("secret: %q", env["AWS_SECRET_ACCESS_KEY"])
	}
	if env["AWS_SESSION_TOKEN"] != "token" {
		t.Fatalf("token: %q", env["AWS_SESSION_TOKEN"])
	}
	if env["AWS_REGION"] != "us-west-2" {
		t.Fatalf("region: %q", env["AWS_REGION"])
	}

	sess := SessionFromEnvMap(env)
	if !sess.complete() || sess.Region != "us-west-2" {
		t.Fatalf("session: %+v", sess)
	}
}

func TestProfileSessionFromEnvOutput(t *testing.T) {
	stdout := `export AWS_ACCESS_KEY_ID=AKIAEXAMPLE
export AWS_SECRET_ACCESS_KEY=secret
AWS_SESSION_TOKEN=token
`
	sess, err := ProfileSessionFromEnvOutput(stdout)
	if err != nil {
		t.Fatal(err)
	}
	if !sess.complete() || sess.AccessKeyID != "AKIAEXAMPLE" {
		t.Fatalf("session: %+v", sess)
	}
}

func TestProfileSessionFromEnvOutputIncomplete(t *testing.T) {
	_, err := ProfileSessionFromEnvOutput("AWS_REGION=us-east-1\n")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSessionFromEnvMapDefaultRegion(t *testing.T) {
	sess := SessionFromEnvMap(map[string]string{
		"AWS_ACCESS_KEY_ID":     "a",
		"AWS_SECRET_ACCESS_KEY": "b",
		"AWS_SESSION_TOKEN":     "c",
	})
	if sess.Region != "us-east-1" {
		t.Fatalf("region = %q", sess.Region)
	}
}
