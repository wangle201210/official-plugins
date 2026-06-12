// rules_config_test.go covers the database-backed runtime-rule read/write path.
// The integration test is gated on LINA_TEST_PGSQL_LINK and remains self-contained
// by truncating the rule_config table before it runs.

package rules

import (
	"context"
	"testing"
)

// TestUpdatePersistsAndReadsRules verifies the complete rule set is persisted and
// subsequent accessors read the operator-maintained values.
func TestUpdatePersistsAndReadsRules(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLRulesDB(t, ctx)

	svc := New(nil)
	input := defaultRuleSet()
	input.ActivationLBSThresholdMeters = 88
	input.ActivationDailyAttemptLimit = 9
	input.ActivationMaxSpeedMps = 30
	input.PosterCampusBadge = " 川农 120 "
	input.CheckinMinAmount = 31
	input.CheckinMaxAmount = 33
	input.StealDailyTargets = 7
	input.RankingTopN = 17
	input.AnomalyListLimit = 11
	input.MiniappURL = " weixin://dl/business/?t=rules "

	updated, err := svc.Update(ctx, &input)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}
	if updated.PosterCampusBadge != "川农 120" || updated.MiniappURL != "weixin://dl/business/?t=rules" {
		t.Fatalf("expected trimmed text fields, got badge=%q url=%q", updated.PosterCampusBadge, updated.MiniappURL)
	}

	got, err := svc.Rules(ctx)
	if err != nil {
		t.Fatalf("Rules failed: %v", err)
	}
	if got.ActivationLBSThresholdMeters != 88 || got.CheckinMinAmount != 31 || got.CheckinMaxAmount != 33 {
		t.Fatalf("unexpected persisted numeric rules: %+v", got)
	}

	topN, err := svc.RankingTopN(ctx)
	if err != nil {
		t.Fatalf("RankingTopN failed: %v", err)
	}
	if topN != 17 {
		t.Fatalf("expected TopN 17, got %d", topN)
	}

	anomaly, err := svc.AnomalyRules(ctx)
	if err != nil {
		t.Fatalf("AnomalyRules failed: %v", err)
	}
	if anomaly.ListLimit != 11 {
		t.Fatalf("expected anomaly list limit 11, got %+v", anomaly)
	}

	guards, err := svc.ActivationGuards(ctx)
	if err != nil {
		t.Fatalf("ActivationGuards failed: %v", err)
	}
	if guards.DailyAttemptLimit != 9 || guards.MaxSpeedMps != 30 {
		t.Fatalf("expected guards 9/30, got %+v", guards)
	}
}
