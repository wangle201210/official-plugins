// grasssocial_rules.go adapts the optional runtime-rule service into normalized
// social limits. When the rules service is absent, tests and fallback routes use
// the constructor scalar fields.

package grasssocial

import (
	"context"

	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

// socialRules returns the current steal/gift limits.
func (s *serviceImpl) socialRules(ctx context.Context) (rulessvc.SocialRules, error) {
	if s.rulesSvc != nil {
		return s.rulesSvc.SocialRules(ctx)
	}
	return rulessvc.SocialRules{
		StealDailyTargets: s.stealDailyTargets,
		StealDailyLimit:   s.stealDailyLimit,
		StealMinAmount:    s.stealMinAmount,
		StealMaxAmount:    s.stealMaxAmount,
		GiftDailyLimit:    s.giftDailyLimit,
		GiftMinAmount:     s.giftMinAmount,
	}, nil
}
