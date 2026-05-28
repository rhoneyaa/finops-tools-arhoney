package execsummary

import "time"

// ComputePayerKPIs calculates all key performance indicators for one payer or the all-payers aggregate.
func ComputePayerKPIs(
	enriched []EnrichedCostRecord,
	payerLabel string,
	lastMonth time.Time,
	prevMonth *time.Time,
	hcpAccounts map[string]bool,
	clusterCounts ClusterCounts,
	envCounts EnvCounts,
	anomalies []AnomalyRecord,
	spData *SPData,
) *PayerKPIs {
	// Slice data by payer
	payerEnriched, payerAnom, payerSPCov, payerSPUtil := resolvePayerSlice(
		enriched,
		payerLabel,
		anomalies,
		spData,
	)

	lastLabel := MonthLabel(lastMonth)
	var prevLabel string
	if prevMonth != nil {
		prevLabel = MonthLabel(*prevMonth)
	}

	// Calculate cost totals
	totalLast := monthCostTotal(payerEnriched, lastLabel)
	totalPrev := monthCostTotal(payerEnriched, prevLabel)

	// Month-over-month percentage
	var momPct float64
	if totalPrev > 0 {
		momPct = ((totalLast - totalPrev) / totalPrev) * 100
	}

	// HCP cost
	hcpCost := hcpMonthCost(payerEnriched, lastLabel, hcpAccounts)

	// Cluster metrics
	clusterCount, clusterDelta, unitDenom := computeClusterMetrics(
		payerLabel,
		lastLabel,
		prevLabel,
		clusterCounts,
		envCounts,
		prevMonth,
	)

	// HCP unit cost (cost per production cluster)
	var hcpUnit *float64
	if unitDenom > 0 {
		unit := hcpCost / float64(unitDenom)
		hcpUnit = &unit
	}

	return &PayerKPIs{
		TotalLast:     totalLast,
		TotalPrev:     totalPrev,
		MoMPct:        momPct,
		HCPCost:       hcpCost,
		HCPUnit:       hcpUnit,
		ClusterCount:  clusterCount,
		ClusterDelta:  clusterDelta,
		AnomalyCount:  len(payerAnom),
		Anomalies:     payerAnom,
		SPCoverage:    payerSPCov,
		SPUtilization: payerSPUtil,
		CostData:      buildCostDataFromEnriched(payerEnriched),
	}
}

// resolvePayerSlice filters enriched data, anomalies, and SP metrics by payer.
// Returns: enriched data, anomalies, SP coverage, SP utilization for the payer.
func resolvePayerSlice(
	enriched []EnrichedCostRecord,
	payerLabel string,
	anomalies []AnomalyRecord,
	spData *SPData,
) ([]EnrichedCostRecord, []AnomalyRecord, []SavingsPlanMetric, []SavingsPlanMetric) {
	// "all" means return everything
	if payerLabel == "all" {
		return enriched, anomalies, nil, nil
	}

	// Filter enriched records by payer
	payerEnriched := FilterByPayer(enriched, payerLabel)

	// Filter anomalies by payer
	var payerAnom []AnomalyRecord
	for _, anom := range anomalies {
		if anom.Payer == payerLabel {
			payerAnom = append(payerAnom, anom)
		}
	}

	// Extract payer-specific SP data
	var payerSPCov, payerSPUtil []SavingsPlanMetric
	if spData != nil {
		if payerSP, ok := spData.SPByPayer[payerLabel]; ok {
			payerSPCov = payerSP.Coverage
			payerSPUtil = payerSP.Utilization
		}
	}

	return payerEnriched, payerAnom, payerSPCov, payerSPUtil
}

// monthCostTotal returns the total cost for a specific month.
func monthCostTotal(enriched []EnrichedCostRecord, monthKey string) float64 {
	if monthKey == "" || len(enriched) == 0 {
		return 0.0
	}

	total := 0.0
	for _, rec := range enriched {
		if rec.Month == monthKey {
			total += rec.Cost
		}
	}
	return total
}

// hcpMonthCost returns the total HCP account cost for a specific month.
func hcpMonthCost(enriched []EnrichedCostRecord, monthKey string, hcpAccounts map[string]bool) float64 {
	if len(enriched) == 0 || len(hcpAccounts) == 0 {
		return 0.0
	}

	total := 0.0
	for _, rec := range enriched {
		if rec.Month == monthKey && hcpAccounts[rec.AccountID] {
			total += rec.Cost
		}
	}
	return total
}

// computeClusterMetrics calculates cluster count, month-over-month delta, and unit denominator.
// Returns: (current_count, delta, unit_denom)
func computeClusterMetrics(
	payerLabel string,
	monthKey string,
	prevMonthKey string,
	clusterCounts ClusterCounts,
	envCounts EnvCounts,
	prevMonth *time.Time,
) (int, *int, int) {
	// For non-"all" payers, use environment counts only
	if payerLabel != "all" {
		current := getEnvCountInt(envCounts, payerLabel)
		return current, nil, current
	}

	// For "all" payers, prefer live total, fallback to historical
	liveTotal := getEnvCountInt(envCounts, "total")
	historicalTotal := clusterCounts[monthKey]
	current := liveTotal
	if liveTotal == 0 {
		current = historicalTotal
	}

	// Calculate delta if prev month available
	var delta *int
	if prevMonth != nil && prevMonthKey != "" {
		if prevTotal, ok := clusterCounts[prevMonthKey]; ok {
			d := current - prevTotal
			delta = &d
		}
	}

	// Unit denominator: prod count, or total if prod not available
	unitDenom := getEnvCountInt(envCounts, "prod")
	if unitDenom == 0 {
		unitDenom = current
	}

	return current, delta, unitDenom
}

// getEnvCountInt safely extracts an int from EnvCounts (which holds interface{}).
func getEnvCountInt(envCounts EnvCounts, key string) int {
	if val, ok := envCounts[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
	}
	return 0
}

// buildCostDataFromEnriched converts enriched records back to indexed CostData.
func buildCostDataFromEnriched(enriched []EnrichedCostRecord) *CostData {
	if len(enriched) == 0 {
		return &CostData{
			Records:        []CostRecord{},
			ByMonth:        make(map[string][]CostRecord),
			ByAccount:      make(map[string][]CostRecord),
			ByMonthAccount: make(map[string]map[string]float64),
		}
	}

	records := make([]CostRecord, len(enriched))
	for i, rec := range enriched {
		records[i] = rec.CostRecord
	}

	// Use existing indexing function
	data := &CostData{
		Records:        records,
		ByMonth:        make(map[string][]CostRecord),
		ByAccount:      make(map[string][]CostRecord),
		ByMonthAccount: make(map[string]map[string]float64),
	}

	for _, rec := range records {
		data.ByMonth[rec.Month] = append(data.ByMonth[rec.Month], rec)
		data.ByAccount[rec.AccountID] = append(data.ByAccount[rec.AccountID], rec)

		if data.ByMonthAccount[rec.Month] == nil {
			data.ByMonthAccount[rec.Month] = make(map[string]float64)
		}
		data.ByMonthAccount[rec.Month][rec.AccountID] += rec.Cost
	}

	return data
}
