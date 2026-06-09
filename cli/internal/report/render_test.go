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
	// StartDate/EndDate use full calendar dates, matching core/report/savingsplans.Build().
	var buf bytes.Buffer
	err := RenderSavingsPlansHTML(&buf, coresp.Report{
		GeneratedAt: time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
		StartDate:   "2026-01-15",
		EndDate:     "2026-06-08",
		Accounts: []coresp.AccountReport{
			{
				AccountName: "RH Control Production",
				Coverage: []coresp.MonthlyMetric{
					{Month: "2026-01", Percentage: 85.0},
					{Month: "2026-02", Percentage: 65.0},
					{Month: "2026-03", Percentage: 70.0},
					{Month: "2026-06", Percentage: 78.0},
				},
				Utilization: []coresp.MonthlyMetric{
					{Month: "2026-01", Percentage: 92.0},
					{Month: "2026-02", Percentage: 55.0},
					{Month: "2026-06", Percentage: 81.0},
				},
			},
			{
				AccountName: "Member One",
				Coverage: []coresp.MonthlyMetric{
					{Month: "2026-01", Percentage: 72.0},
				},
				Utilization: []coresp.MonthlyMetric{
					{Month: "2026-01", Percentage: 88.0},
				},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	for _, want := range []string{
		"Savings Plans Report",
		"RH Control Production",
		"Member One",
		"Coverage",
		"Utilization",
		"2026-01 (from 15)",
		"2026-03",
		"2026-06 (through 8)",
		"85.0%",
		"65.0%",
		"92.0%",
		"55.0%",
		"72.0%",
		"88.0%",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q", want)
		}
	}
	if !strings.Contains(out, "<strong>Period:</strong> 2026-01-15 — 2026-06-08") {
		t.Errorf("period line should use full dates from Build(); got excerpt around Period:\n%s", excerptAround(out, "Period:"))
	}
	for _, want := range []string{"Good", "Low", "Critical"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing status %q", want)
		}
	}
}

func TestNewSavingsPlansReportView_linkedOmitsStatus(t *testing.T) {
	view := NewSavingsPlansReportView(coresp.Report{
		StartDate: "2026-01-01",
		EndDate:   "2026-03-31",
		Accounts: []coresp.AccountReport{{
			AccountName: "Quay",
			IsLinked:    true,
			Coverage: []coresp.MonthlyMetric{
				{Month: "2026-01", Percentage: 50.0},
			},
		}},
	})
	if !view.Accounts[0].IsLinked {
		t.Fatal("expected linked account view")
	}
	if view.Accounts[0].Coverage[0].StatusHTML != "" {
		t.Errorf("linked coverage status = %q, want empty", view.Accounts[0].Coverage[0].StatusHTML)
	}
}

func TestRenderSavingsPlansHTML_linkedOmitsStatusColumn(t *testing.T) {
	var buf bytes.Buffer
	err := RenderSavingsPlansHTML(&buf, coresp.Report{
		GeneratedAt: time.Date(2026, 5, 26, 10, 0, 0, 0, time.UTC),
		StartDate:   "2026-01-01",
		EndDate:     "2026-03-31",
		Accounts: []coresp.AccountReport{{
			AccountName: "Quay",
			IsLinked:    true,
			Coverage: []coresp.MonthlyMetric{
				{Month: "2026-01", Percentage: 72.0},
			},
		}},
	})
	if err != nil {
		t.Fatal(err)
	}
	out := buf.String()
	if strings.Contains(out, "Critical") || strings.Contains(out, "Good") || strings.Contains(out, "Low") {
		t.Errorf("linked account section should not include status labels; got excerpt:\n%s", excerptAround(out, "Quay"))
	}
	quaySection := excerptAround(out, "Quay")
	if strings.Count(quaySection, "<th>Status</th>") != 0 {
		t.Errorf("linked account should not render Status column header; got:\n%s", quaySection)
	}
}

func TestNewSavingsPlansReportView_statusThresholds(t *testing.T) {
	view := NewSavingsPlansReportView(coresp.Report{
		StartDate: "2026-01-01",
		EndDate:   "2026-03-31",
		Accounts: []coresp.AccountReport{{
			AccountName: "test",
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
		}},
	})

	acct := view.Accounts[0]
	assertStatusLabel(t, acct.Coverage[0].StatusHTML, "Good")
	assertStatusLabel(t, acct.Coverage[1].StatusHTML, "Low")
	assertStatusLabel(t, acct.Coverage[2].StatusHTML, "Critical")
	assertStatusLabel(t, acct.Utilization[0].StatusHTML, "Good")
	assertStatusLabel(t, acct.Utilization[1].StatusHTML, "Low")
	assertStatusLabel(t, acct.Utilization[2].StatusHTML, "Critical")
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
