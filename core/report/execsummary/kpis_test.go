package execsummary

import (
	"testing"
	"time"
)

func TestComputePayerKPIs(t *testing.T) {
	enriched := []EnrichedCostRecord{
		// Last month (2026-02)
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "111111111111", Cost: 1000, Payer: "payer1"}},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "222222222222", Cost: 500, Payer: "payer1"}},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "333333333333", Cost: 300, Payer: "payer2"}},
		// Prev month (2026-01)
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Cost: 800, Payer: "payer1"}},
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "222222222222", Cost: 400, Payer: "payer1"}},
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "333333333333", Cost: 200, Payer: "payer2"}},
	}

	hcpAccounts := map[string]bool{
		"111111111111": true,
	}

	clusterCounts := ClusterCounts{
		"2026-01": 100,
		"2026-02": 105,
	}

	envCounts := EnvCounts{
		"total":  105,
		"prod":   50,
		"payer1": 75, // Specific payer count
	}

	anomalies := []AnomalyRecord{
		{AccountID: "111111111111", Payer: "payer1"},
		{AccountID: "333333333333", Payer: "payer2"},
	}

	lastMonth := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	prevMonth := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	kpis := ComputePayerKPIs(
		enriched,
		"payer1",
		lastMonth,
		&prevMonth,
		hcpAccounts,
		clusterCounts,
		envCounts,
		anomalies,
		nil,
	)

	// Verify totals
	if kpis.TotalLast != 1500 {
		t.Errorf("TotalLast = %.2f, want 1500", kpis.TotalLast)
	}

	if kpis.TotalPrev != 1200 {
		t.Errorf("TotalPrev = %.2f, want 1200", kpis.TotalPrev)
	}

	// Verify MoM percentage: (1500 - 1200) / 1200 * 100 = 25%
	expectedMoM := 25.0
	if kpis.MoMPct != expectedMoM {
		t.Errorf("MoMPct = %.2f, want %.2f", kpis.MoMPct, expectedMoM)
	}

	// Verify HCP cost (only account 111)
	if kpis.HCPCost != 1000 {
		t.Errorf("HCPCost = %.2f, want 1000", kpis.HCPCost)
	}

	// Verify HCP unit cost: 1000 / 75 (payer1 clusters) ≈ 13.33
	if kpis.HCPUnit == nil {
		t.Fatal("HCPUnit should not be nil")
	}
	expectedUnit := 1000.0 / 75.0
	if *kpis.HCPUnit != expectedUnit {
		t.Errorf("HCPUnit = %.2f, want %.2f", *kpis.HCPUnit, expectedUnit)
	}

	// Verify anomaly count (only payer1 anomalies)
	if kpis.AnomalyCount != 1 {
		t.Errorf("AnomalyCount = %d, want 1", kpis.AnomalyCount)
	}

	if len(kpis.Anomalies) != 1 {
		t.Errorf("len(Anomalies) = %d, want 1", len(kpis.Anomalies))
	}
}

func TestComputePayerKPIsAllPayers(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "111111111111", Cost: 1000, Payer: "payer1"}},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "222222222222", Cost: 500, Payer: "payer2"}},
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Cost: 800, Payer: "payer1"}},
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "222222222222", Cost: 400, Payer: "payer2"}},
	}

	lastMonth := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	prevMonth := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	kpis := ComputePayerKPIs(
		enriched,
		"all",
		lastMonth,
		&prevMonth,
		make(map[string]bool),
		ClusterCounts{},
		EnvCounts{},
		[]AnomalyRecord{},
		nil,
	)

	// Should include all payers
	if kpis.TotalLast != 1500 {
		t.Errorf("TotalLast (all payers) = %.2f, want 1500", kpis.TotalLast)
	}

	if kpis.TotalPrev != 1200 {
		t.Errorf("TotalPrev (all payers) = %.2f, want 1200", kpis.TotalPrev)
	}
}

