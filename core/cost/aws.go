package cost

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
)

const costExplorerRegion = "us-east-1"

// CostExplorerAPI is the subset of the CE client used for cost fetch (mockable).
type CostExplorerAPI interface {
	GetCostAndUsage(
		ctx context.Context,
		params *costexplorer.GetCostAndUsageInput,
		optFns ...func(*costexplorer.Options),
	) (*costexplorer.GetCostAndUsageOutput, error)
}

// ListAWSAccountNamesFunc maps organization account IDs to display names for split-by-account output.
type ListAWSAccountNamesFunc func(context.Context, aws.Config) (map[string]string, error)

type fetchAWSOptions struct {
	Now              time.Time
	NewCostExplorer  func(aws.Config) CostExplorerAPI
	ListAccountNames ListAWSAccountNamesFunc
}

func fetchAWSNetAmortized(ctx context.Context, q CostQuery) (CostResult, error) {
	opts := fetchAWSOptions{
		Now: time.Now(),
		NewCostExplorer: func(cfg aws.Config) CostExplorerAPI {
			return costexplorer.NewFromConfig(cfg, func(o *costexplorer.Options) {
				o.Region = costExplorerRegion
			})
		},
	}
	if q.AWSFetch != nil {
		opts.ListAccountNames = q.AWSFetch.ListAccountNames
	}
	return fetchAWSNetAmortizedWith(ctx, q, opts)
}

func fetchAWSNetAmortizedWith(ctx context.Context, q CostQuery, opts fetchAWSOptions) (CostResult, error) {
	if opts.Now.IsZero() {
		opts.Now = time.Now()
	}
	if opts.NewCostExplorer == nil {
		opts.NewCostExplorer = func(cfg aws.Config) CostExplorerAPI {
			return costexplorer.NewFromConfig(cfg, func(o *costexplorer.Options) {
				o.Region = costExplorerRegion
			})
		}
	}

	acct := q.Accounts[0]
	accountID := acct.AccountID
	cfg := acct.AWSConfig
	if cfg.Region == "" {
		cfg.Region = costExplorerRegion
	}

	dr := LastNDaysRange(q.Days, opts.Now)
	ce := opts.NewCostExplorer(cfg)
	filter := linkedAccountFilter(accountID, acct.IsLinked())

	var (
		amount    float64
		currency  string
		breakdown []CostBreakdownItem
		fetchErr  error
	)
	switch q.SplitBy {
	case SplitByService:
		amount, currency, breakdown, fetchErr = sumNetAmortizedGrouped(ctx, ce, dr, "SERVICE", SplitByService, filter)
	case SplitByAccount:
		amount, currency, breakdown, fetchErr = sumNetAmortizedGrouped(ctx, ce, dr, "LINKED_ACCOUNT", SplitByAccount, filter)
	default:
		amount, currency, fetchErr = sumNetAmortizedCost(ctx, ce, dr, filter)
	}
	if fetchErr != nil {
		return CostResult{}, fetchErr
	}

	if q.SplitBy == SplitByAccount && opts.ListAccountNames != nil {
		breakdown = applyAWSAccountNames(ctx, cfg, breakdown, opts.ListAccountNames)
	}

	return CostResult{
		Provider:    ProviderAWS,
		AccountName: displayAccountName(acct),
		AccountID:   accountID,
		Metric:      MetricNetAmortized,
		SplitBy:     q.SplitBy,
		StartDate:   formatDate(dr.Start),
		EndDate:     formatDate(dr.End.AddDate(0, 0, -1)),
		Amount:      amount,
		Currency:    currency,
		Breakdown:   breakdown,
		Linked:      acct.IsLinked(),
	}, nil
}

func displayAccountName(acct AccountTarget) string {
	if name := strings.TrimSpace(acct.DisplayName); name != "" {
		return name
	}
	if alias := strings.TrimSpace(acct.DisplayAlias); alias != "" {
		return alias
	}
	return strings.TrimSpace(acct.AccountID)
}

