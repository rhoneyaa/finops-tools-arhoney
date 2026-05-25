package cost

import (
	"testing"
	"time"
)

func TestLastNDaysRange(t *testing.T) {
	now := time.Date(2026, 5, 25, 15, 30, 0, 0, time.UTC)
	dr := LastNDaysRange(30, now)

	wantStart := "2026-04-25"
	wantEnd := "2026-05-25"

	if got := formatDate(dr.Start); got != wantStart {
		t.Errorf("Start = %q, want %q", got, wantStart)
	}
	if got := formatDate(dr.End); got != wantEnd {
		t.Errorf("End (exclusive) = %q, want %q", got, wantEnd)
	}
}

func TestLastNDaysRangeDefaultsOnZero(t *testing.T) {
	now := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)
	dr := LastNDaysRange(0, now)
	if formatDate(dr.End) != "2026-01-31" {
		t.Fatalf("unexpected end: %s", formatDate(dr.End))
	}
	if formatDate(dr.Start) != "2026-01-01" {
		t.Fatalf("unexpected start: %s", formatDate(dr.Start))
	}
}
