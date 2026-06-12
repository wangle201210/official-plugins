// rules_config.go implements bounded reads, transactional writes and narrow
// subset accessors for the sicau-niu runtime rule configuration.

package rules

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// Rules returns the complete normalized rule set.
func (s *serviceImpl) Rules(ctx context.Context) (*RuleSet, error) {
	out := s.defaults
	rows := make([]*entitymodel.RuleConfig, 0, len(ruleDefinitions))
	err := dao.RuleConfig.Ctx(ctx).
		Fields(
			dao.RuleConfig.Columns().ConfigKey,
			dao.RuleConfig.Columns().ConfigValue,
		).
		WhereIn(dao.RuleConfig.Columns().ConfigKey, ruleKeys()).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRuleQueryFailed)
	}
	for _, row := range rows {
		if applyErr := applyRawValue(&out, row.ConfigKey, row.ConfigValue); applyErr != nil {
			return nil, applyErr
		}
	}
	return normalizeRuleSet(&out)
}

// Update validates and persists the complete rule set.
func (s *serviceImpl) Update(ctx context.Context, in *RuleSet) (*RuleSet, error) {
	normalized, err := normalizeRuleSet(in)
	if err != nil {
		return nil, err
	}
	values := stringValuesFromRuleSet(normalized)
	err = dao.RuleConfig.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, definition := range ruleDefinitions {
			if txErr := upsertRuleValue(ctx, definition.key, values[definition.key], definition.remark); txErr != nil {
				return txErr
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.Rules(ctx)
}

// upsertRuleValue writes one fixed rule key. The loop is over a fixed small key
// set, not a dynamic result set, so it remains bounded regardless of player data.
func upsertRuleValue(ctx context.Context, key string, value string, remark string) error {
	count, err := dao.RuleConfig.Ctx(ctx).
		Where(dao.RuleConfig.Columns().ConfigKey, key).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeRuleQueryFailed)
	}
	data := do.RuleConfig{
		ConfigValue: value,
		Remark:      remark,
	}
	if count > 0 {
		if _, err = dao.RuleConfig.Ctx(ctx).
			Where(dao.RuleConfig.Columns().ConfigKey, key).
			Data(data).
			Update(); err != nil {
			return bizerr.WrapCode(err, CodeRuleWriteFailed)
		}
		return nil
	}
	if _, err = dao.RuleConfig.Ctx(ctx).Data(do.RuleConfig{
		ConfigKey:   key,
		ConfigValue: value,
		Remark:      remark,
	}).Insert(); err != nil {
		return bizerr.WrapCode(err, CodeRuleWriteFailed)
	}
	return nil
}

// ActivationLBSThresholdMeters returns the activation LBS threshold in meters.
func (s *serviceImpl) ActivationLBSThresholdMeters(ctx context.Context) (float64, error) {
	rules, err := s.Rules(ctx)
	if err != nil {
		return 0, err
	}
	return float64(rules.ActivationLBSThresholdMeters), nil
}

// ActivationGuards returns the activation anti-cheat guard subset.
func (s *serviceImpl) ActivationGuards(ctx context.Context) (ActivationGuards, error) {
	rules, err := s.Rules(ctx)
	if err != nil {
		return ActivationGuards{}, err
	}
	return ActivationGuards{
		DailyAttemptLimit: rules.ActivationDailyAttemptLimit,
		MaxSpeedMps:       rules.ActivationMaxSpeedMps,
	}, nil
}

// PosterCampusBadge returns the poster/certificate campus badge text.
func (s *serviceImpl) PosterCampusBadge(ctx context.Context) (string, error) {
	rules, err := s.Rules(ctx)
	if err != nil {
		return "", err
	}
	return rules.PosterCampusBadge, nil
}

// CheckinRange returns the normalized daily check-in grant range.
func (s *serviceImpl) CheckinRange(ctx context.Context) (int, int, error) {
	rules, err := s.Rules(ctx)
	if err != nil {
		return 0, 0, err
	}
	return rules.CheckinMinAmount, rules.CheckinMaxAmount, nil
}

// SocialRules returns the steal/gift runtime-rule subset.
func (s *serviceImpl) SocialRules(ctx context.Context) (SocialRules, error) {
	rules, err := s.Rules(ctx)
	if err != nil {
		return SocialRules{}, err
	}
	return SocialRules{
		StealDailyTargets: rules.StealDailyTargets,
		StealDailyLimit:   rules.StealDailyLimit,
		StealMinAmount:    rules.StealMinAmount,
		StealMaxAmount:    rules.StealMaxAmount,
		GiftDailyLimit:    rules.GiftDailyLimit,
		GiftMinAmount:     rules.GiftMinAmount,
	}, nil
}

// IronBonusThresholdMeters returns the feeding iron-cow proximity threshold.
func (s *serviceImpl) IronBonusThresholdMeters(ctx context.Context) (float64, error) {
	rules, err := s.Rules(ctx)
	if err != nil {
		return 0, err
	}
	return float64(rules.IronBonusThresholdMeters), nil
}

// RankingTopN returns the normalized leaderboard Top-N cap.
func (s *serviceImpl) RankingTopN(ctx context.Context) (int, error) {
	rules, err := s.Rules(ctx)
	if err != nil {
		return 0, err
	}
	return rules.RankingTopN, nil
}

// AnomalyRules returns the normalized settlement risk-alert thresholds and cap.
func (s *serviceImpl) AnomalyRules(ctx context.Context) (AnomalyRules, error) {
	rules, err := s.Rules(ctx)
	if err != nil {
		return AnomalyRules{}, err
	}
	return AnomalyRules{
		FeedDailyThreshold:  rules.AnomalyFeedDailyThreshold,
		StealDailyThreshold: rules.AnomalyStealDailyThreshold,
		ListLimit:           rules.AnomalyListLimit,
	}, nil
}

// MiniappURL returns the H5 wall return-to-mini-program URL.
func (s *serviceImpl) MiniappURL(ctx context.Context) (string, error) {
	rules, err := s.Rules(ctx)
	if err != nil {
		return "", err
	}
	return rules.MiniappURL, nil
}

// KnownRemarks returns the fixed rule key descriptions for operator tooling.
func KnownRemarks() map[string]string {
	remarks := make(map[string]string, len(ruleDefinitions))
	for _, definition := range ruleDefinitions {
		remarks[definition.key] = ruleRemark(definition.key)
	}
	return remarks
}