func applyAWSAccountNames(
	ctx context.Context,
	cfg aws.Config,
	breakdown []CostBreakdownItem,
	list ListAWSAccountNamesFunc,
) []CostBreakdownItem {
	if len(breakdown) == 0 || list == nil {
		return breakdown
	}
	names, err := list(ctx, cfg)
	if err != nil {
		return breakdown
	}
	for i := range breakdown {
		if name, ok := names[breakdown[i].Account]; ok {
			breakdown[i].AccountName = name
		}
	}
	return breakdown
}

func linkedAccountFilter(accountID string, linked bool) *types.Expression {
	if !linked {
		return nil
	}
	return &types.Expression{
		Dimensions: &types.DimensionValues{
			Key:    types.DimensionLinkedAccount,
			Values: []string{accountID},
		},
	}
}

func sumNetAmortizedCost(ctx context.Context, ce CostExplorerAPI, dr DateRange, filter *types.Expression) (float64, string, error) {
	var (
		total    float64
		currency = "USD"
		token    *string
	)

	for {
		out, err := ce.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
			TimePeriod: &types.DateInterval{
				Start: aws.String(formatDate(dr.Start)),
				End:   aws.String(formatDate(dr.End)),
			},
			Granularity:   types.GranularityDaily,
			Metrics:       []string{MetricNetAmortized},
			Filter:        filter,
			NextPageToken: token,
		})
		if err != nil {
			return 0, "", fmt.Errorf("cost explorer GetCostAndUsage: %w", err)
		}

		for _, row := range out.ResultsByTime {
			m, ok := row.Total[MetricNetAmortized]
			if !ok {
				continue
			}
			amt, err := strconv.ParseFloat(aws.ToString(m.Amount), 64)
			if err != nil {
				return 0, "", fmt.Errorf("parse %s amount: %w", MetricNetAmortized, err)
			}
			total += amt
			if u := aws.ToString(m.Unit); u != "" {
				currency = u
			}
		}

		if out.NextPageToken == nil || *out.NextPageToken == "" {
			break
		}
		token = out.NextPageToken
	}

	return total, currency, nil
}

func sumNetAmortizedGrouped(
	ctx context.Context,
	ce CostExplorerAPI,
	dr DateRange,
	dimension string,
	splitBy SplitBy,
	filter *types.Expression,
) (float64, string, []CostBreakdownItem, error) {
	byKey := make(map[string]float64)
	currency := "USD"
	var token *string

	groupBy := []types.GroupDefinition{{
		Type: types.GroupDefinitionTypeDimension,
		Key:  aws.String(dimension),
	}}

	for {
		out, err := ce.GetCostAndUsage(ctx, &costexplorer.GetCostAndUsageInput{
			TimePeriod: &types.DateInterval{
				Start: aws.String(formatDate(dr.Start)),
				End:   aws.String(formatDate(dr.End)),
			},
			Granularity:   types.GranularityDaily,
			Metrics:       []string{MetricNetAmortized},
			GroupBy:       groupBy,
			Filter:        filter,
			NextPageToken: token,
		})
		if err != nil {
			return 0, "", nil, fmt.Errorf("cost explorer GetCostAndUsage: %w", err)
		}

		for _, row := range out.ResultsByTime {
			for _, g := range row.Groups {
				if len(g.Keys) == 0 {
					continue
				}
				key := g.Keys[0]
				m, ok := g.Metrics[MetricNetAmortized]
				if !ok {
					continue
				}
				amt, err := strconv.ParseFloat(aws.ToString(m.Amount), 64)
				if err != nil {
					return 0, "", nil, fmt.Errorf("parse %s amount for %q: %w", MetricNetAmortized, key, err)
				}
				byKey[key] += amt
				if u := aws.ToString(m.Unit); u != "" {
					currency = u
				}
			}
		}

		if out.NextPageToken == nil || *out.NextPageToken == "" {
			break
		}
		token = out.NextPageToken
	}

	breakdown := make([]CostBreakdownItem, 0, len(byKey))
	var total float64
	for key, amt := range byKey {
		if amt == 0 {
			continue
		}
		item := CostBreakdownItem{Amount: amt}
		switch splitBy {
		case SplitByAccount:
			item.Account = key
		default:
			item.Service = key
		}
		breakdown = append(breakdown, item)
		total += amt
	}
	sort.Slice(breakdown, func(i, j int) bool {
		return breakdown[i].Amount > breakdown[j].Amount
	})

	return total, currency, breakdown, nil
}
