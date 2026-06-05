// settlement_v1_rules.go implements the operator runtime-rule configuration
// handlers and maps between HTTP DTOs and the rules service contract.

package settlement

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/settlement/v1"
	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

// Rules returns the normalized operator-maintained runtime rules.
func (c *ControllerV1) Rules(ctx context.Context, req *v1.RulesReq) (res *v1.RulesRes, err error) {
	rules, err := c.rulesSvc.Rules(ctx)
	if err != nil {
		return nil, err
	}
	dto := toRuleConfigDTO(rules)
	return (*v1.RulesRes)(dto), nil
}

// UpdateRules replaces the complete operator-maintained runtime rule set.
func (c *ControllerV1) UpdateRules(ctx context.Context, req *v1.UpdateRulesReq) (res *v1.UpdateRulesRes, err error) {
	rules, err := c.rulesSvc.Update(ctx, fromRuleConfigDTO(&req.RuleConfig))
	if err != nil {
		return nil, err
	}
	dto := toRuleConfigDTO(rules)
	return (*v1.UpdateRulesRes)(dto), nil
}

// toRuleConfigDTO projects the rules service model to the public API DTO.
func toRuleConfigDTO(rules *rulessvc.RuleSet) *v1.RuleConfig {
	if rules == nil {
		return &v1.RuleConfig{}
	}
	return &v1.RuleConfig{
		ActivationLBSThresholdMeters: rules.ActivationLBSThresholdMeters,
		PosterCampusBadge:            rules.PosterCampusBadge,
		CheckinMinAmount:             rules.CheckinMinAmount,
		CheckinMaxAmount:             rules.CheckinMaxAmount,
		StealDailyTargets:            rules.StealDailyTargets,
		StealDailyLimit:              rules.StealDailyLimit,
		StealMinAmount:               rules.StealMinAmount,
		StealMaxAmount:               rules.StealMaxAmount,
		GiftDailyLimit:               rules.GiftDailyLimit,
		GiftMinAmount:                rules.GiftMinAmount,
		IronBonusThresholdMeters:     rules.IronBonusThresholdMeters,
		RankingTopN:                  rules.RankingTopN,
		AnomalyFeedDailyThreshold:    rules.AnomalyFeedDailyThreshold,
		AnomalyStealDailyThreshold:   rules.AnomalyStealDailyThreshold,
		AnomalyListLimit:             rules.AnomalyListLimit,
		MiniappURL:                   rules.MiniappURL,
	}
}

// fromRuleConfigDTO maps the public API DTO to the rules service model.
func fromRuleConfigDTO(config *v1.RuleConfig) *rulessvc.RuleSet {
	if config == nil {
		return nil
	}
	return &rulessvc.RuleSet{
		ActivationLBSThresholdMeters: config.ActivationLBSThresholdMeters,
		PosterCampusBadge:            config.PosterCampusBadge,
		CheckinMinAmount:             config.CheckinMinAmount,
		CheckinMaxAmount:             config.CheckinMaxAmount,
		StealDailyTargets:            config.StealDailyTargets,
		StealDailyLimit:              config.StealDailyLimit,
		StealMinAmount:               config.StealMinAmount,
		StealMaxAmount:               config.StealMaxAmount,
		GiftDailyLimit:               config.GiftDailyLimit,
		GiftMinAmount:                config.GiftMinAmount,
		IronBonusThresholdMeters:     config.IronBonusThresholdMeters,
		RankingTopN:                  config.RankingTopN,
		AnomalyFeedDailyThreshold:    config.AnomalyFeedDailyThreshold,
		AnomalyStealDailyThreshold:   config.AnomalyStealDailyThreshold,
		AnomalyListLimit:             config.AnomalyListLimit,
		MiniappURL:                   config.MiniappURL,
	}
}
