// activityday.go is the component main file: it defines the fixed activity
// timezone and the natural-day key helpers shared by every per-day gameplay
// boundary.

// Package activityday pins every per-day gameplay boundary (the activation
// limits, the daily check-in, the steal/gift counters and the stealable-list
// seed) to Beijing time (UTC+8) as required by the activity rules, so the
// natural-day key never depends on the host server timezone. A fixed offset is
// used deliberately: the campaign is mainland-only and a fixed zone works
// without tzdata on minimal deployment images.
package activityday

import "time"

// dateLayout is the YYYY-MM-DD natural-day key layout shared by daily limits.
const dateLayout = "2006-01-02"

// beijing is the fixed UTC+8 activity timezone.
var beijing = time.FixedZone("CST", 8*60*60)

// Date returns t's natural-day key (YYYY-MM-DD) in the activity timezone.
func Date(t time.Time) string {
	return t.In(beijing).Format(dateLayout)
}

// Today returns the current natural-day key in the activity timezone.
func Today() string {
	return Date(time.Now())
}

// DayBounds returns the half-open [start, end) instants of t's natural day in
// the activity timezone, for range queries over timestamp columns.
func DayBounds(t time.Time) (start time.Time, end time.Time) {
	local := t.In(beijing)
	start = time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, beijing)
	return start, start.Add(24 * time.Hour)
}