func TestComputePayerKPIsNoPrevMonth(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "111111111111", Cost: 1000}},
	}

	lastMonth := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)

	kpis := ComputePayerKPIs(
		enriched,
		"all",
		lastMonth,
		nil, // No prev month
		make(map[string]bool),
		ClusterCounts{},
		EnvCounts{},
		[]AnomalyRecord{},
		nil,
	)

	if kpis.TotalPrev != 0 {
		t.Errorf("TotalPrev = %.2f, want 0", kpis.TotalPrev)
	}

	if kpis.MoMPct != 0 {
		t.Errorf("MoMPct = %.2f, want 0 (no prev month)", kpis.MoMPct)
	}

	if kpis.ClusterDelta != nil {
		t.Error("ClusterDelta should be nil when no prev month")
	}
}

func TestMonthCostTotal(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Month: "2026-01", Cost: 100}},
		{CostRecord: CostRecord{Month: "2026-01", Cost: 200}},
		{CostRecord: CostRecord{Month: "2026-02", Cost: 300}},
	}

	total := monthCostTotal(enriched, "2026-01")
	if total != 300 {
		t.Errorf("monthCostTotal(2026-01) = %.2f, want 300", total)
	}

	total2 := monthCostTotal(enriched, "2026-02")
	if total2 != 300 {
		t.Errorf("monthCostTotal(2026-02) = %.2f, want 300", total2)
	}

	total3 := monthCostTotal(enriched, "2026-03")
	if total3 != 0 {
		t.Errorf("monthCostTotal(2026-03) = %.2f, want 0 (no data)", total3)
	}
}

func TestMonthCostTotalEmpty(t *testing.T) {
	total := monthCostTotal([]EnrichedCostRecord{}, "2026-01")
	if total != 0 {
		t.Errorf("monthCostTotal(empty) = %.2f, want 0", total)
	}

	total2 := monthCostTotal(nil, "2026-01")
	if total2 != 0 {
		t.Errorf("monthCostTotal(nil) = %.2f, want 0", total2)
	}
}

func TestHCPMonthCost(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "111111111111", Cost: 500}},
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "222222222222", Cost: 300}},
		{CostRecord: CostRecord{Month: "2026-01", AccountID: "333333333333", Cost: 200}},
		{CostRecord: CostRecord{Month: "2026-02", AccountID: "111111111111", Cost: 600}},
	}

	hcpAccounts := map[string]bool{
		"111111111111": true,
		"222222222222": true,
	}

	hcpCost := hcpMonthCost(enriched, "2026-01", hcpAccounts)
	// Should be 500 + 300 = 800
	if hcpCost != 800 {
		t.Errorf("hcpMonthCost(2026-01) = %.2f, want 800", hcpCost)
	}

	hcpCost2 := hcpMonthCost(enriched, "2026-02", hcpAccounts)
	// Should be 600 (only account 111 in Feb)
	if hcpCost2 != 600 {
		t.Errorf("hcpMonthCost(2026-02) = %.2f, want 600", hcpCost2)
	}
}

func TestComputeClusterMetricsAllPayers(t *testing.T) {
	clusterCounts := ClusterCounts{
		"2026-01": 100,
		"2026-02": 105,
	}

	envCounts := EnvCounts{
		"total": 110, // Live count overrides historical
		"prod":  55,
	}

	prevMonth := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

	count, delta, unitDenom := computeClusterMetrics(
		"all",
		"2026-02",
		"2026-01",
		clusterCounts,
		envCounts,
		&prevMonth,
	)

	// Current should be live total (110)
	if count != 110 {
		t.Errorf("count = %d, want 110 (live total)", count)
	}

	// Delta: 110 - 100 = 10
	if delta == nil {
		t.Fatal("delta should not be nil")
	}
	if *delta != 10 {
		t.Errorf("delta = %d, want 10", *delta)
	}

	// Unit denom should be prod count
	if unitDenom != 55 {
		t.Errorf("unitDenom = %d, want 55 (prod count)", unitDenom)
	}
}

func TestComputeClusterMetricsSpecificPayer(t *testing.T) {
	envCounts := EnvCounts{
		"payer1": 25,
	}

	count, delta, unitDenom := computeClusterMetrics(
		"payer1",
		"2026-02",
		"2026-01",
		ClusterCounts{},
		envCounts,
		nil,
	)

	if count != 25 {
		t.Errorf("count = %d, want 25", count)
	}

	if delta != nil {
		t.Error("delta should be nil for specific payer")
	}

	if unitDenom != 25 {
		t.Errorf("unitDenom = %d, want 25 (same as count)", unitDenom)
	}
}

