// settlement_anomaly.go implements the M13 risk anomaly alerts: players whose
// single-day feeding or steal count exceeds the configured threshold. Each signal
// is one database-side grouped HAVING query (feeding grouped by player and
// created-at day, steal grouped by actor and steal day); the two result sets are
// merged and the player nicknames are batch-assembled in one projected query to
// avoid N+1. The view is read-only and bounded; it surfaces signals for manual
// review and never auto-acts.

package settlement

import (
	"context"
	"sort"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// Anomaly behaviour type labels reported by the alert view.
const (
	anomalyTypeFeed  = "feed"
	anomalyTypeSteal = "steal"
)

// AnomalyAlerts is the risk anomaly alert result.
type AnomalyAlerts struct {
	// List is the anomaly alerts ordered by single-day count descending, bounded.
	List []*AnomalyAlert
}

// AnomalyAlert is one single-day over-threshold behaviour record.
type AnomalyAlert struct {
	// UserId is the player ID.
	UserId int64
	// Nickname is the player nickname; empty when unset.
	Nickname string
	// Type is the behaviour type: feed or steal.
	Type string
	// Date is the natural day, yyyy-mm-dd.
	Date string
	// Count is the player's behaviour count that day.
	Count int64
	// Threshold is the configured threshold the count exceeded.
	Threshold int64
}

// anomalyRow is the temporary projection for one grouped player/day/count
// aggregate.
type anomalyRow struct {
	Uid int64  `json:"uid"`
	D   string `json:"d"`
	Cnt int64  `json:"cnt"`
}

// Anomalies returns the bounded anomaly alerts merged from feeding and steal.
func (s *serviceImpl) Anomalies(ctx context.Context) (*AnomalyAlerts, error) {
	feedRows, err := s.feedAnomalyRows(ctx)
	if err != nil {
		return nil, err
	}
	stealRows, err := s.stealAnomalyRows(ctx)
	if err != nil {
		return nil, err
	}

	alerts := make([]*AnomalyAlert, 0, len(feedRows)+len(stealRows))
	for _, row := range feedRows {
		alerts = append(alerts, &AnomalyAlert{
			UserId: row.Uid, Type: anomalyTypeFeed, Date: row.D,
			Count: row.Cnt, Threshold: int64(s.feedDailyThreshold),
		})
	}
	for _, row := range stealRows {
		alerts = append(alerts, &AnomalyAlert{
			UserId: row.Uid, Type: anomalyTypeSteal, Date: row.D,
			Count: row.Cnt, Threshold: int64(s.stealDailyThreshold),
		})
	}

	// Stable order: highest single-day count first, bounded to the configured cap.
	sort.SliceStable(alerts, func(i, j int) bool { return alerts[i].Count > alerts[j].Count })
	if len(alerts) > s.anomalyLimit {
		alerts = alerts[:s.anomalyLimit]
	}

	if err = s.fillAnomalyNicknames(ctx, alerts); err != nil {
		return nil, err
	}
	return &AnomalyAlerts{List: alerts}, nil
}

// feedAnomalyRows returns the player/day pairs whose feeding count on a day exceeds
// the feed threshold, grouped and filtered on the database side and bounded.
func (s *serviceImpl) feedAnomalyRows(ctx context.Context) ([]*anomalyRow, error) {
	rows := make([]*anomalyRow, 0)
	err := dao.Feeding.Ctx(ctx).
		Fields(
			dao.Feeding.Columns().UserId+" AS uid",
			"TO_CHAR("+dao.Feeding.Columns().CreatedAt+", 'YYYY-MM-DD') AS d",
			"COUNT(*) AS cnt",
		).
		Group(dao.Feeding.Columns().UserId, "TO_CHAR("+dao.Feeding.Columns().CreatedAt+", 'YYYY-MM-DD')").
		Having("COUNT(*) > ?", s.feedDailyThreshold).
		Order("cnt DESC").
		Limit(s.anomalyLimit).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	return rows, nil
}

// stealAnomalyRows returns the actor/day pairs whose steal count on a day exceeds
// the steal threshold, grouped on the persisted steal_date and bounded.
func (s *serviceImpl) stealAnomalyRows(ctx context.Context) ([]*anomalyRow, error) {
	rows := make([]*anomalyRow, 0)
	err := dao.Steal.Ctx(ctx).
		Fields(
			dao.Steal.Columns().ActorUserId+" AS uid",
			dao.Steal.Columns().StealDate+" AS d",
			"COUNT(*) AS cnt",
		).
		Group(dao.Steal.Columns().ActorUserId, dao.Steal.Columns().StealDate).
		Having("COUNT(*) > ?", s.stealDailyThreshold).
		Order("cnt DESC").
		Limit(s.anomalyLimit).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	return rows, nil
}

// fillAnomalyNicknames batch-loads the nicknames for the alerted players in one
// projected query and assigns them in memory to avoid a per-alert lookup.
func (s *serviceImpl) fillAnomalyNicknames(ctx context.Context, alerts []*AnomalyAlert) error {
	if len(alerts) == 0 {
		return nil
	}
	idSet := make(map[int64]struct{}, len(alerts))
	ids := make([]int64, 0, len(alerts))
	for _, alert := range alerts {
		if _, seen := idSet[alert.UserId]; !seen {
			idSet[alert.UserId] = struct{}{}
			ids = append(ids, alert.UserId)
		}
	}
	nicknames, err := s.batchPlayerNicknames(ctx, ids)
	if err != nil {
		return err
	}
	for _, alert := range alerts {
		alert.Nickname = nicknames[alert.UserId]
	}
	return nil
}

// batchPlayerNicknames returns a user-ID to nickname map for ids in one projected
// query. An empty id set short-circuits without a query.
func (s *serviceImpl) batchPlayerNicknames(ctx context.Context, ids []int64) (map[int64]string, error) {
	names := make(map[int64]string, len(ids))
	if len(ids) == 0 {
		return names, nil
	}
	rows := make([]*entitymodel.User, 0, len(ids))
	err := dao.User.Ctx(ctx).
		Fields(dao.User.Columns().Id, dao.User.Columns().Nickname).
		WhereIn(dao.User.Columns().Id, ids).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	for _, row := range rows {
		names[row.Id] = row.Nickname
	}
	return names, nil
}
