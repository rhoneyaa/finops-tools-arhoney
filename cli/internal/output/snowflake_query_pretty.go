// snowflake_query_pretty.go renders Snowflake query results as a colorized table.
package output

import (
	"fmt"
	"io"

	"github.com/olekukonko/tablewriter"
	coresnowflake "github.com/openshift-online/finops-tools/core/snowflake"
)

func writeSnowflakeQueryPretty(w io.Writer, result coresnowflake.QueryResult) error {
	s := newStyler(w)
	if len(result.Columns) == 0 {
		msg := "(no columns)"
		if s.enabled {
			msg = s.dim(msg)
		}
		_, err := fmt.Fprintln(w, msg)
		return err
	}

	countLabel := fmt.Sprintf("%d row", len(result.Rows))
	if len(result.Rows) != 1 {
		countLabel += "s"
	}
	if s.enabled {
		countLabel = s.dim(countLabel)
	}
	if _, err := fmt.Fprintf(w, "  %s\n\n", countLabel); err != nil {
		return err
	}

	title := "Query results"
	if s.enabled {
		title = s.bold(s.cyan(title))
	}
	if _, err := fmt.Fprintln(w, title); err != nil {
		return err
	}

	table := tablewriter.NewWriter(w)
	table.SetAutoWrapText(false)
	table.SetBorder(false)
	table.SetAutoFormatHeaders(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetTablePadding("\t")

	headers := make([]string, len(result.Columns))
	for i, col := range result.Columns {
		headers[i] = cell(s, s.bold, col)
	}
	table.SetHeader(headers)

	alignments := make([]int, len(result.Columns))
	for i := range alignments {
		alignments[i] = tablewriter.ALIGN_LEFT
	}
	table.SetColumnAlignment(alignments)

	for _, row := range result.Rows {
		vals := make([]string, len(result.Columns))
		for i := range result.Columns {
			val := ""
			if i < len(row) {
				val = row[i]
			}
			vals[i] = snowflakeCellValue(s, val)
		}
		table.Append(vals)
	}
	table.Render()
	return nil
}

func snowflakeCellValue(s styler, val string) string {
	if val == "" {
		return cell(s, s.dim, "-")
	}
	return val
}
