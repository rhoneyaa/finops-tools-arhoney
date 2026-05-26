// period.go defines inclusive/exclusive date ranges for cost queries (e.g. last N calendar days).
package cost

import "time"

// DateRange holds inclusive start and exclusive end calendar dates (YYYY-MM-DD).
type DateRange struct {
	Start time.Time
	End   time.Time
}

// LastNDaysRange returns a range covering the last n calendar days ending today (UTC).
// Start is inclusive; End is exclusive (Cost Explorer TimePeriod convention).
func LastNDaysRange(n int, now time.Time) DateRange {
	if n <= 0 {
		n = DefaultDays
	}
	utc := now.UTC()
	end := dateOnly(utc)
	start := end.AddDate(0, 0, -n)
	return DateRange{Start: start, End: end}
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func formatDate(t time.Time) string {
	return t.Format("2006-01-02")
}
