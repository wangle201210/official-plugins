// activityday_test.go covers the Beijing-time natural-day key and day-bounds
// helpers with fixed instants so the assertions never depend on the timezone of
// the machine running the tests.

package activityday

import (
	"testing"
	"time"
)

// TestDateUsesBeijingDay asserts the natural-day key is derived in UTC+8
// regardless of the instant's original location.
func TestDateUsesBeijingDay(t *testing.T) {
	// 2026-06-11 17:30 UTC is already 2026-06-12 01:30 in Beijing.
	instant := time.Date(2026, 6, 11, 17, 30, 0, 0, time.UTC)
	if got := Date(instant); got != "2026-06-12" {
		t.Fatalf("expected Beijing day 2026-06-12, got %s", got)
	}

	// 2026-06-11 15:59 UTC is still 2026-06-11 23:59 in Beijing.
	instant = time.Date(2026, 6, 11, 15, 59, 0, 0, time.UTC)
	if got := Date(instant); got != "2026-06-11" {
		t.Fatalf("expected Beijing day 2026-06-11, got %s", got)
	}
}

// TestDayBoundsHalfOpen asserts DayBounds returns the Beijing midnight start and
// the next midnight as a half-open interval containing the input instant.
func TestDayBoundsHalfOpen(t *testing.T) {
	instant := time.Date(2026, 6, 11, 17, 30, 0, 0, time.UTC)
	start, end := DayBounds(instant)

	if got := Date(start); got != "2026-06-12" {
		t.Fatalf("expected day start on 2026-06-12, got %s", got)
	}
	if !start.Before(instant) && !start.Equal(instant) {
		t.Fatalf("expected start %v not after instant %v", start, instant)
	}
	if !instant.Before(end) {
		t.Fatalf("expected instant %v before end %v", instant, end)
	}
	if got := end.Sub(start); got != 24*time.Hour {
		t.Fatalf("expected 24h day span, got %v", got)
	}
	if got := Date(end); got != "2026-06-13" {
		t.Fatalf("expected end to land on the next Beijing day, got %s", got)
	}
}
