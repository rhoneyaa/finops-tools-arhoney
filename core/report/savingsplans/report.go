// Package savingsplans fetches Savings Plans coverage and utilization from AWS Cost Explorer.
package savingsplans

import (
	"context"
	"errors"
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
	IsLinked    bool
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
	ceDR := monthlyCERange(dr)
	ceClients := make(map[string]SavingsPlansAPI)

	sections := make([]AccountReport, 0, len(accounts))
	for _, acct := range accounts {
		credID := acct.CredentialsAccountID()
		ce, ok := ceClients[credID]
		if !ok {
			ce = newClient(acct.AWSConfig)
			ceClients[credID] = ce
		}

		coverage, utilization, err := buildAccountWith(ctx, ce, ceDR, acct)
		if err != nil {
			return Report{}, fmt.Errorf("%s: %w", accountDisplayName(acct), err)
		}
		sections = append(sections, AccountReport{
			AccountID:   acct.AccountID,
			AccountName: accountDisplayName(acct),
			IsLinked:    acct.IsLinked(),
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
	acct cost.AccountTarget,
) ([]MonthlyMetric, []MonthlyMetric, error) {
	ceDR := monthlyCERange(dr)
	filter := linkedAccountFilter(acct)
	interval := &types.DateInterval{
		Start: aws.String(ceDR.Start.Format("2006-01-02")),
		End:   aws.String(ceDR.End.Format("2006-01-02")),
	}

	coverageResp, err := ce.GetSavingsPlansCoverage(ctx, &costexplorer.GetSavingsPlansCoverageInput{
		TimePeriod:  interval,
		Granularity: types.GranularityMonthly,
		Filter:      filter,
	})
	var coverages []types.SavingsPlansCoverage
	if err != nil {
		if !isDataUnavailable(err) {
			return nil, nil, fmt.Errorf("fetch SP coverage: %w", err)
		}
	} else {
		if coverageResp == nil {
			return nil, nil, fmt.Errorf("nil response from GetSavingsPlansCoverage")
		}
		coverages = coverageResp.SavingsPlansCoverages
	}

	utilizationResp, err := ce.GetSavingsPlansUtilization(ctx, &costexplorer.GetSavingsPlansUtilizationInput{
		TimePeriod:  interval,
		Granularity: types.GranularityMonthly,
		Filter:      filter,
	})
	var utils []types.SavingsPlansUtilizationByTime
	if err != nil {
		if !isDataUnavailable(err) {
			return nil, nil, fmt.Errorf("fetch SP utilization: %w", err)
		}
	} else {
		if utilizationResp == nil {
			return nil, nil, fmt.Errorf("nil response from GetSavingsPlansUtilization")
		}
		utils = utilizationResp.SavingsPlansUtilizationsByTime
	}

	return parseCoverageMetrics(coverages), parseUtilizationMetrics(utils), nil
}

// monthlyCERange aligns dr to calendar-month boundaries for CE MONTHLY granularity.
// Start moves to the first day of its month. End (exclusive) moves to the first day of
// the month after the last included calendar day, but never beyond the caller's End —
// extending the window past the latest available CE data triggers ValidationException.
func monthlyCERange(dr cost.DateRange) cost.DateRange {
	start := time.Date(dr.Start.Year(), dr.Start.Month(), 1, 0, 0, 0, 0, time.UTC)
	lastIncluded := dr.End.AddDate(0, 0, -1)
	end := time.Date(lastIncluded.Year(), lastIncluded.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, 0)
	if end.After(dr.End) {
		end = dr.End
	}
	return cost.DateRange{Start: start, End: end}
}

func isDataUnavailable(err error) bool {
	var du *types.DataUnavailableException
	return errors.As(err, &du)
}

func linkedAccountFilter(acct cost.AccountTarget) *types.Expression {
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
