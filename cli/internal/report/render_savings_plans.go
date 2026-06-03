package report

import (
	"fmt"
	"io"
	"sync"

	"github.com/nikolalohinski/gonja/v2"
	"github.com/nikolalohinski/gonja/v2/config"
	"github.com/nikolalohinski/gonja/v2/exec"
	"github.com/nikolalohinski/gonja/v2/loaders"
	coresp "github.com/openshift-online/finops-tools/core/report/savingsplans"
)

const savingsPlansTemplate = "savings-plans.html.j2"

var (
	spTplOnce sync.Once
	spTpl     *exec.Template
	spTplErr  error
)

func savingsPlansTemplateCompiled() (*exec.Template, error) {
	spTplOnce.Do(func() {
		embedLoader, err := loaders.NewEmbedFSLoader("templates", &templateFS)
		if err != nil {
			spTplErr = err
			return
		}
		loader := newStableLoader(embedLoader)
		cfg := config.New()
		environment := &exec.Environment{
			Context:           exec.EmptyContext(),
			Filters:           gonja.DefaultEnvironment.Filters,
			Tests:             gonja.DefaultEnvironment.Tests,
			ControlStructures: gonja.DefaultEnvironment.ControlStructures,
			Methods:           gonja.DefaultEnvironment.Methods,
		}
		spTpl, spTplErr = exec.NewTemplate(savingsPlansTemplate, cfg, loader, environment)
	})
	return spTpl, spTplErr
}

// SavingsPlansMetricView is one row in the coverage or utilization table.
type SavingsPlansMetricView struct {
	Month               string
	Percentage          float64
	PercentageFormatted string
}

// SavingsPlansReportView is the template context for savings-plans.html.j2.
type SavingsPlansReportView struct {
	GeneratedAt    string
	AccountSummary string
	StartDate      string
	EndDate        string
	Coverage       []SavingsPlansMetricView
	Utilization    []SavingsPlansMetricView
}

// NewSavingsPlansReportView maps a core savingsplans.Report to the template context.
func NewSavingsPlansReportView(r coresp.Report, accountSummary string) SavingsPlansReportView {
	return SavingsPlansReportView{
		GeneratedAt:    r.GeneratedAt.UTC().Format("2006-01-02 15:04:05 UTC"),
		AccountSummary: accountSummary,
		StartDate:      r.StartDate,
		EndDate:        r.EndDate,
		Coverage:       metricsToView(r.Coverage),
		Utilization:    metricsToView(r.Utilization),
	}
}

func metricsToView(metrics []coresp.MonthlyMetric) []SavingsPlansMetricView {
	rows := make([]SavingsPlansMetricView, 0, len(metrics))
	for _, m := range metrics {
		rows = append(rows, SavingsPlansMetricView{
			Month:               m.Month,
			Percentage:          m.Percentage,
			PercentageFormatted: fmt.Sprintf("%.1f%%", m.Percentage),
		})
	}
	return rows
}

// RenderSavingsPlansHTML renders the savings plans report as HTML to w.
func RenderSavingsPlansHTML(w io.Writer, r coresp.Report, accountSummary string) error {
	t, err := savingsPlansTemplateCompiled()
	if err != nil {
		return fmt.Errorf("compile savings-plans template: %w", err)
	}
	view := NewSavingsPlansReportView(r, accountSummary)
	ctx := exec.NewContext(map[string]any{
		"generated_at":    view.GeneratedAt,
		"account_summary": view.AccountSummary,
		"start_date":      view.StartDate,
		"end_date":        view.EndDate,
		"coverage":        view.Coverage,
		"utilization":     view.Utilization,
	})
	if err := t.Execute(w, ctx); err != nil {
		return fmt.Errorf("render savings-plans template: %w", err)
	}
	return nil
}
