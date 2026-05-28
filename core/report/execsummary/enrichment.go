package execsummary

import (
	"sort"
	"time"
)

// EnrichWithMapping performs a left-join of cost data with account mappings.
// Returns a new slice of EnrichedCostRecord with normalized display fields.
func EnrichWithMapping(costData *CostData, mappings []AccountMapping) []EnrichedCostRecord {
	if len(costData.Records) == 0 {
		return nil
	}

	// Build lookup map: account_id -> AccountMapping
	mappingLookup := make(map[string]AccountMapping)
	for _, m := range mappings {
		mappingLookup[m.AccountID] = m
	}

	enriched := make([]EnrichedCostRecord, len(costData.Records))
	for i, rec := range costData.Records {
		enriched[i] = EnrichedCostRecord{
			CostRecord: rec,
		}

		// Left join: lookup mapping, use defaults if not found
		if mapping, found := mappingLookup[rec.AccountID]; found {
			enriched[i].RefinedCategory = normalizeCategory(mapping.RefinedCategory)
			enriched[i].SubType = mapping.SubType
			enriched[i].OwnerTeam = mapping.OwnerTeam
			enriched[i].AccountName = normalizeAccountName(mapping.AccountName, rec.AccountID)
		} else {
			// Unmapped account
			enriched[i].RefinedCategory = "Unmapped"
			enriched[i].SubType = ""
			enriched[i].OwnerTeam = ""
			enriched[i].AccountName = rec.AccountID
		}
	}

	return enriched
}

// CategoryMonthly aggregates cost by Refined Category and month.
// Only includes months in the provided time window.
func CategoryMonthly(enriched []EnrichedCostRecord, months []time.Time) []CategoryMonthRecord {
	if len(enriched) == 0 {
		return nil
	}

	// Build set of month labels
	monthSet := make(map[string]bool)
	for _, m := range months {
		monthSet[MonthLabel(m)] = true
	}

	// Aggregate by category + month
	aggregates := make(map[string]map[string]float64) // category -> month -> cost
	for _, rec := range enriched {
		if !monthSet[rec.Month] {
			continue // Skip months outside the window
		}

		category := rec.RefinedCategory
		if aggregates[category] == nil {
			aggregates[category] = make(map[string]float64)
		}
		aggregates[category][rec.Month] += rec.Cost
	}

	// Convert to slice
	var results []CategoryMonthRecord
	for category, monthCosts := range aggregates {
		for month, cost := range monthCosts {
			results = append(results, CategoryMonthRecord{
				RefinedCategory: category,
				Month:           month,
				Cost:            cost,
			})
		}
	}

	// Sort by category, then month
	sort.Slice(results, func(i, j int) bool {
		if results[i].RefinedCategory != results[j].RefinedCategory {
			return results[i].RefinedCategory < results[j].RefinedCategory
		}
		return results[i].Month < results[j].Month
	})

	return results
}

// AccountMonthlyTotal aggregates cost by account and month.
// Returns map[account_id]map[month]cost.
func AccountMonthlyTotal(enriched []EnrichedCostRecord) map[string]map[string]float64 {
	aggregates := make(map[string]map[string]float64)

	for _, rec := range enriched {
		if aggregates[rec.AccountID] == nil {
			aggregates[rec.AccountID] = make(map[string]float64)
		}
		aggregates[rec.AccountID][rec.Month] += rec.Cost
	}

	return aggregates
}

// ServiceCostsForAccount returns top N services for a specific account/month.
func ServiceCostsForAccount(enriched []EnrichedCostRecord, accountID, month string, topN int) []ServiceCostPair {
	serviceCosts := make(map[string]float64)

	for _, rec := range enriched {
		if rec.AccountID == accountID && rec.Month == month && rec.Service != "" {
			serviceCosts[rec.Service] += rec.Cost
		}
	}

	// Convert to slice
	var pairs []ServiceCostPair
	for service, cost := range serviceCosts {
		pairs = append(pairs, ServiceCostPair{
			Service: service,
			Cost:    cost,
		})
	}

	// Sort descending by cost
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Cost > pairs[j].Cost
	})

	// Return top N
	if len(pairs) > topN {
		pairs = pairs[:topN]
	}

	return pairs
}

// FilterByMonth returns enriched records matching a specific month.
func FilterByMonth(enriched []EnrichedCostRecord, month string) []EnrichedCostRecord {
	var filtered []EnrichedCostRecord
	for _, rec := range enriched {
		if rec.Month == month {
			filtered = append(filtered, rec)
		}
	}
	return filtered
}

// FilterByPayer returns enriched records matching a specific payer.
func FilterByPayer(enriched []EnrichedCostRecord, payer string) []EnrichedCostRecord {
	if payer == "" || payer == "all" {
		return enriched
	}

	var filtered []EnrichedCostRecord
	for _, rec := range enriched {
		if rec.Payer == payer {
			filtered = append(filtered, rec)
		}
	}
	return filtered
}

// FilterByAccounts returns enriched records matching a set of account IDs.
func FilterByAccounts(enriched []EnrichedCostRecord, accountIDs map[string]bool) []EnrichedCostRecord {
	if len(accountIDs) == 0 {
		return enriched
	}

	var filtered []EnrichedCostRecord
	for _, rec := range enriched {
		if accountIDs[rec.AccountID] {
			filtered = append(filtered, rec)
		}
	}
	return filtered
}

// TotalCost sums cost across all records.
func TotalCost(enriched []EnrichedCostRecord) float64 {
	var total float64
	for _, rec := range enriched {
		total += rec.Cost
	}
	return total
}

// normalizeCategory ensures non-empty category, defaulting to "Unmapped".
func normalizeCategory(category string) string {
	if category == "" {
		return "Unmapped"
	}
	return category
}

// normalizeAccountName uses account name if present, else falls back to account ID.
func normalizeAccountName(name, accountID string) string {
	if name == "" {
		return accountID
	}
	return name
}
