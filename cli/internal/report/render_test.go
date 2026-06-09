package report

import (
	"bytes"
	"html/template"
	"strings"
	"testing"
	"time"

	corereport "github.com/openshift-online/finops-tools/core/report"
	coresp "github.com/openshift-online/finops-tools/core/report/savingsplans"
	"github.com/openshift-online/finops-tools/core/cost"
)

func TestRenderCostsHTML(t *testing.T) {
	var buf bytes.Buffer
	err := RenderCostsHTML(&buf, corereport.CostsReport{
		GeneratedAt: time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
		StartDate:   "2026-04-25",
		EndDate:     "2026-05-24",
		Currency:    "USD",
		Metric:      "NetAmortizedCost",
		Total:       1000,
		ByAccount: []cost.CostBreakdownItem{
			{Account: "111111111111", AccountName: "Member", Amount: 600},
		},
		ByService: []cost.CostBreakdownItem{
			{Service: "Amazon EC2", Amount: 700},
		},
		Daily: []cost.DailyCostItem{
			{Date: "2026-05-23", Amount: 30},
			{Date: "2026-05-24", Amount: 40},
		},
		Accounts: []cost.AccountTarget{{
			AccountID:   "123456789012",
			DisplayName: "RH Control Production",
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"Costs Report",
		"RH Control Production",
		"USD 1,000.00",
		"Member",
		"Amazon EC2",
		`<svg class="daily-chart"`,
		"2026-05-23",
		"2026-05-24",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestRenderSavingsPlansHTML(t *testing.T) {
	// StartDate/EndDate use YYYY-MM month labels, matching core/report/savingsplans.Build().
	var buf bytes.Buffer
	err := RenderSavingsPlansHTML(&buf, coresp.Report{
		GeneratedAt: time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
		StartDate:   "2026-01",
		EndDate:     "2026-03",
		Coverage: []coresp.MonthlyMetric{
			{Month: "2026-01", Percentage: 85.0},
			{Month: "2026-02", Percentage: 65.0},
		},
		Utilization: []coresp.MonthlyMetric{
			{Month: "2026-01", Percentage: 92.0},
			{Month: "2026-02", Percentage: 55.0},
		},
	}, "RH Control Production")
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"Savings Plans Report",
		"RH Control Production",
		"Coverage",
		"Utilization",
		"85.0%",
		"65.0%",
		"92.0%",
		"55.0%",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
	if !strings.Contains(out, "<strong>Period:</strong> 2026-01 — 2026-03") {
		t.Errorf("period line should use YYYY-MM labels from Build(); got excerpt around Period:\n%s", excerptAround(out, "Period:"))
	}
	if strings.Contains(out, "2026-01-01") || strings.Contains(out, "2026-03-31") {
		t.Error("period should not use YYYY-MM-DD day labels")
	}
	for _, want := range []string{"Good", "Low", "Critical"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing status %q", want)
		}
	}
}

func TestNewSavingsPlansReportView_statusThresholds(t *testing.T) {
	view := NewSavingsPlansReportView(coresp.Report{
		Coverage: []coresp.MonthlyMetric{
			{Month: "2026-01", Percentage: 85.0},
			{Month: "2026-02", Percentage: 65.0},
			{Month: "2026-03", Percentage: 50.0},
		},
		Utilization: []coresp.MonthlyMetric{
			{Month: "2026-01", Percentage: 92.0},
			{Month: "2026-02", Percentage: 75.0},
			{Month: "2026-03", Percentage: 55.0},
		},
	}, "test")

	assertStatusLabel(t, view.Coverage[0].StatusHTML, "Good")
	assertStatusLabel(t, view.Coverage[1].StatusHTML, "Low")
	assertStatusLabel(t, view.Coverage[2].StatusHTML, "Critical")
	assertStatusLabel(t, view.Utilization[0].StatusHTML, "Good")
	assertStatusLabel(t, view.Utilization[1].StatusHTML, "Low")
	assertStatusLabel(t, view.Utilization[2].StatusHTML, "Critical")
}

func assertStatusLabel(t *testing.T, html template.HTML, want string) {
	t.Helper()
	if !strings.Contains(string(html), want) {
		t.Errorf("status HTML = %q, want label %q", html, want)
	}
}

func excerptAround(s, needle string) string {
	i := strings.Index(s, needle)
	if i < 0 {
		return "(not found)"
	}
	end := i + 80
	if end > len(s) {
		end = len(s)
	}
	return s[i:end]
}

func TestFormatAccountSummary(t *testing.T) {
	s := formatAccountSummary([]cost.AccountTarget{{
		DisplayName: "Quay Production",
		AccountID:   "111111111111",
	}})
	if s != "Quay Production" {
		t.Errorf("got %q", s)
	}
}

func TestFormatAccountSummaryFallsBackToAccountID(t *testing.T) {
	s := formatAccountSummary([]cost.AccountTarget{{
		DisplayAlias: "quay",
		AccountID:    "111111111111",
	}})
	if s != "111111111111" {
		t.Errorf("got %q", s)
	}
}
