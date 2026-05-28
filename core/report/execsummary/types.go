// Package execsummary provides executive summary report data assembly and analytics.
package execsummary

import "time"

// CostRecord represents a single cost data point from AWS Cost Explorer.
type CostRecord struct {
	Month     string  `json:"month"`      // YYYY-MM format
	AccountID string  `json:"account_id"` // AWS account ID (12 digits)
	Cost      float64 `json:"cost"`       // Net amortized cost
	Service   string  `json:"service,omitempty"`
	Payer     string  `json:"payer,omitempty"`
}

// CostData holds cost records with indexed lookups for efficient aggregation.
type CostData struct {
	Records []CostRecord

	// ByMonth indexes records by month key (YYYY-MM)
	ByMonth map[string][]CostRecord

	// ByAccount indexes records by account ID
	ByAccount map[string][]CostRecord

	// ByMonthAccount provides quick lookup of cost by month and account
	ByMonthAccount map[string]map[string]float64
}

// AccountMapping maps AWS account IDs to business metadata.
type AccountMapping struct {
	AccountID       string `json:"account_id"`
	RefinedCategory string `json:"refined_category"` // Business category (e.g., "HCP", "Tooling")
	SubType         string `json:"sub_type,omitempty"`
	OwnerTeam       string `json:"owner_team,omitempty"`
	AccountName     string `json:"account_name"`
}

// EnrichedCostRecord is a CostRecord merged with AccountMapping metadata.
type EnrichedCostRecord struct {
	CostRecord
	RefinedCategory string
	SubType         string
	OwnerTeam       string
	AccountName     string
}

// CategoryMonthRecord aggregates cost by refined category and month.
type CategoryMonthRecord struct {
	RefinedCategory string  `json:"refined_category"`
	Month           string  `json:"month"` // YYYY-MM
	Cost            float64 `json:"cost"`
}

// GrowingAccountRecord identifies an account with significant month-over-month growth.
type GrowingAccountRecord struct {
	AccountID     string             `json:"account_id"`
	AccountName   string             `json:"account_name"`
	Category      string             `json:"category"`
	Owner         string             `json:"owner"`
	LastMonthCost float64            `json:"last_month_cost"`
	PrevMonthCost float64            `json:"prev_month_cost"`
	Delta         float64            `json:"delta"`
	TopServices   []ServiceCostPair  `json:"top_services"`
}

// ServiceCostPair represents a service name and its cost.
type ServiceCostPair struct {
	Service string  `json:"service"`
	Cost    float64 `json:"cost"`
}

// AnomalyRecord identifies an account with statistical anomaly (z-score breach).
type AnomalyRecord struct {
	AccountID   string  `json:"account_id"`
	AccountName string  `json:"account_name"`
	Category    string  `json:"category"`
	Service     string  `json:"service"`
	Month       string  `json:"month"` // YYYY-MM
	CurrentCost float64 `json:"current_cost"`
	MeanCost    float64 `json:"mean_cost"`
	ZScore      float64 `json:"z_score"`
	PctChange   float64 `json:"pct_change"` // percentage
	Direction   string  `json:"direction"`  // "spike" or "drop"
	Payer       string  `json:"payer,omitempty"`
}

// SavingsPlanMetric represents savings plan coverage or utilization data.
type SavingsPlanMetric struct {
	Month      string  `json:"month"` // YYYY-MM
	Coverage   float64 `json:"coverage,omitempty"`   // percentage
	Utilization float64 `json:"utilization,omitempty"` // percentage
	Payer      string  `json:"payer,omitempty"`
}

// PayerKPIs aggregates all key performance indicators for one payer or all payers.
type PayerKPIs struct {
	TotalLast     float64               `json:"total_last"`      // Last month total cost
	TotalPrev     float64               `json:"total_prev"`      // Previous month total cost
	MoMPct        float64               `json:"mom_pct"`         // Month-over-month percentage change
	HCPCost       float64               `json:"hcp_cost"`        // HCP account cost for last month
	HCPUnit       *float64              `json:"hcp_unit,omitempty"` // Cost per production cluster
	ClusterCount  int                   `json:"cluster_count"`
	ClusterDelta  *int                  `json:"cluster_delta,omitempty"` // MoM cluster count change
	AnomalyCount  int                   `json:"anomaly_count"`
	Anomalies     []AnomalyRecord       `json:"anomalies"`
	SPCoverage    []SavingsPlanMetric   `json:"sp_coverage"`
	SPUtilization []SavingsPlanMetric   `json:"sp_utilization"`
	CostData      *CostData             `json:"-"` // Not serialized
}

// SPData holds savings plan coverage and utilization data.
type SPData struct {
	Coverage     []SavingsPlanMetric           `json:"coverage"`
	Utilization  []SavingsPlanMetric           `json:"utilization"`
	SPByPayer    map[string]PayerSPData        `json:"sp_by_payer,omitempty"`
}

// PayerSPData holds payer-specific savings plan data.
type PayerSPData struct {
	Coverage    []SavingsPlanMetric `json:"coverage"`
	Utilization []SavingsPlanMetric `json:"utilization"`
}

// ClusterCounts maps month keys (YYYY-MM) to cluster count.
type ClusterCounts map[string]int

// EnvCounts holds live cluster counts by environment and payer.
type EnvCounts map[string]interface{} // "total", "prod", or payer labels → int

// TimeWindow represents a date range for report generation.
type TimeWindow struct {
	Start  time.Time
	End    time.Time
	Months []time.Time // First day of each month in the window
}
