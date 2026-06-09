// Package savingsplans fetches Savings Plans coverage and utilization from AWS Cost Explorer.
package savingsplans

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
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

// AccountReport holds Savings Plans coverage and utilization for one account.
type AccountReport struct {
	AccountID   string
	AccountName string
	Coverage    []MonthlyMetric
	Utilization []MonthlyMetric
}

// Report holds Savings Plans coverage and utilization ready for HTML rendering.
type Report struct {
	GeneratedAt time.Time
	StartDate   string
	EndDate     string
	Accounts    []AccountReport
}

// Build fetches Savings Plans coverage and utilization from AWS Cost Explorer
// for each account in the given date range.
func Build(ctx context.Context, accounts []cost.AccountTarget, dr cost.DateRange) (Report, error) {
	return buildWith(ctx, defaultCEClientFactory, accounts, dr)
}

type ceClientFactory func(cfg aws.Config) SavingsPlansAPI

func defaultCEClientFactory(cfg aws.Config) SavingsPlansAPI {
	region := cfg.Region
	if region == "" {
		region = costExplorerRegion
	}
	cfg.Region = region
	return costexplorer.NewFromConfig(cfg, func(o *costexplorer.Options) {
		o.Region = costExplorerRegion
	})
}

func buildWith(ctx context.Context, newClient ceClientFactory, accounts []cost.AccountTarget, dr cost.DateRange) (Report, error) {
	if len(accounts) == 0 {
		return Report{}, fmt.Errorf("at least one account required")
	}
	targets := cost.FilterOverlappingTargets(accounts)
	ceClients := make(map[string]SavingsPlansAPI)

	sections := make([]AccountReport, 0, len(targets))
	for _, acct := range targets {
		credID := acct.CredentialsAccountID()
		ce, ok := ceClients[credID]
		if !ok {
			ce = newClient(acct.AWSConfig)
			ceClients[credID] = ce
		}

		coverage, utilization, err := buildAccountWith(ctx, ce, dr, accountFilter(acct))
		if err != nil {
			return Report{}, fmt.Errorf("%s: %w", accountDisplayName(acct), err)
		}
		sections = append(sections, AccountReport{
			AccountID:   acct.AccountID,
			AccountName: accountDisplayName(acct),
			Coverage:    coverage,
			Utilization: utilization,
		})
	}

	return Report{
		GeneratedAt: time.Now().UTC(),
		StartDate:   dr.Start.Format("2006-01-02"),
		EndDate:     dr.End.AddDate(0, 0, -1).Format("2006-01-02"),
		Accounts:    sections,
	}, nil
}

// buildAccountWith is the testable core for one account scope.
func buildAccountWith(
	ctx context.Context,
	ce SavingsPlansAPI,
	dr cost.DateRange,
	filter *types.Expression,
) ([]MonthlyMetric, []MonthlyMetric, error) {
	interval := &types.DateInterval{
		Start: aws.String(dr.Start.Format("2006-01-02")),
		End:   aws.String(dr.End.Format("2006-01-02")),
	}

	coverageResp, err := ce.GetSavingsPlansCoverage(ctx, &costexplorer.GetSavingsPlansCoverageInput{
		TimePeriod:  interval,
		Granularity: types.GranularityMonthly,
		Filter:      filter,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("fetch SP coverage: %w", err)
	}

	utilizationResp, err := ce.GetSavingsPlansUtilization(ctx, &costexplorer.GetSavingsPlansUtilizationInput{
		TimePeriod:  interval,
		Granularity: types.GranularityMonthly,
		Filter:      filter,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("fetch SP utilization: %w", err)
	}

	return parseCoverageMetrics(coverageResp.SavingsPlansCoverages),
		parseUtilizationMetrics(utilizationResp.SavingsPlansUtilizationsByTime),
		nil
}

func accountFilter(acct cost.AccountTarget) *types.Expression {
	if !acct.ScopeToAccount() {
		return nil
	}
	return &types.Expression{
		Dimensions: &types.DimensionValues{
			Key:    types.DimensionLinkedAccount,
			Values: []string{acct.AccountID},
		},
	}
}

func accountDisplayName(acct cost.AccountTarget) string {
	if name := strings.TrimSpace(acct.DisplayName); name != "" {
		return name
	}
	if alias := strings.TrimSpace(acct.DisplayAlias); alias != "" {
		return alias
	}
	return strings.TrimSpace(acct.AccountID)
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
