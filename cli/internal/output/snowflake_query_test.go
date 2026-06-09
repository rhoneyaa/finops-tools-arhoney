package output

import (
	"encoding/json"
	"strings"
	"testing"

	coresnowflake "github.com/openshift-online/finops-tools/core/snowflake"
)

func fixtureSnowflakeQueryResult() coresnowflake.QueryResult {
	return coresnowflake.QueryResult{
		Columns: []string{"USER", "ROLE"},
		Rows: [][]string{
			{"alice", "ANALYST"},
			{"bob", "ADMIN"},
		},
	}
}

func TestWriteSnowflakeQueryPretty(t *testing.T) {
	var buf strings.Builder
	if err := WriteSnowflakeQueryResult(&buf, FormatPrettyPrint, fixtureSnowflakeQueryResult()); err != nil {
		t.Fatal(err)
	}
	out := stripANSI(buf.String())
	for _, want := range []string{
		"2 rows",
		"Query results",
		"USER",
		"ROLE",
		"alice",
		"ANALYST",
		"bob",
		"ADMIN",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestWriteSnowflakeQueryPrettyEmptyRowsShowsColumns(t *testing.T) {
	result := coresnowflake.QueryResult{
		Columns: []string{"USER", "ROLE"},
	}
	var buf strings.Builder
	if err := WriteSnowflakeQueryResult(&buf, FormatPrettyPrint, result); err != nil {
		t.Fatal(err)
	}
	out := stripANSI(buf.String())
	for _, want := range []string{
		"0 rows",
		"USER",
		"ROLE",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}

func TestWriteSnowflakeQueryPrettyNullCell(t *testing.T) {
	result := coresnowflake.QueryResult{
		Columns: []string{"A"},
		Rows:    [][]string{{""}},
	}
	var buf strings.Builder
	if err := WriteSnowflakeQueryResult(&buf, FormatPrettyPrint, result); err != nil {
		t.Fatal(err)
	}
	out := stripANSI(buf.String())
	if !strings.Contains(out, "-") {
		t.Fatalf("expected placeholder for empty cell, got:\n%s", out)
	}
}

func TestWriteSnowflakeQueryJSON(t *testing.T) {
	var buf strings.Builder
	if err := WriteSnowflakeQueryResult(&buf, FormatJSON, fixtureSnowflakeQueryResult()); err != nil {
		t.Fatal(err)
	}
	var decoded coresnowflake.QueryResult
	if err := json.Unmarshal([]byte(buf.String()), &decoded); err != nil {
		t.Fatal(err)
	}
	if len(decoded.Columns) != 2 || len(decoded.Rows) != 2 {
		t.Fatalf("decoded = %+v", decoded)
	}
}

func TestWriteSnowflakeQueryCSV(t *testing.T) {
	var buf strings.Builder
	if err := WriteSnowflakeQueryResult(&buf, FormatCSV, fixtureSnowflakeQueryResult()); err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"USER,ROLE",
		"alice,ANALYST",
		"bob,ADMIN",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %q in:\n%s", want, out)
		}
	}
}
