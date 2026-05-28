package execsummary

import (
	"fmt"
	"time"
)

// MonthLabel formats a date as YYYY-MM for use as a month key.
func MonthLabel(t time.Time) string {
	return t.Format("2006-01")
}

// MonthName formats a date as abbreviated month + year (e.g., "Jan 2026").
func MonthName(t time.Time) string {
	return t.Format("Jan 2006")
}

// MonthsInWindow returns the first day of each month from start up to (but excluding) end.
// Both start and end are normalized to the first day of their respective months.
func MonthsInWindow(start, end time.Time) []time.Time {
	// Normalize to first day of month
	current := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
	endNorm := time.Date(end.Year(), end.Month(), 1, 0, 0, 0, 0, time.UTC)

	var months []time.Time
	for current.Before(endNorm) {
		months = append(months, current)
		current = nextMonth(current)
	}
	return months
}

// ComputeWindow calculates a time window of nMonths ending at the first day of the current month.
// If today is nil, uses time.Now().
// Returns (start, end) where end is the first day of the current month.
func ComputeWindow(nMonths int, today *time.Time) (TimeWindow, error) {
	if nMonths < 1 {
		return TimeWindow{}, fmt.Errorf("nMonths must be >= 1, got %d", nMonths)
	}

	reference := time.Now().UTC()
	if today != nil {
		reference = *today
	}

	// End is first day of current month
	end := time.Date(reference.Year(), reference.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Start is nMonths - 1 months before end
	start := shiftMonths(end, -(nMonths - 1))

	// MonthsInWindow excludes end, so we need to include the end month
	months := MonthsInWindow(start, nextMonth(end))

	return TimeWindow{
		Start:  start,
		End:    end,
		Months: months,
	}, nil
}

// nextMonth returns the first day of the next month.
func nextMonth(t time.Time) time.Time {
	// Add 32 days and normalize to first day
	next := t.AddDate(0, 0, 32)
	return time.Date(next.Year(), next.Month(), 1, 0, 0, 0, 0, time.UTC)
}

// shiftMonths shifts a date by deltaMonths months while preserving day=1.
func shiftMonths(t time.Time, deltaMonths int) time.Time {
	year := t.Year()
	month := int(t.Month()) + deltaMonths

	// Handle year wrapping
	for month <= 0 {
		month += 12
		year--
	}
	for month > 12 {
		month -= 12
		year++
	}

	return time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
}
