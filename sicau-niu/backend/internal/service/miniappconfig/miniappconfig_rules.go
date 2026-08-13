// miniappconfig_rules.go projects the non-sensitive subset of normalized
// gameplay rules returned to the mini program. The caller performs one bounded
// RuleSet read; this file deliberately contains no additional store access.

package miniappconfig

import rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"

// applyRuntimeRules overlays fields owned by the runtime rule service.
func applyRuntimeRules(out *Snapshot, rules *rulessvc.RuleSet) {
	if out == nil || rules == nil {
		return
	}
	out.ActivateRadiusM = rules.ActivationLBSThresholdMeters
	out.StealDailyLimit = rules.StealDailyLimit
	out.GiftDailyLimit = rules.GiftDailyLimit
	out.GiftMinAmount = rules.GiftMinAmount
}
