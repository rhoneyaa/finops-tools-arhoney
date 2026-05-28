package execsummary

import (
	"testing"
	"time"
)

func TestMonthLabel(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "January 2026",
			input:    time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
			expected: "2026-01",
		},
		{
			name:     "December 2025",
			input:    time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: "2025-12",
		},
		{
			name:     "First day of month",
			input:    time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			expected: "2026-05",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MonthLabel(tt.input)
			if result != tt.expected {
				t.Errorf("MonthLabel(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMonthName(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string
	}{
		{
			name:     "January 2026",
			input:    time.Date(2026, 1, 15, 10, 30, 0, 0, time.UTC),
			expected: "Jan 2026",
		},
		{
			name:     "December 2025",
			input:    time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC),
			expected: "Dec 2025",
		},
		{
			name:     "May 2026",
			input:    time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC),
			expected: "May 2026",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MonthName(tt.input)
			if result != tt.expected {
				t.Errorf("MonthName(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMonthsInWindow(t *testing.T) {
	tests := []struct {
		name     string
		start    time.Time
		end      time.Time
		expected []string // Month labels for easier comparison
	}{
		{
			name:  "Single month",
			start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			expected: []string{"2026-01"},
		},
		{
			name:  "Three months",
			start: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC),
			expected: []string{"2026-01", "2026-02", "2026-03"},
		},
		{
			name:  "Year boundary",
			start: time.Date(2025, 11, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			expected: []string{"2025-11", "2025-12", "2026-01"},
		},
		{
			name:  "Start not on first day",
			start: time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
			expected: []string{"2026-01", "2026-02"},
		},
		{
			name:     "Same month",
			start:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MonthsInWindow(tt.start, tt.end)
			if len(result) != len(tt.expected) {
				t.Errorf("MonthsInWindow() returned %d months, want %d", len(result), len(tt.expected))
				return
			}
			for i, month := range result {
				label := MonthLabel(month)
				if label != tt.expected[i] {
					t.Errorf("MonthsInWindow()[%d] = %q, want %q", i, label, tt.expected[i])
				}
			}
		})
	}
}

func TestComputeWindow(t *testing.T) {
	tests := []struct {
		name          string
		nMonths       int
		today         *time.Time
		wantStart     string // Month label
		wantEnd       string // Month label
		wantMonthCnt  int
		wantErr       bool
	}{
		{
			name:         "Single month",
			nMonths:      1,
			today:        timePtr(time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)),
			wantStart:    "2026-05",
			wantEnd:      "2026-05",
			wantMonthCnt: 1,
		},
		{
			name:         "Three months",
			nMonths:      3,
			today:        timePtr(time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)),
			wantStart:    "2026-03",
			wantEnd:      "2026-05",
			wantMonthCnt: 3,
		},
		{
			name:         "Twelve months",
			nMonths:      12,
			today:        timePtr(time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC)),
			wantStart:    "2025-06",
			wantEnd:      "2026-05",
			wantMonthCnt: 12,
		},
		{
			name:         "Year boundary",
			nMonths:      6,
			today:        timePtr(time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)),
			wantStart:    "2025-09",
			wantEnd:      "2026-02",
			wantMonthCnt: 6,
		},
		{
			name:    "Invalid nMonths zero",
			nMonths: 0,
			today:   timePtr(time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)),
			wantErr: true,
		},
		{
			name:    "Invalid nMonths negative",
			nMonths: -5,
			today:   timePtr(time.Date(2026, 5, 28, 0, 0, 0, 0, time.UTC)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			window, err := ComputeWindow(tt.nMonths, tt.today)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ComputeWindow() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("ComputeWindow() unexpected error: %v", err)
			}

			startLabel := MonthLabel(window.Start)
			if startLabel != tt.wantStart {
				t.Errorf("ComputeWindow() start = %q, want %q", startLabel, tt.wantStart)
			}

			endLabel := MonthLabel(window.End)
			if endLabel != tt.wantEnd {
				t.Errorf("ComputeWindow() end = %q, want %q", endLabel, tt.wantEnd)
			}

			if len(window.Months) != tt.wantMonthCnt {
				t.Errorf("ComputeWindow() months count = %d, want %d", len(window.Months), tt.wantMonthCnt)
			}

			// Verify first and last months
			if len(window.Months) > 0 {
				firstLabel := MonthLabel(window.Months[0])
				if firstLabel != tt.wantStart {
					t.Errorf("First month in Months = %q, want %q", firstLabel, tt.wantStart)
				}
				lastLabel := MonthLabel(window.Months[len(window.Months)-1])
				if lastLabel != tt.wantEnd {
					t.Errorf("Last month in Months = %q, want %q", lastLabel, tt.wantEnd)
				}
			}
		})
	}
}

func TestNextMonth(t *testing.T) {
	tests := []struct {
		name     string
		input    time.Time
		expected string // Month label
	}{
		{
			name:     "January to February",
			input:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expected: "2026-02",
		},
		{
			name:     "December to January (year wrap)",
			input:    time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			expected: "2026-01",
		},
		{
			name:     "Mid-month start",
			input:    time.Date(2026, 5, 15, 0, 0, 0, 0, time.UTC),
			expected: "2026-06",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := nextMonth(tt.input)
			label := MonthLabel(result)
			if label != tt.expected {
				t.Errorf("nextMonth(%v) = %q, want %q", tt.input, label, tt.expected)
			}
			// Verify it's always the first of the month
			if result.Day() != 1 {
				t.Errorf("nextMonth(%v) day = %d, want 1", tt.input, result.Day())
			}
		})
	}
}

func TestShiftMonths(t *testing.T) {
	tests := []struct {
		name         string
		input        time.Time
		deltaMonths  int
		expected     string // Month label
	}{
		{
			name:        "Forward 1 month",
			input:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			deltaMonths: 1,
			expected:    "2026-02",
		},
		{
			name:        "Backward 1 month",
			input:       time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			deltaMonths: -1,
			expected:    "2026-01",
		},
		{
			name:        "Forward 12 months",
			input:       time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC),
			deltaMonths: 12,
			expected:    "2026-05",
		},
		{
			name:        "Backward 12 months",
			input:       time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			deltaMonths: -12,
			expected:    "2025-05",
		},
		{
			name:        "Year wrap forward",
			input:       time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			deltaMonths: 2,
			expected:    "2026-02",
		},
		{
			name:        "Year wrap backward",
			input:       time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			deltaMonths: -2,
			expected:    "2025-11",
		},
		{
			name:        "No change",
			input:       time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC),
			deltaMonths: 0,
			expected:    "2026-05",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := shiftMonths(tt.input, tt.deltaMonths)
			label := MonthLabel(result)
			if label != tt.expected {
				t.Errorf("shiftMonths(%v, %d) = %q, want %q", tt.input, tt.deltaMonths, label, tt.expected)
			}
			// Verify it's always the first of the month
			if result.Day() != 1 {
				t.Errorf("shiftMonths(%v, %d) day = %d, want 1", tt.input, tt.deltaMonths, result.Day())
			}
		})
	}
}

// timePtr is a helper to create *time.Time for test cases.
func timePtr(t time.Time) *time.Time {
	return &t
}
