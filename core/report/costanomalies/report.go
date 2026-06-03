// Package costanomalies fetches cost anomalies detected by AWS Cost Anomaly Detection.
package costanomalies

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer"
	"github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/openshift-online/finops-tools/core/cost"
)

const costExplorerRegion = "us-east-1"

// CostAnomaliesAPI is the subset of the CE client used for anomaly fetch (mockable).
type CostAnomaliesAPI interface {
	GetAnomalies(
		ctx context.Context,
		params *costexplorer.GetAnomaliesInput,
		optFns ...func(*costexplorer.Options),
	) (*costexplorer.GetAnomaliesOutput, error)
}

// RootCause is one root-cause attribution for an anomaly.
type RootCause struct {
	Account      string  // linked account ID
	AccountName  string
	Region       string
	Service      string
	UsageType    string
	Contribution float64 // dollar contribution to the anomaly
}

// Anomaly is one detected cost anomaly from AWS Cost Anomaly Detection.
type Anomaly struct {
	ID             string
	StartDate      string  // YYYY-MM-DD
	EndDate        string  // YYYY-MM-DD; empty if still open
	Service        string  // DimensionValue (service or account dimension)
	CurrentScore   float64 // anomaly score at time of report
	MaxScore       float64
	TotalImpact    float64 // actual − expected spend (dollars)
	ActualSpend    float64
	ExpectedSpend  float64
	ImpactPct      float64 // (TotalImpact / ExpectedSpend) × 100
	RootCauses     []RootCause
}

// Report holds all anomalies in the requested date range.
type Report struct {
	GeneratedAt time.Time
	StartDate   string
	EndDate     string
	Anomalies   []Anomaly
}

// Build fetches cost anomalies from AWS Cost Anomaly Detection for the given
// date range, using cfg for authentication. Anomalies are sorted by total
// dollar impact descending.
func Build(ctx context.Context, cfg aws.Config, dr cost.DateRange) (Report, error) {
	cfg.Region = costExplorerRegion
	ce := costexplorer.NewFromConfig(cfg, func(o *costexplorer.Options) {
		o.Region = costExplorerRegion
	})
	return buildWith(ctx, ce, dr)
}

// buildWith is the testable core: accepts any CostAnomaliesAPI implementation.
func buildWith(ctx context.Context, ce CostAnomaliesAPI, dr cost.DateRange) (Report, error) {
	raw, err := fetchAll(ctx, ce, dr)
	if err != nil {
		return Report{}, err
	}

	anomalies := parseAnomalies(raw)

	return Report{
		GeneratedAt: time.Now().UTC(),
		StartDate:   dr.Start.Format("2006-01-02"),
		EndDate:     dr.End.AddDate(0, 0, -1).Format("2006-01-02"),
		Anomalies:   anomalies,
	}, nil
}

// fetchAll paginates through GetAnomalies until all results are collected.
func fetchAll(ctx context.Context, ce CostAnomaliesAPI, dr cost.DateRange) ([]types.Anomaly, error) {
	input := &costexplorer.GetAnomaliesInput{
		DateInterval: &types.AnomalyDateInterval{
			StartDate: aws.String(dr.Start.Format("2006-01-02")),
			EndDate:   aws.String(dr.End.AddDate(0, 0, -1).Format("2006-01-02")),
		},
	}

	var all []types.Anomaly
	for {
		resp, err := ce.GetAnomalies(ctx, input)
		if err != nil {
			return nil, fmt.Errorf("fetch anomalies: %w", err)
		}
		all = append(all, resp.Anomalies...)

		if resp.NextPageToken == nil || *resp.NextPageToken == "" {
			break
		}
		input.NextPageToken = resp.NextPageToken
	}

	return all, nil
}

// parseAnomalies converts AWS Anomaly slice into the Report's Anomaly slice,
// sorted by TotalImpact descending.
func parseAnomalies(raw []types.Anomaly) []Anomaly {
	result := make([]Anomaly, 0, len(raw))
	for _, a := range raw {
		if a.AnomalyId == nil || a.AnomalyScore == nil || a.Impact == nil {
			continue
		}

		anomaly := Anomaly{
			ID:           *a.AnomalyId,
			Service:      aws.ToString(a.DimensionValue),
			CurrentScore: a.AnomalyScore.CurrentScore,
			MaxScore:     a.AnomalyScore.MaxScore,
			TotalImpact:  a.Impact.TotalImpact,
			RootCauses:   parseRootCauses(a.RootCauses),
		}

		if a.AnomalyStartDate != nil {
			anomaly.StartDate = *a.AnomalyStartDate
		}
		if a.AnomalyEndDate != nil {
			anomaly.EndDate = *a.AnomalyEndDate
		}
		if a.Impact.TotalActualSpend != nil {
			anomaly.ActualSpend = *a.Impact.TotalActualSpend
		}
		if a.Impact.TotalExpectedSpend != nil {
			anomaly.ExpectedSpend = *a.Impact.TotalExpectedSpend
		}
		if a.Impact.TotalImpactPercentage != nil {
			anomaly.ImpactPct = *a.Impact.TotalImpactPercentage
		}

		// Derive service label from root causes when DimensionValue is empty.
		if anomaly.Service == "" && len(anomaly.RootCauses) > 0 {
			anomaly.Service = anomaly.RootCauses[0].Service
		}

		result = append(result, anomaly)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalImpact > result[j].TotalImpact
	})

	return result
}

func parseRootCauses(raw []types.RootCause) []RootCause {
	causes := make([]RootCause, 0, len(raw))
	for _, r := range raw {
		rc := RootCause{
			Account:     aws.ToString(r.LinkedAccount),
			AccountName: aws.ToString(r.LinkedAccountName),
			Region:      aws.ToString(r.Region),
			Service:     aws.ToString(r.Service),
			UsageType:   aws.ToString(r.UsageType),
		}
		if r.Impact != nil {
			rc.Contribution = r.Impact.Contribution
		}
		causes = append(causes, rc)
	}
	return causes
}
