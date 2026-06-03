// Package savingsplans fetches Savings Plans coverage and utilization from AWS Cost Explorer.
package savingsplans

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/openshift-online/finops-tools/core/cost"
)

const costExplorerRegion = "us-east-1"

// SavingsPlansAPI is the subset of the CE client used for savings plans fetch (mockable).
type SavingsPlansAPI interface {
	GetSavingsPlansCoverage(
		ctx context.Context,
		params *costexplorer.GetSavingsPlansCoverageInput,
		optFns ...func(*costexplorer.Options),
	) (*costexplorer.GetSavingsPlansCoverageOutput, error)

	GetSavingsPlansUtilization(
		ctx context.Context,
		params *costexplorer.GetSavingsPlansUtilizationInput,
		optFns ...func(*costexplorer.Options),
	) (*costexplorer.GetSavingsPlansUtilizationOutput, error)
}

// MonthlyMetric holds a savings plan percentage metric for one calendar month.
type MonthlyMetric struct {
	Month      string  // YYYY-MM
	Percentage float64 // 0–100
}

// Report holds Savings Plans coverage and utilization ready for HTML rendering.
type Report struct {
	GeneratedAt time.Time
	StartDate   string
	EndDate     string
	Coverage    []MonthlyMetric
	Utilization []MonthlyMetric
}

// Build fetches Savings Plans coverage and utilization from AWS Cost Explorer
// for the given date range, using cfg for authentication.
func Build(ctx context.Context, cfg aws.Config, dr cost.DateRange) (Report, error) {
	region := cfg.Region
	if region == "" {
		region = costExplorerRegion
	}
	cfg.Region = region

	ce := costexplorer.NewFromConfig(cfg, func(o *costexplorer.Options) {
		o.Region = costExplorerRegion
	})
	return buildWith(ctx, ce, dr)
}

// buildWith is the testable core: it accepts any SavingsPlansAPI implementation.
func buildWith(ctx context.Context, ce SavingsPlansAPI, dr cost.DateRange) (Report, error) {
	interval := &types.DateInterval{
		Start: aws.String(dr.Start.Format("2006-01-02")),
		End:   aws.String(dr.End.Format("2006-01-02")),
	}

	coverageResp, err := ce.GetSavingsPlansCoverage(ctx, &costexplorer.GetSavingsPlansCoverageInput{
		TimePeriod:  interval,
		Granularity: types.GranularityMonthly,
	})
	if err != nil {
		return Report{}, fmt.Errorf("fetch SP coverage: %w", err)
	}

	utilizationResp, err := ce.GetSavingsPlansUtilization(ctx, &costexplorer.GetSavingsPlansUtilizationInput{
		TimePeriod:  interval,
		Granularity: types.GranularityMonthly,
	})
	if err != nil {
		return Report{}, fmt.Errorf("fetch SP utilization: %w", err)
	}

	coverage := parseCoverageMetrics(coverageResp.SavingsPlansCoverages)
	utilization := parseUtilizationMetrics(utilizationResp.SavingsPlansUtilizationsByTime)

	return Report{
		GeneratedAt: time.Now().UTC(),
		StartDate:   dr.Start.Format("2006-01"),
		EndDate:     lastMonthLabel(dr.End),
		Coverage:    coverage,
		Utilization: utilization,
	}, nil
}

// parseCoverageMetrics converts AWS SavingsPlansCoverages to sorted MonthlyMetric slice.
func parseCoverageMetrics(coverages []types.SavingsPlansCoverage) []MonthlyMetric {
	metrics := make([]MonthlyMetric, 0, len(coverages))
	for _, c := range coverages {
		if c.TimePeriod == nil || c.Coverage == nil {
			continue
		}
		pct := parseFloatPtr(c.Coverage.CoveragePercentage)
		metrics = append(metrics, MonthlyMetric{
			Month:      monthLabel(aws.ToString(c.TimePeriod.Start)),
			Percentage: pct,
		})
	}
	sortMetrics(metrics)
	return metrics
}

// parseUtilizationMetrics converts AWS SavingsPlansUtilizationsByTime to sorted MonthlyMetric slice.
func parseUtilizationMetrics(utils []types.SavingsPlansUtilizationByTime) []MonthlyMetric {
	metrics := make([]MonthlyMetric, 0, len(utils))
	for _, u := range utils {
		if u.TimePeriod == nil || u.Utilization == nil {
			continue
		}
		pct := parseFloatPtr(u.Utilization.UtilizationPercentage)
		metrics = append(metrics, MonthlyMetric{
			Month:      monthLabel(aws.ToString(u.TimePeriod.Start)),
			Percentage: pct,
		})
	}
	sortMetrics(metrics)
	return metrics
}

func sortMetrics(m []MonthlyMetric) {
	sort.Slice(m, func(i, j int) bool { return m[i].Month < m[j].Month })
}

// monthLabel converts an AWS date string (YYYY-MM-DD) to YYYY-MM.
func monthLabel(dateStr string) string {
	if len(dateStr) >= 7 {
		return dateStr[:7]
	}
	return dateStr
}

// lastMonthLabel returns the YYYY-MM label for the month immediately before the exclusive end date.
func lastMonthLabel(end time.Time) string {
	prev := end.AddDate(0, -1, 0)
	return prev.Format("2006-01")
}

// parseFloatPtr parses a *string to float64, returning 0 on nil or error.
func parseFloatPtr(s *string) float64 {
	if s == nil {
		return 0
	}
	f, err := strconv.ParseFloat(*s, 64)
	if err != nil {
		return 0
	}
	return f
}
