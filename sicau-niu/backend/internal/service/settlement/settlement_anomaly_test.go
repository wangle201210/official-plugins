// settlement_anomaly_test.go holds the database-gated integration tests for the
// risk anomaly alerts: single-day feeding and steal counts over their configured
// thresholds are surfaced, threshold-equal counts are not, and nicknames are
// assembled. The tests are skipped unless LINA_TEST_PGSQL_LINK is set.

package settlement

import (
	"context"
	"testing"
	"time"
)

// TestAnomaliesFeedAndStealOverThreshold verifies feed/steal single-day counts over
// the configured thresholds are alerted, threshold-equal counts are not, and the
// nicknames are assembled.
func TestAnomaliesFeedAndStealOverThreshold(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSettlementDB(t, ctx)
	// Small thresholds keep the fixtures tiny: feed alerts above 2/day, steal above
	// 1/day.
	svc := New(nil, Config{FeedDailyThreshold: 2, StealDailyThreshold: 1})

	heavy := insertUserRow(t, ctx, "重度喂草", "student", "")
	normal := insertUserRow(t, ctx, "正常喂草", "student", "")
	thief := insertUserRow(t, ctx, "频繁偷草", "student", "")

	// heavy feeds 3 times today (> 2 -> alert); normal feeds 2 times today (= 2, no
	// alert). Today's rows keep the auto-written created_at.
	insertFeedingRow(t, ctx, heavy, 10)
	insertFeedingRow(t, ctx, heavy, 10)
	insertFeedingRow(t, ctx, heavy, 10)
	insertFeedingRow(t, ctx, normal, 10)
	insertFeedingRow(t, ctx, normal, 10)

	// thief steals twice on the same day (> 1 -> alert).
	insertStealRow(t, ctx, thief, normal, 5)
	insertStealRow(t, ctx, thief, heavy, 5)

	alerts, err := svc.Anomalies(ctx)
	if err != nil {
		t.Fatalf("Anomalies returned error: %v", err)
	}
	if len(alerts.List) != 2 {
		t.Fatalf("expected 2 anomaly alerts, got %d", len(alerts.List))
	}

	var feedAlert, stealAlert *AnomalyAlert
	for _, alert := range alerts.List {
		switch alert.Type {
		case anomalyTypeFeed:
			feedAlert = alert
		case anomalyTypeSteal:
			stealAlert = alert
		}
	}
	if feedAlert == nil || feedAlert.UserId != heavy {
		t.Fatalf("missing feed alert for heavy feeder: %+v", feedAlert)
	}
	today := time.Now().Format("2006-01-02")
	if feedAlert.Count != 3 || feedAlert.Threshold != 2 || feedAlert.Date != today {
		t.Fatalf("feed alert = count %d threshold %d date %q, want 3/2/%s",
			feedAlert.Count, feedAlert.Threshold, feedAlert.Date, today)
	}
	if feedAlert.Nickname != "重度喂草" {
		t.Fatalf("feed alert nickname not assembled, got %q", feedAlert.Nickname)
	}
	if stealAlert == nil || stealAlert.UserId != thief || stealAlert.Count != 2 || stealAlert.Threshold != 1 {
		t.Fatalf("steal alert = %+v, want thief count 2 threshold 1", stealAlert)
	}
}

// TestAnomaliesEmptyWhenUnderThreshold verifies no alerts when every player stays
// within both thresholds.
func TestAnomaliesEmptyWhenUnderThreshold(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSettlementDB(t, ctx)
	svc := New(nil, Config{FeedDailyThreshold: 5, StealDailyThreshold: 5})

	user := insertUserRow(t, ctx, "守规矩", "student", "")
	insertFeedingRow(t, ctx, user, 10)
	insertFeedingRow(t, ctx, user, 10)
	insertStealRow(t, ctx, user, user, 5)

	alerts, err := svc.Anomalies(ctx)
	if err != nil {
		t.Fatalf("Anomalies returned error: %v", err)
	}
	if len(alerts.List) != 0 {
		t.Fatalf("expected no anomaly alerts, got %d", len(alerts.List))
	}
}
