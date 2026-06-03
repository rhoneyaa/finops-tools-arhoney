package savingsplans

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/openshift-online/finops-tools/core/cost"
)

// fakeSavingsPlansClient implements SavingsPlansAPI for testing.
type fakeSavingsPlansClient struct {
	coverageResp    *costexplorer.GetSavingsPlansCoverageOutput
	utilizationResp *costexplorer.GetSavingsPlansUtilizationOutput
	coverageErr     error
	utilizationErr  error
}

func (f *fakeSavingsPlansClient) GetSavingsPlansCoverage(
	_ context.Context,
	_ *costexplorer.GetSavingsPlansCoverageInput,
	_ ...func(*costexplorer.Options),
) (*costexplorer.GetSavingsPlansCoverageOutput, error) {
	return f.coverageResp, f.coverageErr
}

func (f *fakeSavingsPlansClient) GetSavingsPlansUtilization(
	_ context.Context,
	_ *costexplorer.GetSavingsPlansUtilizationInput,
	_ ...func(*costexplorer.Options),
) (*costexplorer.GetSavingsPlansUtilizationOutput, error) {
	return f.utilizationResp, f.utilizationErr
}

func TestParseCoverageMetrics(t *testing.T) {
	input := []types.SavingsPlansCoverage{
		{
			TimePeriod: &types.DateInterval{Start: aws.String("2026-01-01"), End: aws.String("2026-02-01")},
			Coverage:   &types.SavingsPlansCoverageData{CoveragePercentage: aws.String("75.5")},
		},
		{
			TimePeriod: &types.DateInterval{Start: aws.String("2026-02-01"), End: aws.String("2026-03-01")},
			Coverage:   &types.SavingsPlansCoverageData{CoveragePercentage: aws.String("82.3")},
		},
	}

	metrics := parseCoverageMetrics(input)

	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}
	if metrics[0].Month != "2026-01" {
		t.Errorf("metrics[0].Month = %q, want %q", metrics[0].Month, "2026-01")
	}
	if metrics[0].Percentage != 75.5 {
		t.Errorf("metrics[0].Percentage = %f, want 75.5", metrics[0].Percentage)
	}
	if metrics[1].Month != "2026-02" {
		t.Errorf("metrics[1].Month = %q, want %q", metrics[1].Month, "2026-02")
	}
	if metrics[1].Percentage != 82.3 {
		t.Errorf("metrics[1].Percentage = %f, want 82.3", metrics[1].Percentage)
	}
}

func TestParseCoverageMetrics_SkipsNilEntries(t *testing.T) {
	input := []types.SavingsPlansCoverage{
		{TimePeriod: nil, Coverage: nil}, // should be skipped
		{
			TimePeriod: &types.DateInterval{Start: aws.String("2026-03-01"), End: aws.String("2026-04-01")},
			Coverage:   &types.SavingsPlansCoverageData{CoveragePercentage: aws.String("60.0")},
		},
	}

	metrics := parseCoverageMetrics(input)

	if len(metrics) != 1 {
		t.Fatalf("expected 1 metric, got %d", len(metrics))
	}
	if metrics[0].Month != "2026-03" {
		t.Errorf("metrics[0].Month = %q, want %q", metrics[0].Month, "2026-03")
	}
}

func TestParseUtilizationMetrics(t *testing.T) {
	input := []types.SavingsPlansUtilizationByTime{
		{
			TimePeriod:  &types.DateInterval{Start: aws.String("2026-01-01"), End: aws.String("2026-02-01")},
			Utilization: &types.SavingsPlansUtilization{UtilizationPercentage: aws.String("88.2")},
		},
		{
			TimePeriod:  &types.DateInterval{Start: aws.String("2026-02-01"), End: aws.String("2026-03-01")},
			Utilization: &types.SavingsPlansUtilization{UtilizationPercentage: aws.String("91.0")},
		},
	}

	metrics := parseUtilizationMetrics(input)

	if len(metrics) != 2 {
		t.Fatalf("expected 2 metrics, got %d", len(metrics))
	}
	if metrics[0].Percentage != 88.2 {
		t.Errorf("metrics[0].Percentage = %f, want 88.2", metrics[0].Percentage)
	}
	if metrics[1].Percentage != 91.0 {
		t.Errorf("metrics[1].Percentage = %f, want 91.0", metrics[1].Percentage)
	}
}

