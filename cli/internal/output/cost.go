package output

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/openshift-online/finops-tools/core/cost"
)

// Format identifies how cost results are written.
type Format string

const (
	FormatPrettyPrint Format = "pretty-print"
	FormatJSON        Format = "json"
	FormatCSV         Format = "csv"
)

// ParseFormat validates a --format flag value (case-insensitive).
func ParseFormat(s string) (Format, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case string(FormatPrettyPrint), "":
		return FormatPrettyPrint, nil
	case string(FormatJSON):
		return FormatJSON, nil
	case string(FormatCSV):
		return FormatCSV, nil
	default:
		return "", fmt.Errorf("unknown format %q (supported: pretty-print, json, csv)", s)
	}
}

// WriteCostResult writes a cost summary in the requested format.
func WriteCostResult(w io.Writer, format Format, r cost.CostResult) error {
	switch format {
	case FormatPrettyPrint:
		return writePrettyPrint(w, r)
	case FormatJSON:
		return writeJSON(w, r)
	case FormatCSV:
		return writeCSV(w, r)
	default:
		return fmt.Errorf("unknown format %q", format)
	}
}

func writeJSON(w io.Writer, r cost.CostResult) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

func writeCSV(w io.Writer, r cost.CostResult) error {
	cw := csv.NewWriter(w)
	defer cw.Flush()

	if len(r.Breakdown) > 0 {
		dimCol := breakdownCSVColumn(r.SplitBy)
		header := []string{
			"provider", "account_name", "account_id", "metric",
			"currency", "start_date", "end_date", dimCol, "amount",
		}
		if err := cw.Write(header); err != nil {
			return err
		}
		for _, item := range r.Breakdown {
			if err := cw.Write(csvBreakdownRow(r, item)); err != nil {
				return err
			}
		}
		return cw.Error()
	}

	header := []string{
		"provider", "account_name", "account_id", "metric",
		"currency", "amount", "start_date", "end_date",
	}
	if err := cw.Write(header); err != nil {
		return err
	}
	row := []string{
		string(r.Provider),
		r.AccountName,
		r.AccountID,
		r.Metric,
		r.Currency,
		fmt.Sprintf("%.10f", r.Amount),
		r.StartDate,
		r.EndDate,
	}
	if err := cw.Write(row); err != nil {
		return err
	}
	return cw.Error()
}

func breakdownCSVColumn(splitBy cost.SplitBy) string {
	switch splitBy {
	case cost.SplitByAccount:
		return "linked_account_id"
	default:
		return "service"
	}
}

func csvBreakdownRow(r cost.CostResult, item cost.CostBreakdownItem) []string {
	return []string{
		string(r.Provider),
		r.AccountName,
		r.AccountID,
		r.Metric,
		r.Currency,
		r.StartDate,
		r.EndDate,
		item.DisplayLabel(r.SplitBy),
		fmt.Sprintf("%.10f", item.Amount),
	}
}

func formatAmount(amount float64) string {
	s := fmt.Sprintf("%.2f", amount)
	parts := strings.Split(s, ".")
	intPart := parts[0]
	neg := strings.HasPrefix(intPart, "-")
	if neg {
		intPart = intPart[1:]
	}
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	for i, c := range intPart {
		if i > 0 && (len(intPart)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(c)
	}
	b.WriteByte('.')
	if len(parts) > 1 {
		b.WriteString(parts[1])
	} else {
		b.WriteString("00")
	}
	return b.String()
}
