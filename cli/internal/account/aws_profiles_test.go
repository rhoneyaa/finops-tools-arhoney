// aws_profiles_test.go tests AWS profile name ordering (alias before account ID).
package account

import (
	"reflect"
	"testing"
)

func TestAWSProfileNames(t *testing.T) {
	got := awsProfileNames("123456789012", "rh-control", nil)
	want := []string{"rh-control", "123456789012"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	got = awsProfileNames("123456789012", "", nil)
	want = []string{"123456789012"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}

	custom := []string{"custom"}
	if got := awsProfileNames("123456789012", "rh-control", custom); !reflect.DeepEqual(got, custom) {
		t.Fatalf("got %v want %v", got, custom)
	}
}
