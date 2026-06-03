package costanomalies

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

// fakeAnomaliesClient implements CostAnomaliesAPI for testing.
type fakeAnomaliesClient struct {
	pages    [][]types.Anomaly
	callsIdx int
	err      error
}

func (f *fakeAnomaliesClient) GetAnomalies(
	_ context.Context,
	_ *costexplorer.GetAnomaliesInput,
	_ ...func(*costexplorer.Options),
) (*costexplorer.GetAnomaliesOutput, error) {
	if f.err != nil {
		return nil, f.err
	}
	idx := f.callsIdx
	f.callsIdx++
	if idx >= len(f.pages) {
		return &costexplorer.GetAnomaliesOutput{}, nil
	}
	var nextToken *string
	if idx+1 < len(f.pages) {
		tok := "page2"
		nextToken = &tok
	}
	return &costexplorer.GetAnomaliesOutput{
		Anomalies:     f.pages[idx],
		NextPageToken: nextToken,
	}, nil
}

func makeAnomaly(id, service, start, end string, impact, actual, expected, pct, score float64, causes []types.RootCause) types.Anomaly {
	a := types.Anomaly{
		AnomalyId:      aws.String(id),
		DimensionValue: aws.String(service),
		AnomalyStartDate: aws.String(start),
		AnomalyEndDate:   aws.String(end),
		AnomalyScore:     &types.AnomalyScore{CurrentScore: score, MaxScore: score},
		Impact: &types.Impact{
			TotalImpact:           impact,
			TotalActualSpend:      aws.Float64(actual),
			TotalExpectedSpend:    aws.Float64(expected),
			TotalImpactPercentage: aws.Float64(pct),
		},
		RootCauses: causes,
	}
	return a
}

var testDateRange = cost.DateRange{
	Start: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
	End:   time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
}

func TestBuildWith_HappyPath(t *testing.T) {
	fake := &fakeAnomaliesClient{
		pages: [][]types.Anomaly{
			{
				makeAnomaly("anom-1", "Amazon EC2", "2026-05-10", "2026-05-12",
					1500.0, 3000.0, 1500.0, 100.0, 0.9, nil),
				makeAnomaly("anom-2", "Amazon S3", "2026-05-20", "2026-05-21",
					200.0, 450.0, 250.0, 80.0, 0.6, nil),
			},
		},
	}

	report, err := buildWith(context.Background(), fake, testDateRange)
	if err != nil {
		t.Fatalf("buildWith returned error: %v", err)
	}

	if len(report.Anomalies) != 2 {
		t.Fatalf("expected 2 anomalies, got %d", len(report.Anomalies))
	}

	// Should be sorted by TotalImpact descending.
	if report.Anomalies[0].ID != "anom-1" {
		t.Errorf("Anomalies[0].ID = %q, want anom-1 (highest impact first)", report.Anomalies[0].ID)
	}
	if report.Anomalies[0].TotalImpact != 1500.0 {
		t.Errorf("Anomalies[0].TotalImpact = %f, want 1500.0", report.Anomalies[0].TotalImpact)
	}
	if report.Anomalies[0].Service != "Amazon EC2" {
		t.Errorf("Anomalies[0].Service = %q, want Amazon EC2", report.Anomalies[0].Service)
	}
	if report.Anomalies[1].ID != "anom-2" {
		t.Errorf("Anomalies[1].ID = %q, want anom-2", report.Anomalies[1].ID)
	}
	if report.StartDate != "2026-05-01" {
		t.Errorf("StartDate = %q, want 2026-05-01", report.StartDate)
	}
	if report.EndDate != "2026-05-31" {
		t.Errorf("EndDate = %q, want 2026-05-31", report.EndDate)
	}
	if report.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should not be zero")
	}
}

func TestBuildWith_Pagination(t *testing.T) {
	fake := &fakeAnomaliesClient{
		pages: [][]types.Anomaly{
			{makeAnomaly("anom-p1", "Amazon EC2", "2026-05-01", "2026-05-02", 500.0, 1000.0, 500.0, 100.0, 0.8, nil)},
			{makeAnomaly("anom-p2", "Amazon RDS", "2026-05-05", "2026-05-06", 300.0, 600.0, 300.0, 100.0, 0.7, nil)},
		},
	}

	report, err := buildWith(context.Background(), fake, testDateRange)
	if err != nil {
		t.Fatalf("buildWith returned error: %v", err)
	}
	if len(report.Anomalies) != 2 {
		t.Errorf("expected 2 anomalies across pages, got %d", len(report.Anomalies))
	}
	if fake.callsIdx != 2 {
		t.Errorf("expected 2 API calls for pagination, got %d", fake.callsIdx)
	}
}

