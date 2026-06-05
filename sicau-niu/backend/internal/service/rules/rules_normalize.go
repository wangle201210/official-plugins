// rules_normalize.go validates and normalizes complete runtime rule sets before
// reads are returned to callers or writes reach the database.

package rules

import (
	"strconv"
	"strings"

	"lina-core/pkg/bizerr"
)

// normalizeRuleSet returns a sanitized copy of in or CodeRuleInvalid when any
// cross-field rule, positive bound or persisted string length is invalid.
func normalizeRuleSet(in *RuleSet) (*RuleSet, error) {
	if in == nil {
		return nil, bizerr.NewCode(CodeRuleInvalid)
	}
	out := *in
	out.PosterCampusBadge = strings.TrimSpace(out.PosterCampusBadge)
	out.MiniappURL = strings.TrimSpace(out.MiniappURL)
	if len(out.PosterCampusBadge) > maxStringValueLen || len(out.MiniappURL) > maxStringValueLen {
		return nil, bizerr.NewCode(CodeRuleInvalid)
	}
	if out.ActivationLBSThresholdMeters <= 0 ||
		out.CheckinMinAmount <= 0 ||
		out.CheckinMaxAmount < out.CheckinMinAmount ||
		out.StealDailyTargets <= 0 ||
		out.StealDailyLimit <= 0 ||
		out.StealMinAmount <= 0 ||
		out.StealMaxAmount < out.StealMinAmount ||
		out.GiftDailyLimit <= 0 ||
		out.GiftMinAmount <= 0 ||
		out.IronBonusThresholdMeters <= 0 ||
		out.RankingTopN <= 0 ||
		out.RankingTopN > maxRankingTopN ||
		out.AnomalyFeedDailyThreshold <= 0 ||
		out.AnomalyStealDailyThreshold <= 0 ||
		out.AnomalyListLimit <= 0 ||
		out.AnomalyListLimit > maxAnomalyListLimit {
		return nil, bizerr.NewCode(CodeRuleInvalid)
	}
	return &out, nil
}

// applyRawValue parses one DB row value into the rule set.
func applyRawValue(out *RuleSet, key string, value string) error {
	switch key {
	case keyPosterCampusBadge:
		out.PosterCampusBadge = value
		return nil
	case keyMiniappURL:
		out.MiniappURL = value
		return nil
	}
	parsed, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return bizerr.NewCode(CodeRuleInvalid)
	}
	switch key {
	case keyActivationLBSThresholdMeters:
		out.ActivationLBSThresholdMeters = parsed
	case keyCheckinMinAmount:
		out.CheckinMinAmount = parsed
	case keyCheckinMaxAmount:
		out.CheckinMaxAmount = parsed
	case keyStealDailyTargets:
		out.StealDailyTargets = parsed
	case keyStealDailyLimit:
		out.StealDailyLimit = parsed
	case keyStealMinAmount:
		out.StealMinAmount = parsed
	case keyStealMaxAmount:
		out.StealMaxAmount = parsed
	case keyGiftDailyLimit:
		out.GiftDailyLimit = parsed
	case keyGiftMinAmount:
		out.GiftMinAmount = parsed
	case keyIronBonusThresholdMeters:
		out.IronBonusThresholdMeters = parsed
	case keyRankingTopN:
		out.RankingTopN = parsed
	case keyAnomalyFeedDailyThreshold:
		out.AnomalyFeedDailyThreshold = parsed
	case keyAnomalyStealDailyThreshold:
		out.AnomalyStealDailyThreshold = parsed
	case keyAnomalyListLimit:
		out.AnomalyListLimit = parsed
	default:
		return nil
	}
	return nil
}

// stringValuesFromRuleSet converts a normalized rule set to persisted text
// values keyed by the stable config keys.
func stringValuesFromRuleSet(in *RuleSet) map[string]string {
	return map[string]string{
		keyActivationLBSThresholdMeters: strconv.Itoa(in.ActivationLBSThresholdMeters),
		keyPosterCampusBadge:            in.PosterCampusBadge,
		keyCheckinMinAmount:             strconv.Itoa(in.CheckinMinAmount),
		keyCheckinMaxAmount:             strconv.Itoa(in.CheckinMaxAmount),
		keyStealDailyTargets:            strconv.Itoa(in.StealDailyTargets),
		keyStealDailyLimit:              strconv.Itoa(in.StealDailyLimit),
		keyStealMinAmount:               strconv.Itoa(in.StealMinAmount),
		keyStealMaxAmount:               strconv.Itoa(in.StealMaxAmount),
		keyGiftDailyLimit:               strconv.Itoa(in.GiftDailyLimit),
		keyGiftMinAmount:                strconv.Itoa(in.GiftMinAmount),
		keyIronBonusThresholdMeters:     strconv.Itoa(in.IronBonusThresholdMeters),
		keyRankingTopN:                  strconv.Itoa(in.RankingTopN),
		keyAnomalyFeedDailyThreshold:    strconv.Itoa(in.AnomalyFeedDailyThreshold),
		keyAnomalyStealDailyThreshold:   strconv.Itoa(in.AnomalyStealDailyThreshold),
		keyAnomalyListLimit:             strconv.Itoa(in.AnomalyListLimit),
		keyMiniappURL:                   in.MiniappURL,
	}
}
