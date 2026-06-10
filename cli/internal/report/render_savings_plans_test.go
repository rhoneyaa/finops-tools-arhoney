package report

import "testing"

func TestMonthDisplayLabel(t *testing.T) {
	tests := []struct {
		month, start, end, want string
	}{
		{"2026-02", "2026-01-01", "2026-06-30", "2026-02"},
		{"2026-01", "2026-01-01", "2026-06-30", "2026-01"},
		{"2026-01", "2026-01-15", "2026-06-08", "2026-01 (from 15)"},
		{"2026-06", "2026-01-15", "2026-06-08", "2026-06 (through 8)"},
		{"2026-03", "2026-01-15", "2026-06-08", "2026-03"},
		{"2026-01", "2026-01-15", "2026-01-28", "2026-01 (15 – 28)"},
		{"2026-02", "2026-02-01", "2026-02-28", "2026-02"},
		{"2026-02", "2026-02-01", "2026-02-15", "2026-02 (through 15)"},
	}
	for _, tt := range tests {
		got := monthDisplayLabel(tt.month, tt.start, tt.end)
		if got != tt.want {
			t.Errorf("monthDisplayLabel(%q, %q, %q) = %q, want %q", tt.month, tt.start, tt.end, got, tt.want)
		}
	}
}

func TestMonthDisplayLabel_invalidDates(t *testing.T) {
	if got := monthDisplayLabel("2026-01", "bad", "2026-06-08"); got != "2026-01" {
		t.Errorf("got %q", got)
	}
}