func TestBuildWith_EmptyResult(t *testing.T) {
	fake := &fakeAnomaliesClient{
		pages: [][]types.Anomaly{{}},
	}
	report, err := buildWith(context.Background(), fake, testDateRange)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Anomalies) != 0 {
		t.Errorf("expected 0 anomalies, got %d", len(report.Anomalies))
	}
}

func TestBuildWith_APIError(t *testing.T) {
	fake := &fakeAnomaliesClient{
		err: fmt.Errorf("access denied"),
	}
	_, err := buildWith(context.Background(), fake, testDateRange)
	if err == nil {
		t.Fatal("expected error from API failure, got nil")
	}
}

func TestBuildWith_SkipsNilRequiredFields(t *testing.T) {
	// Anomaly missing required fields should be silently skipped.
	fake := &fakeAnomaliesClient{
		pages: [][]types.Anomaly{
			{
				{AnomalyId: nil, AnomalyScore: nil, Impact: nil}, // skipped
				makeAnomaly("anom-ok", "Amazon EC2", "2026-05-10", "2026-05-12",
					1000.0, 2000.0, 1000.0, 100.0, 0.85, nil),
			},
		},
	}
	report, err := buildWith(context.Background(), fake, testDateRange)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Anomalies) != 1 {
		t.Errorf("expected 1 anomaly (nil entry skipped), got %d", len(report.Anomalies))
	}
}

func TestBuildWith_RootCauses(t *testing.T) {
	causes := []types.RootCause{
		{
			LinkedAccount:     aws.String("123456789012"),
			LinkedAccountName: aws.String("my-account"),
			Region:            aws.String("us-east-1"),
			Service:           aws.String("Amazon EC2"),
			UsageType:         aws.String("BoxUsage:m5.xlarge"),
			Impact:            &types.RootCauseImpact{Contribution: 750.0},
		},
	}

	fake := &fakeAnomaliesClient{
		pages: [][]types.Anomaly{
			{makeAnomaly("anom-rc", "Amazon EC2", "2026-05-01", "2026-05-03",
				750.0, 1500.0, 750.0, 100.0, 0.9, causes)},
		},
	}

	report, err := buildWith(context.Background(), fake, testDateRange)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Anomalies) != 1 {
		t.Fatalf("expected 1 anomaly, got %d", len(report.Anomalies))
	}

	a := report.Anomalies[0]
	if len(a.RootCauses) != 1 {
		t.Fatalf("expected 1 root cause, got %d", len(a.RootCauses))
	}
	rc := a.RootCauses[0]
	if rc.Account != "123456789012" {
		t.Errorf("rc.Account = %q, want 123456789012", rc.Account)
	}
	if rc.AccountName != "my-account" {
		t.Errorf("rc.AccountName = %q, want my-account", rc.AccountName)
	}
	if rc.Service != "Amazon EC2" {
		t.Errorf("rc.Service = %q, want Amazon EC2", rc.Service)
	}
	if rc.Contribution != 750.0 {
		t.Errorf("rc.Contribution = %f, want 750.0", rc.Contribution)
	}
}

func TestBuildWith_ServiceFallsBackToRootCause(t *testing.T) {
	// When DimensionValue is empty, service should come from first root cause.
	a := types.Anomaly{
		AnomalyId:      aws.String("anom-fallback"),
		DimensionValue: aws.String(""), // empty
		AnomalyScore:   &types.AnomalyScore{CurrentScore: 0.7, MaxScore: 0.7},
		Impact:         &types.Impact{TotalImpact: 500.0},
		RootCauses: []types.RootCause{
			{Service: aws.String("Amazon RDS")},
		},
	}
	fake := &fakeAnomaliesClient{
		pages: [][]types.Anomaly{{a}},
	}
	report, err := buildWith(context.Background(), fake, testDateRange)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(report.Anomalies) != 1 {
		t.Fatalf("expected 1 anomaly, got %d", len(report.Anomalies))
	}
	if report.Anomalies[0].Service != "Amazon RDS" {
		t.Errorf("Service = %q, want Amazon RDS (fallback from root cause)", report.Anomalies[0].Service)
	}
}

func TestParseAnomalies_SortedByImpactDesc(t *testing.T) {
	raw := []types.Anomaly{
		makeAnomaly("low", "S3", "2026-05-01", "2026-05-01", 100.0, 200.0, 100.0, 100.0, 0.5, nil),
		makeAnomaly("high", "EC2", "2026-05-02", "2026-05-02", 5000.0, 8000.0, 3000.0, 166.0, 0.95, nil),
		makeAnomaly("mid", "RDS", "2026-05-03", "2026-05-03", 500.0, 900.0, 400.0, 125.0, 0.7, nil),
	}

	result := parseAnomalies(raw)

	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}
	if result[0].ID != "high" || result[1].ID != "mid" || result[2].ID != "low" {
		t.Errorf("wrong sort order: %s, %s, %s", result[0].ID, result[1].ID, result[2].ID)
	}
}