func TestParseUtilizationMetrics_SortedByMonth(t *testing.T) {
	// Input is intentionally out of order.
	input := []types.SavingsPlansUtilizationByTime{
		{
			TimePeriod:  &types.DateInterval{Start: aws.String("2026-03-01"), End: aws.String("2026-04-01")},
			Utilization: &types.SavingsPlansUtilization{UtilizationPercentage: aws.String("95.0")},
		},
		{
			TimePeriod:  &types.DateInterval{Start: aws.String("2026-01-01"), End: aws.String("2026-02-01")},
			Utilization: &types.SavingsPlansUtilization{UtilizationPercentage: aws.String("80.0")},
		},
	}

	metrics := parseUtilizationMetrics(input)

	if metrics[0].Month != "2026-01" {
		t.Errorf("expected sorted: metrics[0].Month = %q, want 2026-01", metrics[0].Month)
	}
	if metrics[1].Month != "2026-03" {
		t.Errorf("expected sorted: metrics[1].Month = %q, want 2026-03", metrics[1].Month)
	}
}

func TestBuildWith_HappyPath(t *testing.T) {
	fake := &fakeSavingsPlansClient{
		coverageResp: &costexplorer.GetSavingsPlansCoverageOutput{
			SavingsPlansCoverages: []types.SavingsPlansCoverage{
				{
					TimePeriod: &types.DateInterval{Start: aws.String("2026-01-01"), End: aws.String("2026-02-01")},
					Coverage:   &types.SavingsPlansCoverageData{CoveragePercentage: aws.String("72.0")},
				},
			},
		},
		utilizationResp: &costexplorer.GetSavingsPlansUtilizationOutput{
			SavingsPlansUtilizationsByTime: []types.SavingsPlansUtilizationByTime{
				{
					TimePeriod:  &types.DateInterval{Start: aws.String("2026-01-01"), End: aws.String("2026-02-01")},
					Utilization: &types.SavingsPlansUtilization{UtilizationPercentage: aws.String("85.5")},
				},
			},
		},
	}

	dr := cost.DateRange{
		Start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}

	report, err := buildWith(context.Background(), fake, dr)
	if err != nil {
		t.Fatalf("buildWith returned error: %v", err)
	}

	if report.StartDate != "2026-01" {
		t.Errorf("StartDate = %q, want %q", report.StartDate, "2026-01")
	}
	if report.EndDate != "2026-01" {
		t.Errorf("EndDate = %q, want %q", report.EndDate, "2026-01")
	}
	if len(report.Coverage) != 1 {
		t.Fatalf("Coverage len = %d, want 1", len(report.Coverage))
	}
	if report.Coverage[0].Percentage != 72.0 {
		t.Errorf("Coverage[0].Percentage = %f, want 72.0", report.Coverage[0].Percentage)
	}
	if len(report.Utilization) != 1 {
		t.Fatalf("Utilization len = %d, want 1", len(report.Utilization))
	}
	if report.Utilization[0].Percentage != 85.5 {
		t.Errorf("Utilization[0].Percentage = %f, want 85.5", report.Utilization[0].Percentage)
	}
	if report.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should not be zero")
	}
}

func TestBuildWith_CoverageAPIError(t *testing.T) {
	fake := &fakeSavingsPlansClient{
		coverageErr: fmt.Errorf("access denied"),
	}
	_, err := buildWith(context.Background(), fake, cost.DateRange{
		Start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Fatal("expected error from coverage API failure, got nil")
	}
}

func TestBuildWith_UtilizationAPIError(t *testing.T) {
	fake := &fakeSavingsPlansClient{
		coverageResp:   &costexplorer.GetSavingsPlansCoverageOutput{},
		utilizationErr: fmt.Errorf("throttled"),
	}
	_, err := buildWith(context.Background(), fake, cost.DateRange{
		Start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		End:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Fatal("expected error from utilization API failure, got nil")
	}
}

func TestMonthLabel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"2026-01-01", "2026-01"},
		{"2026-12-31", "2026-12"},
		{"2025-06-15", "2025-06"},
		{"short", "short"}, // too short, returned as-is
	}
	for _, tt := range tests {
		got := monthLabel(tt.input)
		if got != tt.want {
			t.Errorf("monthLabel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestLastMonthLabel(t *testing.T) {
	end := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	got := lastMonthLabel(end)
	if got != "2026-05" {
		t.Errorf("lastMonthLabel(%v) = %q, want %q", end, got, "2026-05")
	}
}
