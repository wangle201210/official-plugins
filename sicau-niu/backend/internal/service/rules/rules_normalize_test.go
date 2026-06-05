// rules_normalize_test.go covers rule-set validation without requiring a
// database. Database read/write paths are covered by package compilation and the
// plugin-wide PostgreSQL tests when LINA_TEST_PGSQL_LINK is provided.

package rules

import "testing"

// TestNormalizeRuleSetRejectsInvalidRanges verifies cross-field and hard-cap
// validation catches values an operator must not persist.
func TestNormalizeRuleSetRejectsInvalidRanges(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*RuleSet)
	}{
		{name: "checkin max below min", mutate: func(in *RuleSet) { in.CheckinMaxAmount = in.CheckinMinAmount - 1 }},
		{name: "steal max below min", mutate: func(in *RuleSet) { in.StealMaxAmount = in.StealMinAmount - 1 }},
		{name: "ranking over cap", mutate: func(in *RuleSet) { in.RankingTopN = maxRankingTopN + 1 }},
		{name: "anomaly list over cap", mutate: func(in *RuleSet) { in.AnomalyListLimit = maxAnomalyListLimit + 1 }},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			input := defaultRuleSet()
			tt.mutate(&input)
			if _, err := normalizeRuleSet(&input); err == nil {
				t.Fatal("expected invalid rule set to be rejected")
			}
		})
	}
}

// TestNormalizeRuleSetTrimsStrings verifies operator text values are normalized
// before persistence and returned values are predictable.
func TestNormalizeRuleSetTrimsStrings(t *testing.T) {
	input := defaultRuleSet()
	input.PosterCampusBadge = "  120 周年  "
	input.MiniappURL = "  weixin://dl/business/?t=abc  "
	out, err := normalizeRuleSet(&input)
	if err != nil {
		t.Fatalf("normalize failed: %v", err)
	}
	if out.PosterCampusBadge != "120 周年" {
		t.Fatalf("badge not trimmed: %q", out.PosterCampusBadge)
	}
	if out.MiniappURL != "weixin://dl/business/?t=abc" {
		t.Fatalf("miniapp url not trimmed: %q", out.MiniappURL)
	}
}
