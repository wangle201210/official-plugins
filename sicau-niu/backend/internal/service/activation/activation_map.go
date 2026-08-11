// activation_map.go implements the player-facing visible-cattle map list and the
// online-time visibility computation. Listing pushes the online-time filter to
// the database, then applies the optional weekday/time-window match to the
// bounded result set in memory, and batch-assembles the current player's
// activation flags in one query keyed by the visible cattle IDs to avoid N+1.

package activation

import (
	"context"
	"math"
	"strconv"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
	miniappconfigsvc "lina-plugin-sicau-niu/backend/internal/service/miniappconfig"
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
	// CampusId is the nearest supported campus.
	CampusId string
	// Skin is the deterministic mini-program cattle skin.
	Skin string
	// Lat and Lng are returned only after the cattle has been first-activated.
	Lat *float64
	Lng *float64
	// Area is returned only before first activation and never contains the true anchor.
	Area *VisibleNiuArea
	// Status is the shared-pool activation status string.
	Status string
	// ActivatedByMe reports whether the current player has already activated this cattle.
	ActivatedByMe bool
	// ActivatedBy is the first activator nickname; empty before first activation.
	ActivatedBy string
	// FeedCount is the cumulative number of feeding records for this cattle.
	FeedCount int
	// IronBoost reports whether a stored iron-cow coordinate is currently in range.
	IronBoost bool
}

// VisibleNiuArea is the stable fuzzy region for an inactive cattle.
type VisibleNiuArea struct {
	Name    string
	Lat     float64
	Lng     float64
	RadiusM int
}

// VisibleNiu returns the cattle currently visible to playerID.
func (s *serviceImpl) VisibleNiu(ctx context.Context, playerID int64) ([]*VisibleNiuItem, error) {
	now := time.Now()

	rows := make([]*entitymodel.Niu, 0)
	err := dao.Niu.Ctx(ctx).
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
		if niuCurrentlyVisible(row, now) {
			visible = append(visible, row)
		}
	}

	activatedByMe, err := s.batchActivatedByMe(ctx, playerID, visible)
	if err != nil {
		return nil, err
	}
	feedCounts, err := batchFeedCounts(ctx, visible)
	if err != nil {
		return nil, err
	}
	firstActivators, err := batchFirstActivators(ctx, visible)
	if err != nil {
		return nil, err
	}
	irons, err := currentIronPositions(ctx)
	if err != nil {
		return nil, err
	}
	ironThreshold, err := s.ironBonusThreshold(ctx)
	if err != nil {
		return nil, err
	}

	list := make([]*VisibleNiuItem, 0, len(visible))
	for _, row := range visible {
		item := &VisibleNiuItem{
			Id:            row.Id,
			Code:          row.Code,
			NiuType:       row.NiuType,
			Name:          row.Name,
			CampusId:      miniappconfigsvc.CampusID(row.Lat, row.Lng),
			Skin:          niuSkin(row),
			Status:        row.Status,
			ActivatedByMe: activatedByMe[row.Id],
			ActivatedBy:   firstActivators[row.Id],
			FeedCount:     feedCounts[row.Id],
			IronBoost:     ironInRange(row.Lat, row.Lng, irons, ironThreshold),
		}
		if row.Status == cattlesvc.NiuStatusActive.String() {
			lat, lng := row.Lat, row.Lng
			item.Lat, item.Lng = &lat, &lng
		} else {
			area := miniappconfigsvc.FuzzyArea(row.Id, row.Lat, row.Lng)
			item.Area = &VisibleNiuArea{Name: area.Name, Lat: area.Lat, Lng: area.Lng, RadiusM: area.RadiusM}
		}
		list = append(list, item)
	}
	return list, nil
}

type feedCountRow struct {
	NiuId     int64 `orm:"niu_id"`
	FeedCount int   `orm:"feed_count"`
}