func TestComputeClusterMetricsNoLiveData(t *testing.T) {
	clusterCounts := ClusterCounts{
		"2026-02": 100,
	}

	envCounts := EnvCounts{
		"total": 0, // No live data
	}

	count, _, _ := computeClusterMetrics(
		"all",
		"2026-02",
		"",
		clusterCounts,
		envCounts,
		nil,
	)

	// Should fall back to historical
	if count != 100 {
		t.Errorf("count = %d, want 100 (historical fallback)", count)
	}
}

func TestResolvePayerSlice(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Payer: "payer1", AccountID: "111111111111"}},
		{CostRecord: CostRecord{Payer: "payer1", AccountID: "222222222222"}},
		{CostRecord: CostRecord{Payer: "payer2", AccountID: "333333333333"}},
	}

	anomalies := []AnomalyRecord{
		{AccountID: "111111111111", Payer: "payer1"},
		{AccountID: "333333333333", Payer: "payer2"},
	}

	spData := &SPData{
		SPByPayer: map[string]PayerSPData{
			"payer1": {
				Coverage:    []SavingsPlanMetric{{Month: "2026-01", Coverage: 85.0}},
				Utilization: []SavingsPlanMetric{{Month: "2026-01", Utilization: 92.0}},
			},
		},
	}

	payerEnr, payerAnom, cov, util := resolvePayerSlice(enriched, "payer1", anomalies, spData)

	// Should have 2 enriched records for payer1
	if len(payerEnr) != 2 {
		t.Errorf("len(payerEnr) = %d, want 2", len(payerEnr))
	}

	// Should have 1 anomaly for payer1
	if len(payerAnom) != 1 {
		t.Errorf("len(payerAnom) = %d, want 1", len(payerAnom))
	}

	// Should have SP data
	if len(cov) != 1 {
		t.Errorf("len(coverage) = %d, want 1", len(cov))
	}

	if len(util) != 1 {
		t.Errorf("len(utilization) = %d, want 1", len(util))
	}
}

func TestResolvePayerSliceAll(t *testing.T) {
	enriched := []EnrichedCostRecord{
		{CostRecord: CostRecord{Payer: "payer1"}},
		{CostRecord: CostRecord{Payer: "payer2"}},
	}

	anomalies := []AnomalyRecord{
		{Payer: "payer1"},
		{Payer: "payer2"},
	}

	payerEnr, payerAnom, cov, util := resolvePayerSlice(enriched, "all", anomalies, nil)

	// Should return all data
	if len(payerEnr) != 2 {
		t.Errorf("len(payerEnr) = %d, want 2 (all)", len(payerEnr))
	}

	if len(payerAnom) != 2 {
		t.Errorf("len(payerAnom) = %d, want 2 (all)", len(payerAnom))
	}

	if cov != nil {
		t.Error("coverage should be nil for 'all'")
	}

	if util != nil {
		t.Error("utilization should be nil for 'all'")
	}
}

func TestGetEnvCountInt(t *testing.T) {
	envCounts := EnvCounts{
		"total": 100,
		"prod":  50,
		"month": "2026-01", // String value, not int
	}

	if getEnvCountInt(envCounts, "total") != 100 {
		t.Errorf("getEnvCountInt(total) = %d, want 100", getEnvCountInt(envCounts, "total"))
	}

	if getEnvCountInt(envCounts, "prod") != 50 {
		t.Errorf("getEnvCountInt(prod) = %d, want 50", getEnvCountInt(envCounts, "prod"))
	}

	if getEnvCountInt(envCounts, "month") != 0 {
		t.Errorf("getEnvCountInt(month) = %d, want 0 (not an int)", getEnvCountInt(envCounts, "month"))
	}

	if getEnvCountInt(envCounts, "missing") != 0 {
		t.Errorf("getEnvCountInt(missing) = %d, want 0", getEnvCountInt(envCounts, "missing"))
	}
}
