// settlement_activity.go implements the M5 dashboard activity metrics: the daily
// active users (DAU) series and the next-day / 7-day retention. Both derive from
// the existing C1-C4 behaviour tables (activation, feeding, check-in, steal, gift)
// with no new table and no scheduled job. "Active on a day" is the de-duplicated
// union of those tables' (user, day) pairs; the metrics are aggregated entirely on
// the database side with a bounded date range, so the assembly never loads rows
// into memory and never issues a per-user lookup. The union is a parameterized
// analytics query (a cross-table UNION is not expressible through the generated
// DAO model); the table and column identifiers come from the generated DAO
// constants so the query never drifts from the schema, and only the date bounds
// are bound as parameters.

package settlement

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
)

// Activity bounds for the dashboard activity window.
const (
	// defaultActivityDays is the DAU window used when the request omits days.
	defaultActivityDays = 14
	// maxActivityDays caps the DAU window so the query is always bounded.
	maxActivityDays = 60
)

// retentionBuckets are the retention day offsets reported by the dashboard.
const (
	retentionBucketD1 = 1
	retentionBucketD7 = 7
)

// Activity is the dashboard activity result: the DAU series plus next-day and
// 7-day retention.
type Activity struct {
	// Dau is the per-day active-user series ordered by date ascending, calendar
	// gaps filled with zero.
	Dau []*DauPoint
	// RetentionD1 is the next-day retention over elapsed cohorts.
	RetentionD1 *RetentionStat
	// RetentionD7 is the 7-day retention over elapsed cohorts.
	RetentionD7 *RetentionStat
}

// DauPoint is one day's de-duplicated active-user count.
type DauPoint struct {
	Date        string
	ActiveUsers int64
}

// RetentionStat is one retention bucket: the elapsed cohort size, the returned
// count and the rate in [0,1].
type RetentionStat struct {
	CohortUsers   int64
	ReturnedUsers int64
	Rate          float64
}

// behaviourSource names one behaviour table and the column holding its acting
// player, so the active-union branches are built from DAO constants without drift.
type behaviourSource struct {
	table   string
	userCol string
}

// behaviourSources returns the behaviour tables that define player activity: an
// activation, feeding, check-in, steal (actor) or gift (sender) on a day marks the
// player active that day. The identifiers come from the generated DAO so the union
// never drifts from the schema.
func behaviourSources() []behaviourSource {
	return []behaviourSource{
		{dao.Activation.Table(), dao.Activation.Columns().UserId},
		{dao.Feeding.Table(), dao.Feeding.Columns().UserId},
		{dao.Checkin.Table(), dao.Checkin.Columns().UserId},
		{dao.Steal.Table(), dao.Steal.Columns().ActorUserId},
		{dao.Gift.Table(), dao.Gift.Columns().FromUserId},
	}
}

// activeUnionSQL builds the de-duplicated active (user_id, day) relation as a
// UNION across the behaviour tables. Each branch filters soft-deleted rows and,
// when withSince is true, restricts to created_at >= ? (one bound per branch). The
// created_at and deleted_at columns are shared by every behaviour table.
func activeUnionSQL(withSince bool) string {
	createdCol := dao.Feeding.Columns().CreatedAt
	deletedCol := dao.Feeding.Columns().DeletedAt
	branches := make([]string, 0, 5)
	for _, src := range behaviourSources() {
		branch := fmt.Sprintf(
			"SELECT %s AS user_id, CAST(%s AS DATE) AS day FROM %s WHERE %s IS NULL",
			src.userCol, createdCol, src.table, deletedCol,
		)
		if withSince {
			branch += fmt.Sprintf(" AND %s >= ?", createdCol)
		}
		branches = append(branches, branch)
	}
	return strings.Join(branches, "\nUNION\n")
}

// Activity returns the DAU series and the next-day / 7-day retention.
func (s *serviceImpl) Activity(ctx context.Context, days int) (*Activity, error) {
	if days <= 0 {
		days = defaultActivityDays
	}
	if days > maxActivityDays {
		days = maxActivityDays
	}

	dau, err := s.dauSeries(ctx, days)
	if err != nil {
		return nil, err
	}
	retD1, err := s.retention(ctx, retentionBucketD1)
	if err != nil {
		return nil, err
	}
	retD7, err := s.retention(ctx, retentionBucketD7)
	if err != nil {
		return nil, err
	}
	return &Activity{Dau: dau, RetentionD1: retD1, RetentionD7: retD7}, nil
}

// dauSeries returns the per-day active-user counts for the last days days, filling
// calendar gaps with zero. The de-duplicated active relation is grouped by day and
// counted on the database side; the lower bound is the start of the first day in
// the window, bound once per union branch.
func (s *serviceImpl) dauSeries(ctx context.Context, days int) ([]*DauPoint, error) {
	startDay := truncateToDay(time.Now()).AddDate(0, 0, -(days - 1))
	query := fmt.Sprintf(
		"SELECT day, COUNT(*) AS cnt FROM (\n%s\n) t GROUP BY day",
		activeUnionSQL(true),
	)
	args := make([]interface{}, 0, 5)
	for range behaviourSources() {
		args = append(args, startDay)
	}
	records, err := g.DB().GetAll(ctx, query, args...)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	counts := make(map[string]int64, len(records))
	for _, record := range records {
		counts[record["day"].GTime().Format("Y-m-d")] = record["cnt"].Int64()
	}

	series := make([]*DauPoint, 0, days)
	for offset := 0; offset < days; offset++ {
		day := startDay.AddDate(0, 0, offset).Format("2006-01-02")
		series = append(series, &DauPoint{Date: day, ActiveUsers: counts[day]})
	}
	return series, nil
}

// retention returns the retention for the given day offset: of the players whose
// registration day plus the offset has already elapsed, how many were active on
// registration day + offset. The cohort and the return flag are joined on the
// database side; the whole computation is one grouped query.
func (s *serviceImpl) retention(ctx context.Context, offsetDays int) (*RetentionStat, error) {
	query := fmt.Sprintf(
		`WITH active AS (
%s
), cohort AS (
SELECT %s AS user_id, CAST(%s AS DATE) AS reg FROM %s WHERE %s IS NULL
)
SELECT COUNT(*) AS cohort_users, COUNT(a.user_id) AS returned_users
FROM cohort c
LEFT JOIN active a ON a.user_id = c.user_id AND a.day = c.reg + ?::integer
WHERE c.reg + ?::integer <= CURRENT_DATE`,
		activeUnionSQL(false),
		dao.User.Columns().Id, dao.User.Columns().CreatedAt, dao.User.Table(), dao.User.Columns().DeletedAt,
	)
	record, err := g.DB().GetOne(ctx, query, offsetDays, offsetDays)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	cohort := record["cohort_users"].Int64()
	returned := record["returned_users"].Int64()
	rate := 0.0
	if cohort > 0 {
		rate = float64(returned) / float64(cohort)
	}
	return &RetentionStat{CohortUsers: cohort, ReturnedUsers: returned, Rate: rate}, nil
}

// truncateToDay returns the local midnight at the start of t's day.
func truncateToDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}
