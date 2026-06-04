// settlement_activity_test.go holds the database-gated integration tests for the
// dashboard activity metrics: DAU de-duplication across behaviour tables and the
// next-day / 7-day retention cohort semantics (including exclusion of cohorts whose
// retention window has not elapsed). The tests are skipped unless
// LINA_TEST_PGSQL_LINK is set.

package settlement

import (
	"context"
	"testing"
	"time"

	"github.com/gogf/gf/v2/frame/g"
)

// backdateCreatedAt rewrites the created_at of every row of a player in a table to
// the given instant via raw SQL, bypassing GoFrame time maintenance so registration
// and activity days can be controlled in tests.
func backdateCreatedAt(t *testing.T, ctx context.Context, table, userCol string, userID int64, at time.Time) {
	t.Helper()
	_, err := g.DB().Exec(ctx, "UPDATE "+table+" SET created_at = ? WHERE "+userCol+" = ?", at, userID)
	if err != nil {
		t.Fatalf("backdate %s.created_at failed: %v", table, err)
	}
}

// dayAtNoon returns local noon daysAgo days before today, kept away from midnight so
// a small server/local timezone offset cannot flip the calendar day.
func dayAtNoon(daysAgo int) time.Time {
	return truncateToDay(time.Now()).AddDate(0, 0, -daysAgo).Add(12 * time.Hour)
}

// TestActivityDauDeduplicates verifies a player active through multiple behaviours
// on the same day counts once, and distinct players add up.
func TestActivityDauDeduplicates(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSettlementDB(t, ctx)
	svc := New()

	userA := insertUserRow(t, ctx, "A", "student", "")
	userB := insertUserRow(t, ctx, "B", "student", "")
	// A acts twice today (feed + check-in); B acts once today. Today's rows keep the
	// auto-written created_at, so they fall in the default DAU window.
	insertFeedingRow(t, ctx, userA, 10)
	insertCheckinRow(t, ctx, userA, 1)
	insertFeedingRow(t, ctx, userB, 10)

	activity, err := svc.Activity(ctx, 14)
	if err != nil {
		t.Fatalf("Activity returned error: %v", err)
	}
	today := time.Now().Format("2006-01-02")
	var todayPoint *DauPoint
	for _, point := range activity.Dau {
		if point.Date == today {
			todayPoint = point
		}
	}
	if todayPoint == nil {
		t.Fatalf("today %s missing from DAU series", today)
	}
	if todayPoint.ActiveUsers != 2 {
		t.Fatalf("DAU for today = %d, want 2 (A de-duplicated across feed+checkin, plus B)", todayPoint.ActiveUsers)
	}
	if len(activity.Dau) != 14 {
		t.Fatalf("DAU series length = %d, want 14", len(activity.Dau))
	}
}

// TestActivityRetention verifies next-day retention counts only returning cohort
// members and 7-day retention excludes cohorts whose window has not elapsed.
func TestActivityRetention(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSettlementDB(t, ctx)
	svc := New()

	// Two players registered 2 days ago; one returns on registration day + 1, one
	// does not. The next-day window (reg+1 = yesterday) has elapsed for both; the
	// 7-day window (reg+7 = 5 days from now) has not elapsed for either.
	returner := insertUserRow(t, ctx, "Returner", "student", "")
	absent := insertUserRow(t, ctx, "Absent", "student", "")
	backdateCreatedAt(t, ctx, "plugin_sicau_niu_user", "id", returner, dayAtNoon(2))
	backdateCreatedAt(t, ctx, "plugin_sicau_niu_user", "id", absent, dayAtNoon(2))

	// The returner feeds on registration day + 1 (yesterday); backdate that action.
	insertFeedingRow(t, ctx, returner, 10)
	backdateCreatedAt(t, ctx, "plugin_sicau_niu_feeding", "user_id", returner, dayAtNoon(1))

	activity, err := svc.Activity(ctx, 14)
	if err != nil {
		t.Fatalf("Activity returned error: %v", err)
	}

	if activity.RetentionD1.CohortUsers != 2 || activity.RetentionD1.ReturnedUsers != 1 {
		t.Fatalf("D1 retention = cohort %d returned %d, want cohort 2 returned 1",
			activity.RetentionD1.CohortUsers, activity.RetentionD1.ReturnedUsers)
	}
	if activity.RetentionD1.Rate != 0.5 {
		t.Fatalf("D1 retention rate = %v, want 0.5", activity.RetentionD1.Rate)
	}
	// reg+7 has not elapsed, so neither player is in the D7 cohort.
	if activity.RetentionD7.CohortUsers != 0 {
		t.Fatalf("D7 cohort = %d, want 0 (window not elapsed)", activity.RetentionD7.CohortUsers)
	}
}
