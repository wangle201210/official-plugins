// activation_map_test.go covers the online-time visibility helpers (weekday and
// time-window matching) used to filter the visible-cattle map list. These are
// pure-logic tests with no database dependency.

package activation

import (
	"testing"
	"time"

	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// mondayNoon is a fixed reference instant (a Monday at 12:00) used to make the
// weekday and time-window assertions deterministic and timezone-stable.
var mondayNoon = time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

// TestWeekdayMatchesEmptyIsUnrestricted verifies an empty weekday list matches.
func TestWeekdayMatchesEmptyIsUnrestricted(t *testing.T) {
	if !weekdayMatches("", mondayNoon) {
		t.Fatal("expected empty weekday list to be unrestricted")
	}
}

// TestWeekdayMatchesIncludedAndExcluded verifies the ISO weekday membership test.
func TestWeekdayMatchesIncludedAndExcluded(t *testing.T) {
	// 2024-01-01 is a Monday => ISO weekday 1.
	if !weekdayMatches("1,3,5", mondayNoon) {
		t.Fatal("expected Monday to match weekday list 1,3,5")
	}
	if weekdayMatches("2,4,6", mondayNoon) {
		t.Fatal("expected Monday not to match weekday list 2,4,6")
	}
}

// TestTimeWindowMatches verifies inclusive bounds and unrestricted empty bounds.
func TestTimeWindowMatches(t *testing.T) {
	if !timeWindowMatches("", "", mondayNoon) {
		t.Fatal("expected empty window to be unrestricted")
	}
	if !timeWindowMatches("09:00", "18:00", mondayNoon) {
		t.Fatal("expected noon to fall within 09:00-18:00")
	}
	if timeWindowMatches("13:00", "18:00", mondayNoon) {
		t.Fatal("expected noon to be before a 13:00 window start")
	}
	if timeWindowMatches("06:00", "11:00", mondayNoon) {
		t.Fatal("expected noon to be after an 11:00 window end")
	}
}

// TestNiuCurrentlyVisibleCombines verifies the combined online-time, weekday and
// time-window gate used per visible cattle.
func TestNiuCurrentlyVisibleCombines(t *testing.T) {
	past := mondayNoon.Add(-time.Hour)
	future := mondayNoon.Add(time.Hour)
	row := &entitymodel.Niu{
		OnlineAt:        &past,
		VisibleWeekdays: "1",
		VisibleStart:    "09:00",
		VisibleEnd:      "18:00",
	}
	if !niuCurrentlyVisible(row, mondayNoon) {
		t.Fatal("expected Monday noon to be within the visible window")
	}

	rowWrongDay := &entitymodel.Niu{OnlineAt: &past, VisibleWeekdays: "2"}
	if niuCurrentlyVisible(rowWrongDay, mondayNoon) {
		t.Fatal("expected a non-matching weekday to be invisible")
	}

	rowFuture := &entitymodel.Niu{OnlineAt: &future}
	if niuCurrentlyVisible(rowFuture, mondayNoon) {
		t.Fatal("expected a future online time to be invisible")
	}

	rowUnset := &entitymodel.Niu{}
	if niuCurrentlyVisible(rowUnset, mondayNoon) {
		t.Fatal("expected an unset online time to be invisible")
	}
}
