package execsummary

import (
	"math"
	"sort"
	"time"
)

// TopGrowingAccounts returns accounts with the largest month-over-month cost growth.
// Compares the last two months in the time window.
func TopGrowingAccounts(enriched []EnrichedCostRecord, months []time.Time, topN, topSvcs int) []GrowingAccountRecord {
	if len(enriched) == 0 || len(months) < 2 {
		return nil
	}

	lastMonth := MonthLabel(months[len(months)-1])
	prevMonth := MonthLabel(months[len(months)-2])

	// Calculate account cost deltas
	deltas, lastByAcct, prevByAcct := accountCostDeltas(enriched, lastMonth, prevMonth)

	// Sort by delta descending
	type deltaPair struct {
		accountID string
		delta     float64
	}
	var pairs []deltaPair
	for accountID, delta := range deltas {
		pairs = append(pairs, deltaPair{accountID, delta})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].delta > pairs[j].delta
	})

	// Take top N
	if len(pairs) > topN {
		pairs = pairs[:topN]
	}

	// Build result records
	var results []GrowingAccountRecord
	for _, pair := range pairs {
		metadata := getAccountMetadata(enriched, pair.accountID)
		topServices := ServiceCostsForAccount(enriched, pair.accountID, lastMonth, topSvcs)

		results = append(results, GrowingAccountRecord{
			AccountID:     pair.accountID,
			AccountName:   metadata.AccountName,
			Category:      metadata.Category,
			Owner:         metadata.Owner,
			LastMonthCost: lastByAcct[pair.accountID],
			PrevMonthCost: prevByAcct[pair.accountID],
			Delta:         pair.delta,
			TopServices:   topServices,
		})
	}

	return results
}

// StatisticalAnomalies detects account-level anomalies using z-score analysis.
// Returns up to 15 anomalies sorted by absolute z-score.
func StatisticalAnomalies(enriched []EnrichedCostRecord, months []time.Time, zThreshold float64) []AnomalyRecord {
	if len(enriched) == 0 || len(months) < 3 {
		return nil
	}

	lastMonth := MonthLabel(months[len(months)-1])

	// Aggregate cost by account and month
	accountMonthly := aggregateByAccountMonth(enriched)

	var anomalies []AnomalyRecord

	// Check each account for anomalies
	for accountID, monthlyCosts := range accountMonthly {
		anomaly := detectAccountAnomaly(enriched, accountID, monthlyCosts, lastMonth, zThreshold)
		if anomaly != nil {
			anomalies = append(anomalies, *anomaly)
		}
	}

	// Sort by absolute z-score descending
	sort.Slice(anomalies, func(i, j int) bool {
		return math.Abs(anomalies[i].ZScore) > math.Abs(anomalies[j].ZScore)
	})

	// Return top 15
	if len(anomalies) > 15 {
		anomalies = anomalies[:15]
	}

	return anomalies
}

// accountCostDeltas calculates month-over-month deltas for all accounts.
// Returns: deltas map, last month totals, prev month totals.
func accountCostDeltas(enriched []EnrichedCostRecord, lastMonth, prevMonth string) (map[string]float64, map[string]float64, map[string]float64) {
	lastByAcct := make(map[string]float64)
	prevByAcct := make(map[string]float64)

	for _, rec := range enriched {
		if rec.Month == lastMonth {
			lastByAcct[rec.AccountID] += rec.Cost
		} else if rec.Month == prevMonth {
			prevByAcct[rec.AccountID] += rec.Cost
		}
	}

	// Calculate deltas
	deltas := make(map[string]float64)
	allAccounts := make(map[string]bool)
	for acct := range lastByAcct {
		allAccounts[acct] = true
	}
	for acct := range prevByAcct {
		allAccounts[acct] = true
	}

	for acct := range allAccounts {
		last := lastByAcct[acct]
		prev := prevByAcct[acct]
		deltas[acct] = last - prev
	}

	return deltas, lastByAcct, prevByAcct
}

// accountMeta holds display metadata for an account.
type accountMeta struct {
	AccountName string
	Category    string
	Owner       string
}

func getAccountMetadata(enriched []EnrichedCostRecord, accountID string) accountMeta {
	// Find first record for this account
	for _, rec := range enriched {
		if rec.AccountID == accountID {
			return accountMeta{
				AccountName: rec.AccountName,
				Category:    rec.RefinedCategory,
				Owner:       rec.OwnerTeam,
			}
		}
	}

	// Not found - return defaults
	return accountMeta{
		AccountName: accountID,
		Category:    "Unmapped",
		Owner:       "",
	}
}

// aggregateByAccountMonth groups costs by account and month.
func aggregateByAccountMonth(enriched []EnrichedCostRecord) map[string]map[string]float64 {
	result := make(map[string]map[string]float64)

	for _, rec := range enriched {
		if result[rec.AccountID] == nil {
			result[rec.AccountID] = make(map[string]float64)
		}
		result[rec.AccountID][rec.Month] += rec.Cost
	}

	return result
}

// detectAccountAnomaly checks if an account has a statistical anomaly in the last month.
func detectAccountAnomaly(enriched []EnrichedCostRecord, accountID string, monthlyCosts map[string]float64, lastMonth string, zThreshold float64) *AnomalyRecord {
	// Get prior months (all months before last month)
	var priorCosts []float64
	for month, cost := range monthlyCosts {
		if month < lastMonth {
			priorCosts = append(priorCosts, cost)
		}
	}

	// Need at least 2 prior months for statistics
	if len(priorCosts) < 2 {
		return nil
	}

	// Calculate mean and standard deviation
	meanCost := mean(priorCosts)
	stdCost := stddev(priorCosts, meanCost)

	// Skip if standard deviation is too low (< $10)
	if stdCost < 10 {
		return nil
	}

	// Get current month cost
	currentCost, exists := monthlyCosts[lastMonth]
	if !exists {
		return nil
	}

	// Calculate z-score
	zScore := (currentCost - meanCost) / stdCost

	// Check if it breaches threshold
	if math.Abs(zScore) < zThreshold {
		return nil
	}

	// Get account metadata
	meta := getAccountMetadata(enriched, accountID)

	// Calculate percentage change
	var pctChange float64
	if meanCost > 0 {
		pctChange = math.Round((currentCost-meanCost)/meanCost*100*10) / 10 // Round to 1 decimal
	}

	direction := "spike"
	if zScore < 0 {
		direction = "drop"
	}

	// Get payer if available
	var payer string
	for _, rec := range enriched {
		if rec.AccountID == accountID {
			payer = rec.Payer
			break
		}
	}

	return &AnomalyRecord{
		AccountID:   accountID,
		AccountName: meta.AccountName,
		Category:    meta.Category,
		Service:     "All Services",
		Month:       lastMonth,
		CurrentCost: currentCost,
		MeanCost:    meanCost,
		ZScore:      math.Round(zScore*10) / 10, // Round to 1 decimal
		PctChange:   pctChange,
		Direction:   direction,
		Payer:       payer,
	}
}

// Statistical helper functions

// mean calculates the arithmetic mean of a slice of floats.
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}

	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// stddev calculates the sample standard deviation.
func stddev(values []float64, mean float64) float64 {
	if len(values) < 2 {
		return 0
	}

	variance := 0.0
	for _, v := range values {
		diff := v - mean
		variance += diff * diff
	}
	variance /= float64(len(values) - 1) // Sample variance (n-1)

	return math.Sqrt(variance)
}