func batchFeedCounts(ctx context.Context, visible []*entitymodel.Niu) (map[int64]int, error) {
	counts := make(map[int64]int, len(visible))
	ids := visibleNiuIDs(visible)
	if len(ids) == 0 {
		return counts, nil
	}
	rows := make([]*feedCountRow, 0)
	err := dao.Feeding.Ctx(ctx).
		Fields(dao.Feeding.Columns().NiuId, "COUNT(*) AS feed_count").
		WhereIn(dao.Feeding.Columns().NiuId, ids).
		Group(dao.Feeding.Columns().NiuId).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	for _, row := range rows {
		counts[row.NiuId] = row.FeedCount
	}
	return counts, nil
}

func batchFirstActivators(ctx context.Context, visible []*entitymodel.Niu) (map[int64]string, error) {
	result := make(map[int64]string, len(visible))
	ids := visibleNiuIDs(visible)
	if len(ids) == 0 {
		return result, nil
	}
	activations := make([]*entitymodel.Activation, 0)
	err := dao.Activation.Ctx(ctx).
		Fields(dao.Activation.Columns().NiuId, dao.Activation.Columns().UserId).
		WhereIn(dao.Activation.Columns().NiuId, ids).
		Where(dao.Activation.Columns().IsFirst, firstActivatorFlag).
		Scan(&activations)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	userIDs := make([]int64, 0, len(activations))
	for _, row := range activations {
		userIDs = append(userIDs, row.UserId)
	}
	users := make([]*entitymodel.User, 0)
	if len(userIDs) > 0 {
		if err = dao.User.Ctx(ctx).Fields(dao.User.Columns().Id, dao.User.Columns().Nickname).WhereIn(dao.User.Columns().Id, userIDs).Scan(&users); err != nil {
			return nil, bizerr.WrapCode(err, CodeQueryFailed)
		}
	}
	names := make(map[int64]string, len(users))
	for _, user := range users {
		names[user.Id] = user.Nickname
	}
	for _, row := range activations {
		result[row.NiuId] = names[row.UserId]
	}
	return result, nil
}

func currentIronPositions(ctx context.Context) ([]*entitymodel.Iron, error) {
	rows := make([]*entitymodel.Iron, 0)
	err := dao.Iron.Ctx(ctx).
		Fields(dao.Iron.Columns().LastLat, dao.Iron.Columns().LastLng).
		Where(dao.Iron.Columns().LocatedAt+" IS NOT NULL").
		WhereNot(dao.Iron.Columns().LastLat, 0).
		WhereNot(dao.Iron.Columns().LastLng, 0).
		Limit(visibleNiuCap).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return rows, nil
}

func (s *serviceImpl) ironBonusThreshold(ctx context.Context) (float64, error) {
	if s.rulesSvc == nil {
		return 12, nil
	}
	return s.rulesSvc.IronBonusThresholdMeters(ctx)
}

func visibleNiuIDs(rows []*entitymodel.Niu) []int64 {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.Id)
	}
	return ids
}

func ironInRange(lat, lng float64, irons []*entitymodel.Iron, threshold float64) bool {
	for _, iron := range irons {
		if haversineMeters(lat, lng, iron.LastLat, iron.LastLng) <= threshold {
			return true
		}
	}
	return false
}

func niuSkin(row *entitymodel.Niu) string {
	if row == nil || row.NiuType != cattlesvc.NiuTypeSpecial.String() {
		return "normal"
	}
	//nolint:misspell // CowSkin uses the mini-program's established British spelling.
	skins := []string{"buffalo", "dairy", "plough"}
	return skins[int(math.Abs(float64(row.Id)))%len(skins)]
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

// niuCurrentlyVisible reports whether the cattle has reached its online time and
// is within its optional visible weekday/time window at now.
func niuCurrentlyVisible(row *entitymodel.Niu, now time.Time) bool {
	if row == nil || row.OnlineAt == nil || row.OnlineAt.After(now) {
		return false
	}
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
