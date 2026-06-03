// activation_map.go implements the player-facing visible-cattle map list and the
// release-schedule visibility computation. Listing pushes the stage/online-time
// filter to the database, then applies the optional weekday/time-window match to
// the bounded result set in memory, and batch-assembles the current player's
// activation flags in one query keyed by the visible cattle IDs to avoid N+1.

package activation

import (
	"context"
	"strconv"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// visibleNiuCap bounds the visible-cattle map page. The activity has ≤120 cattle,
// so the full visible set is returned in one bounded response.
const visibleNiuCap = 200

// VisibleNiuItem defines one visible cattle projected for the player map.
type VisibleNiuItem struct {
	// Id is the cattle ID.
	Id int64
	// Code is the cattle serial code.
	Code string
	// NiuType is the cattle type string.
	NiuType string
	// Name is the cattle name; empty for common cattle.
	Name string
	// Lat is the GPS latitude anchor.
	Lat float64
	// Lng is the GPS longitude anchor.
	Lng float64
	// Status is the shared-pool activation status string.
	Status string
	// ActivatedByMe reports whether the current player has already activated this cattle.
	ActivatedByMe bool
}

// VisibleNiu returns the cattle currently visible to playerID.
func (s *serviceImpl) VisibleNiu(ctx context.Context, playerID int64) ([]*VisibleNiuItem, error) {
	now := time.Now()

	rows := make([]*entitymodel.Niu, 0)
	err := dao.Niu.Ctx(ctx).
		WhereNot(dao.Niu.Columns().ReleaseStage, "").
		Where(dao.Niu.Columns().OnlineAt+" IS NOT NULL").
		WhereLTE(dao.Niu.Columns().OnlineAt, now).
		OrderAsc(dao.Niu.Columns().Id).
		Limit(visibleNiuCap).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}

	visible := make([]*entitymodel.Niu, 0, len(rows))
	for _, row := range rows {
		if isWithinVisibleWindow(row, now) {
			visible = append(visible, row)
		}
	}

	activatedByMe, err := s.batchActivatedByMe(ctx, playerID, visible)
	if err != nil {
		return nil, err
	}

	list := make([]*VisibleNiuItem, 0, len(visible))
	for _, row := range visible {
		list = append(list, &VisibleNiuItem{
			Id:            row.Id,
			Code:          row.Code,
			NiuType:       row.NiuType,
			Name:          row.Name,
			Lat:           row.Lat,
			Lng:           row.Lng,
			Status:        row.Status,
			ActivatedByMe: activatedByMe[row.Id],
		})
	}
	return list, nil
}

// batchActivatedByMe returns a cattle-ID set the player has activated among the
// visible cattle, fetched in one projected query keyed by the visible IDs. An
// empty visible set or a non-positive player short-circuits without a query.
func (s *serviceImpl) batchActivatedByMe(
	ctx context.Context,
	playerID int64,
	visible []*entitymodel.Niu,
) (map[int64]bool, error) {
	activated := make(map[int64]bool, len(visible))
	if playerID <= 0 || len(visible) == 0 {
		return activated, nil
	}
	ids := make([]int64, 0, len(visible))
	for _, row := range visible {
		ids = append(ids, row.Id)
	}

	records := make([]*entitymodel.Activation, 0, len(ids))
	err := dao.Activation.Ctx(ctx).
		Fields(dao.Activation.Columns().NiuId).
		Where(dao.Activation.Columns().UserId, playerID).
		WhereIn(dao.Activation.Columns().NiuId, ids).
		Scan(&records)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	for _, record := range records {
		activated[record.NiuId] = true
	}
	return activated, nil
}

// isWithinVisibleWindow reports whether the cattle is within its optional visible
// weekday and time window at now. An empty weekday list or an empty window bound
// means that constraint is unrestricted.
func isWithinVisibleWindow(row *entitymodel.Niu, now time.Time) bool {
	if !weekdayMatches(row.VisibleWeekdays, now) {
		return false
	}
	return timeWindowMatches(row.VisibleStart, row.VisibleEnd, now)
}

// weekdayMatches reports whether now's ISO weekday (1=Monday..7=Sunday) is in the
// comma-separated weekday list. An empty list is unrestricted and matches.
func weekdayMatches(weekdays string, now time.Time) bool {
	trimmed := strings.TrimSpace(weekdays)
	if trimmed == "" {
		return true
	}
	current := isoWeekday(now)
	for _, part := range strings.Split(trimmed, ",") {
		token := strings.TrimSpace(part)
		if token == "" {
			continue
		}
		value, err := strconv.Atoi(token)
		if err != nil {
			continue
		}
		if value == current {
			return true
		}
	}
	return false
}

// isoWeekday returns the ISO weekday number for t, with Monday=1 and Sunday=7.
func isoWeekday(t time.Time) int {
	weekday := int(t.Weekday())
	if weekday == 0 {
		return 7
	}
	return weekday
}

// timeWindowMatches reports whether now's local minute-of-day falls within the
// optional HH:MM window [start, end]. Either empty bound leaves that side
// unrestricted; an unparsable bound is treated as unrestricted on that side.
func timeWindowMatches(start, end string, now time.Time) bool {
	current := now.Hour()*60 + now.Minute()
	if startMinute, ok := parseClockMinute(start); ok && current < startMinute {
		return false
	}
	if endMinute, ok := parseClockMinute(end); ok && current > endMinute {
		return false
	}
	return true
}

// parseClockMinute parses an HH:MM clock string into a minute-of-day. The ok
// result is false when the value is empty or not a valid HH:MM, so the caller
// treats that bound as unrestricted.
func parseClockMinute(value string) (int, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return 0, false
	}
	parts := strings.Split(trimmed, ":")
	if len(parts) != 2 {
		return 0, false
	}
	hour, err := strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil || hour < 0 || hour > 23 {
		return 0, false
	}
	minute, err := strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil || minute < 0 || minute > 59 {
		return 0, false
	}
	return hour*60 + minute, true
}
