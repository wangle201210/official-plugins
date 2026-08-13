// activation_rules_test.go verifies that one activation consumes one complete
// runtime-rule version instead of combining values from concurrent updates.

package activation

import (
	"context"
	"testing"

	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

type sequencedActivationRuleService struct {
	rulessvc.Service
	calls int
	sets  []rulessvc.RuleSet
}

func (s *sequencedActivationRuleService) Rules(context.Context) (*rulessvc.RuleSet, error) {
	value := s.sets[s.calls]
	s.calls++
	return &value, nil
}

func TestActivationRuleSnapshotUsesOneNormalizedVersion(t *testing.T) {
	rules := &sequencedActivationRuleService{sets: []rulessvc.RuleSet{
		{ActivationDailyAttemptLimit: 20, ActivationMaxSpeedMps: 100, ActivationLBSThresholdMeters: 5},
		{ActivationDailyAttemptLimit: 10, ActivationMaxSpeedMps: 1, ActivationLBSThresholdMeters: 100},
	}}
	svc := &serviceImpl{rulesSvc: rules}

	first, err := svc.activationRuleSnapshot(context.Background())
	if err != nil {
		t.Fatalf("first activation rule snapshot failed: %v", err)
	}
	if rules.calls != 1 || first.dailyAttemptLimit != 20 || first.maxSpeedMps != 100 || first.lbsThreshold != 5 {
		t.Fatalf("first snapshot mixed rule versions: calls=%d snapshot=%+v", rules.calls, first)
	}

	second, err := svc.activationRuleSnapshot(context.Background())
	if err != nil {
		t.Fatalf("second activation rule snapshot failed: %v", err)
	}
	if rules.calls != 2 || second.dailyAttemptLimit != 10 || second.maxSpeedMps != 1 || second.lbsThreshold != 100 {
		t.Fatalf("second snapshot mixed rule versions: calls=%d snapshot=%+v", rules.calls, second)
	}
}
