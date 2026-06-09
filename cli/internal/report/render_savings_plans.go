package report

import (
	"fmt"
	"html/template"
	"io"
	"sync"

	coresp "github.com/openshift-online/finops-tools/core/report/savingsplans"
)

const savingsPlansTemplate = "savings-plans"

var (
	spTplOnce sync.Once
	spTpl     *template.Template
	spTplErr  error
)

func savingsPlansTemplateCompiled() (*template.Template, error) {
	spTplOnce.Do(func() {
		spTpl, spTplErr = template.ParseFS(templateFS,
			"templates/layout.html",
			"templates/savings-plans.html",
		)
	})
	return spTpl, spTplErr
}

// SavingsPlansMetricView is one row in the coverage or utilization table.
type SavingsPlansMetricView struct {
	Month               string
	Percentage          float64
	PercentageFormatted string
	StatusHTML          template.HTML
}

// SavingsPlansReportView is the template context for savings-plans.html.
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
		Coverage:       coverageMetricsToView(r.Coverage),
		Utilization:    utilizationMetricsToView(r.Utilization),
	}
}

func coverageMetricsToView(metrics []coresp.MonthlyMetric) []SavingsPlansMetricView {
	rows := make([]SavingsPlansMetricView, 0, len(metrics))
	for _, m := range metrics {
		rows = append(rows, SavingsPlansMetricView{
			Month:               m.Month,
			Percentage:          m.Percentage,
			PercentageFormatted: fmt.Sprintf("%.1f%%", m.Percentage),
			StatusHTML:          coverageStatusHTML(m.Percentage),
		})
	}
	return rows
}

func utilizationMetricsToView(metrics []coresp.MonthlyMetric) []SavingsPlansMetricView {
	rows := make([]SavingsPlansMetricView, 0, len(metrics))
	for _, m := range metrics {
		rows = append(rows, SavingsPlansMetricView{
			Month:               m.Month,
			Percentage:          m.Percentage,
			PercentageFormatted: fmt.Sprintf("%.1f%%", m.Percentage),
			StatusHTML:          utilizationStatusHTML(m.Percentage),
		})
	}
	return rows
}

func coverageStatusHTML(pct float64) template.HTML {
	switch {
	case pct >= 80:
		return `<span style="color:#22c55e">&#x2713; Good</span>`
	case pct >= 60:
		return `<span style="color:#f59e0b">&#x26A0; Low</span>`
	default:
		return `<span style="color:#ef4444">&#x2717; Critical</span>`
	}
}

func utilizationStatusHTML(pct float64) template.HTML {
	switch {
	case pct >= 90:
		return `<span style="color:#22c55e">&#x2713; Good</span>`
	case pct >= 70:
		return `<span style="color:#f59e0b">&#x26A0; Low</span>`
	default:
		return `<span style="color:#ef4444">&#x2717; Critical</span>`
	}
}

// RenderSavingsPlansHTML renders the savings plans report as HTML to w.
func RenderSavingsPlansHTML(w io.Writer, r coresp.Report, accountSummary string) error {
	t, err := savingsPlansTemplateCompiled()
	if err != nil {
		return fmt.Errorf("compile savings-plans template: %w", err)
	}
	view := NewSavingsPlansReportView(r, accountSummary)
	if err := t.ExecuteTemplate(w, savingsPlansTemplate, view); err != nil {
		return fmt.Errorf("render savings-plans template: %w", err)
	}
	return nil
}
