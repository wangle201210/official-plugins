// activation_rules.go resolves the single normalized runtime-rule snapshot used
// by one activation after player serialization and successful-request replay
// detection. It prevents attempt, speed and LBS checks from mixing rule versions
// while an operator update commits concurrently.

package activation

import "context"

type activationRules struct {
	dailyAttemptLimit int
	maxSpeedMps       int
	lbsThreshold      float64
}

func (s *serviceImpl) activationRuleSnapshot(ctx context.Context) (activationRules, error) {
	if s.rulesSvc == nil {
		return activationRules{
			dailyAttemptLimit: defaultDailyAttemptLimit,
			maxSpeedMps:       defaultMaxSpeedMps,
			lbsThreshold:      s.lbsThreshold,
		}, nil
	}
	rules, err := s.rulesSvc.Rules(ctx)
	if err != nil {
		return activationRules{}, err
	}
	return activationRules{
		dailyAttemptLimit: rules.ActivationDailyAttemptLimit,
		maxSpeedMps:       rules.ActivationMaxSpeedMps,
		lbsThreshold:      float64(rules.ActivationLBSThresholdMeters),
	}, nil
}
