package report

import (
	"context"
	"fmt"

	coreaccount "github.com/openshift-online/finops-tools/core/account"
	"github.com/openshift-online/finops-tools/core/cost"
	corereport "github.com/openshift-online/finops-tools/core/report"
)

type costsGenerator struct{}

func (costsGenerator) Validate(in GenerateInput) error {
	return validateTemplateFormat(TemplateCosts, in.Format)
}

func (costsGenerator) Generate(ctx context.Context, in GenerateInput) error {
	if len(in.Targets) == 0 {
		report := corereport.EmptyCostsReport(cost.CostQuery{
			Provider: cost.ProviderAWS,
			Range:    in.Range,
		}, in.Now)
		in.Progress.Step("Rendering HTML report…")
		return RenderCostsHTML(in.Out, report)
	}

	if len(in.Targets) > 1 {
		in.Progress.Step(fmt.Sprintf("Fetching net amortized costs for %d account(s) from AWS Cost Explorer…", len(in.Targets)))
	}

	costQuery := cost.CostQuery{
		Provider: cost.ProviderAWS,
		Accounts: in.Targets,
		Range:    in.Range,
		Progress: in.Progress,
		AWSFetch: &cost.AWSFetchOptions{
			ResolveAccountNames: coreaccount.ResolveAccountNames,
		},
	}

	report, err := corereport.BuildCostsReport(ctx, costQuery, in.Progress)
	if err != nil {
		return err
	}
	in.Progress.Step("Rendering HTML report…")
	return RenderCostsHTML(in.Out, report)
}
