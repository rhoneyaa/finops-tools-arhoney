package snowflake

import (
	"database/sql"
	"testing"
	"time"
)

func TestResolveColumnNames(t *testing.T) {
	t.Parallel()

	names := resolveColumnNames(
		[]string{"", "ACCOUNT_ID"},
		[]*sql.ColumnType{nil, nil},
	)
	if names[0] != "COLUMN_1" {
		t.Fatalf("names[0] = %q, want COLUMN_1", names[0])
	}
	if names[1] != "ACCOUNT_ID" {
		t.Fatalf("names[1] = %q, want ACCOUNT_ID", names[1])
	}
}

func TestValueString(t *testing.T) {
	t.Parallel()

	ts := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	tests := []struct {
		in   any
		want string
	}{
		{nil, ""},
		{"alpha", "alpha"},
		{[]byte("beta"), "beta"},
		{true, "true"},
		{false, "false"},
		{ts, ts.Format(time.RFC3339Nano)},
		{42, "42"},
	}
	for _, tc := range tests {
		if got := valueString(tc.in); got != tc.want {
			t.Fatalf("valueString(%#v) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
