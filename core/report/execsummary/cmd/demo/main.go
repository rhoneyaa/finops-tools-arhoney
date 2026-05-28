// Demo program showing execsummary package usage
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/openshift-online/finops-tools/core/report/execsummary"
)

func main() {
	fmt.Println("Executive Summary Package Demo")
	fmt.Println("================================\n")

	// Demo 1: Date utilities
	fmt.Println("1. Date Utilities")
	fmt.Println("-----------------")
	now := time.Now()
	fmt.Printf("Current date: %s\n", now.Format("2006-01-02"))
	fmt.Printf("Month label: %s\n", execsummary.MonthLabel(now))
	fmt.Printf("Month name: %s\n", execsummary.MonthName(now))

	// Demo 2: Compute 6-month window
	fmt.Println("\n2. Compute 6-Month Window")
	fmt.Println("-------------------------")
	window, err := execsummary.ComputeWindow(6, nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Start: %s (%s)\n", execsummary.MonthLabel(window.Start), execsummary.MonthName(window.Start))
	fmt.Printf("End:   %s (%s)\n", execsummary.MonthLabel(window.End), execsummary.MonthName(window.End))
	fmt.Printf("Months in window: %d\n", len(window.Months))
	fmt.Print("  ")
	for _, m := range window.Months {
		fmt.Printf("%s ", execsummary.MonthLabel(m))
	}
	fmt.Println()

	// Demo 3: Sample cost data
	fmt.Println("\n3. Sample Cost Data Processing")
	fmt.Println("-------------------------------")

	costData := &execsummary.CostData{
		Records: []execsummary.CostRecord{
			{Month: "2026-01", AccountID: "111111111111", Cost: 1500.50, Service: "EC2", Payer: "payer1"},
			{Month: "2026-01", AccountID: "111111111111", Cost: 800.25, Service: "S3", Payer: "payer1"},
			{Month: "2026-01", AccountID: "222222222222", Cost: 2300.75, Service: "RDS", Payer: "payer1"},
			{Month: "2026-02", AccountID: "111111111111", Cost: 1650.00, Service: "EC2", Payer: "payer1"},
			{Month: "2026-02", AccountID: "222222222222", Cost: 2500.00, Service: "RDS", Payer: "payer1"},
		},
	}

	// Re-index after manual creation
	costData = indexCostData(costData.Records)

	fmt.Printf("Total records: %d\n", len(costData.Records))
	fmt.Printf("Unique months: %d\n", len(costData.ByMonth))
	fmt.Printf("Unique accounts: %d\n", len(costData.ByAccount))

	// Demo 4: Enrichment
	fmt.Println("\n4. Enrichment with Account Mappings")
	fmt.Println("------------------------------------")

	mappings := []execsummary.AccountMapping{
		{
			AccountID:       "111111111111",
			RefinedCategory: "HCP",
			SubType:         "Production",
			OwnerTeam:       "Platform Team",
			AccountName:     "HCP Production",
		},
		{
			AccountID:       "222222222222",
			RefinedCategory: "Tooling",
			OwnerTeam:       "DevOps Team",
			AccountName:     "Shared Tooling",
		},
	}

	enriched := execsummary.EnrichWithMapping(costData, mappings)
	fmt.Printf("Enriched %d records\n", len(enriched))

	// Show first few
	fmt.Println("\nSample enriched records:")
	for i := 0; i < 3 && i < len(enriched); i++ {
		rec := enriched[i]
		fmt.Printf("  %s | %-15s | %-12s | $%.2f | %s\n",
			rec.Month, rec.AccountName, rec.RefinedCategory, rec.Cost, rec.Service)
	}

	// Demo 5: Aggregation
	fmt.Println("\n5. Category Monthly Aggregation")
	fmt.Println("--------------------------------")

	months := []time.Time{
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	categoryMonthly := execsummary.CategoryMonthly(enriched, months)
	fmt.Printf("Generated %d category-month aggregates:\n", len(categoryMonthly))
	for _, cm := range categoryMonthly {
		fmt.Printf("  %s | %s: $%.2f\n", cm.Month, cm.RefinedCategory, cm.Cost)
	}

	// Demo 6: Top services
	fmt.Println("\n6. Top Services for Account")
	fmt.Println("----------------------------")

	topServices := execsummary.ServiceCostsForAccount(enriched, "111111111111", "2026-01", 3)
	fmt.Printf("Top services for account 111111111111 in 2026-01:\n")
	for i, svc := range topServices {
		fmt.Printf("  %d. %-10s $%.2f\n", i+1, svc.Service, svc.Cost)
	}

	// Demo 7: Total cost
	fmt.Println("\n7. Total Cost")
	fmt.Println("-------------")
	total := execsummary.TotalCost(enriched)
	fmt.Printf("Total cost across all records: $%.2f\n", total)

	fmt.Println("\n✅ Demo complete!")
}

// indexCostData helper function to index cost data
func indexCostData(records []execsummary.CostRecord) *execsummary.CostData {
	data := &execsummary.CostData{
		Records:        records,
		ByMonth:        make(map[string][]execsummary.CostRecord),
		ByAccount:      make(map[string][]execsummary.CostRecord),
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
