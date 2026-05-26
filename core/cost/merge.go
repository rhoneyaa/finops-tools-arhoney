// merge.go combines per-account CostResult values into a single summary with merged breakdown rows.
package cost

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// MergeResults combines per-account cost results into one summary.
func MergeResults(results []CostResult) (CostResult, error) {
	if len(results) == 0 {
		return CostResult{}, errors.New("no results to merge")
	}
	if len(results) == 1 {
		return results[0], nil
	}

	first := results[0]
	out := CostResult{
		Provider:  first.Provider,
		Metric:    first.Metric,
		SplitBy:   first.SplitBy,
		StartDate: first.StartDate,
		EndDate:   first.EndDate,
		Currency:  first.Currency,
	}

	var (
		names []string
		ids   []string
	)
	byKey := make(map[string]float64)

	for _, r := range results {
		if r.Provider != out.Provider {
			return CostResult{}, fmt.Errorf("cannot merge different providers")
		}
		if r.Currency != out.Currency {
			return CostResult{}, fmt.Errorf("cannot merge accounts with different currencies (%s vs %s)",
				out.Currency, r.Currency)
		}
		if r.StartDate != out.StartDate || r.EndDate != out.EndDate {
			return CostResult{}, fmt.Errorf("cannot merge accounts with different periods")
		}
		if r.SplitBy != out.SplitBy {
			return CostResult{}, fmt.Errorf("cannot merge accounts with different split-by settings")
		}

		names = append(names, r.AccountName)
		ids = append(ids, r.AccountID)
		if r.Linked {
			out.Linked = true
		}
		out.Amount += r.Amount

		for _, item := range r.Breakdown {
			byKey[item.Label(out.SplitBy)] += item.Amount
		}
	}

	out.AccountName = strings.Join(names, ", ")
	out.AccountID = strings.Join(ids, ", ")

	if len(byKey) > 0 {
		out.Breakdown = make([]CostBreakdownItem, 0, len(byKey))
		for key, amt := range byKey {
			if amt == 0 {
				continue
			}
			item := CostBreakdownItem{Amount: amt}
			switch out.SplitBy {
			case SplitByAccount:
				item.Account = key
			default:
				item.Service = key
			}
			out.Breakdown = append(out.Breakdown, item)
		}
		sort.Slice(out.Breakdown, func(i, j int) bool {
			return out.Breakdown[i].Amount > out.Breakdown[j].Amount
		})
	}

	return out, nil
}
