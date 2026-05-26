// filter.go removes redundant linked-account targets when their payer is also in the request set.
package cost

// FilterOverlappingTargets drops linked accounts whose payer is also requested,
// since linked costs are already included in payer totals.
func FilterOverlappingTargets(targets []AccountTarget) []AccountTarget {
	payers := make(map[string]struct{})
	for _, t := range targets {
		if !t.IsLinked() {
			payers[t.AccountID] = struct{}{}
		}
	}

	out := make([]AccountTarget, 0, len(targets))
	for _, t := range targets {
		if t.IsLinked() {
			if _, ok := payers[t.PayerAccountID]; ok {
				continue
			}
		}
		out = append(out, t)
	}
	return out
}
