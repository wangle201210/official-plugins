// Package rules implements the sicau-niu operator-maintained runtime rule
// configuration capability. The rule_config table is the authoritative source
// for interaction limits, ranking caps, risk thresholds, poster badge text and
// the H5 wall mini-program link. The service deliberately keeps no process-local
// cache: each request reads the fixed rule-key set in one bounded query, so
// operator updates are visible to the next request without cache invalidation or
// cluster coordination. Updates validate the whole rule set before writing fixed
// keys inside one transaction.
package rules

import "context"

// RuleSet is the full operator-maintained sicau-niu runtime rule set.
type RuleSet struct {
	// ActivationLBSThresholdMeters is the activation LBS distance threshold in meters.
	ActivationLBSThresholdMeters int
	// PosterCampusBadge is the badge rendered on activation posters and certificates.
	PosterCampusBadge string
	// CheckinMinAmount is the daily check-in grass grant lower bound.
	CheckinMinAmount int
	// CheckinMaxAmount is the daily check-in grass grant upper bound.
	CheckinMaxAmount int
	// StealDailyTargets is the daily stealable-target list size.
	StealDailyTargets int
	// StealDailyLimit is the per-player daily steal action cap.
	StealDailyLimit int
	// StealMinAmount is the per-steal random grass lower bound.
	StealMinAmount int
	// StealMaxAmount is the per-steal random grass upper bound.
	StealMaxAmount int
	// GiftDailyLimit is the per-player daily gift action cap.
	GiftDailyLimit int
	// GiftMinAmount is the per-gift minimum grass amount.
	GiftMinAmount int
	// IronBonusThresholdMeters is the feeding iron-cow proximity threshold in meters.
	IronBonusThresholdMeters int
	// RankingTopN is the maximum number of rows returned by each leaderboard.
	RankingTopN int
	// AnomalyFeedDailyThreshold is the single-day feeding-count alert threshold.
	AnomalyFeedDailyThreshold int
	// AnomalyStealDailyThreshold is the single-day steal-count alert threshold.
	AnomalyStealDailyThreshold int
	// AnomalyListLimit is the maximum number of anomaly alerts returned.
	AnomalyListLimit int
	// MiniappURL is the H5 wall return-to-mini-program URL.
	MiniappURL string
}

// SocialRules is the subset consumed by the steal/gift service.
type SocialRules struct {
	// StealDailyTargets is the daily stealable-target list size.
	StealDailyTargets int
	// StealDailyLimit is the per-player daily steal action cap.
	StealDailyLimit int
	// StealMinAmount is the per-steal random grass lower bound.
	StealMinAmount int
	// StealMaxAmount is the per-steal random grass upper bound.
	StealMaxAmount int
	// GiftDailyLimit is the per-player daily gift action cap.
	GiftDailyLimit int
	// GiftMinAmount is the per-gift minimum grass amount.
	GiftMinAmount int
}

// AnomalyRules is the subset consumed by the settlement risk-alert view.
type AnomalyRules struct {
	// FeedDailyThreshold is the single-day feeding-count alert threshold.
	FeedDailyThreshold int
	// StealDailyThreshold is the single-day steal-count alert threshold.
	StealDailyThreshold int
	// ListLimit is the maximum number of anomaly alerts returned.
	ListLimit int
}

// Service defines the operator runtime-rule configuration contract.
type Service interface {
	// Rules returns the complete normalized rule set. It reads all known rule keys
	// in one bounded query, merges missing rows with the injected defaults and
	// returns CodeRuleQueryFailed on store failure or CodeRuleInvalid if persisted
	// values are invalid.
	Rules(ctx context.Context) (out *RuleSet, err error)
	// Update validates and persists the complete rule set inside one transaction.
	// It writes a fixed small key set, inserting missing rows and updating existing
	// rows with DO objects. It returns the normalized persisted rule set, or
	// CodeRuleInvalid/CodeRuleWriteFailed on failure.
	Update(ctx context.Context, in *RuleSet) (out *RuleSet, err error)
	// ActivationLBSThresholdMeters returns the activation LBS threshold in meters.
	ActivationLBSThresholdMeters(ctx context.Context) (meters float64, err error)
	// PosterCampusBadge returns the poster/certificate campus badge text.
	PosterCampusBadge(ctx context.Context) (badge string, err error)
	// CheckinRange returns the normalized daily check-in grant range.
	CheckinRange(ctx context.Context) (min int, max int, err error)
	// SocialRules returns the steal/gift runtime-rule subset.
	SocialRules(ctx context.Context) (out SocialRules, err error)
	// IronBonusThresholdMeters returns the feeding iron-cow proximity threshold.
	IronBonusThresholdMeters(ctx context.Context) (meters float64, err error)
	// RankingTopN returns the normalized leaderboard Top-N cap.
	RankingTopN(ctx context.Context) (topN int, err error)
	// AnomalyRules returns the normalized settlement risk-alert thresholds and cap.
	AnomalyRules(ctx context.Context) (out AnomalyRules, err error)
	// MiniappURL returns the H5 wall return-to-mini-program URL.
	MiniappURL(ctx context.Context) (url string, err error)
}

// Interface compliance assertion for the default rules service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service against the plugin-owned rule_config table.
type serviceImpl struct {
	defaults RuleSet // defaults provides fallback values when individual DB rows are missing.
}

// New creates a rules service with optional startup defaults. A nil defaults
// pointer uses the built-in campaign defaults. The service owns no cache or
// external dependencies; it reads the generated DAO on demand.
func New(defaults *RuleSet) Service {
	value := defaultRuleSet()
	if defaults != nil {
		if normalized, err := normalizeRuleSet(defaults); err == nil {
			value = *normalized
		}
	}
	return &serviceImpl{defaults: value}
}
