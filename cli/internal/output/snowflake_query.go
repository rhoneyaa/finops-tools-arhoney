// snowflake_query.go formats Snowflake SQL query results for terminal output.
package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	coresnowflake "github.com/openshift-online/finops-tools/core/snowflake"
)

// WriteSnowflakeQueryResult renders a Snowflake query result in the selected format.
func WriteSnowflakeQueryResult(w io.Writer, format Format, result coresnowflake.QueryResult) error {
	switch format {
	case FormatPrettyPrint:
		return writeSnowflakeQueryPretty(w, result)
	case FormatJSON:
		return writeSnowflakeQueryJSON(w, result)
	case FormatCSV:
		return writeSnowflakeQueryCSV(w, result)
	default:
		return fmt.Errorf("unknown format %q", format)
	}
}

func writeSnowflakeQueryJSON(w io.Writer, result coresnowflake.QueryResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func writeSnowflakeQueryCSV(w io.Writer, result coresnowflake.QueryResult) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	if len(result.Columns) == 0 {
		return cw.Error()
	}
	if err := cw.Write(result.Columns); err != nil {
		return err
	}
	for _, row := range result.Rows {
		vals := make([]string, len(result.Columns))
		for i := range result.Columns {
			if i < len(row) {
				vals[i] = row[i]
			}
		}
		if err := cw.Write(vals); err != nil {
			return err
		}
	}
	return cw.Error()
}
