// Package grasssocial implements the sicau-niu C4 grass-social capability: the
// daily deterministic stealable-target list, the steal action with a daily count
// limit, the gift action with a daily count limit and a per-gift minimum, and the
// player in-app inbox (paged list and mark-read). Steal and gift both move grass
// between two players inside one transaction through the grass ledger service, so
// the two-sided balance change and ledger entries are atomic; each event also
// writes a notification message to the affected player's inbox in the same
// transaction. The stealable list is recomputed from a per-player, per-day seed
// (no stored list) so a steal can be authorized against it without a candidate
// table. Player-facing reads and the inbox are isolated to the authenticated
// player; a steal may only affect a target authorized by that player's own daily
// list. All store access uses the generated DAO/DO objects so GoFrame manages
// soft-delete and timestamps.
package grasssocial

import (
	"context"

	grasssvc "lina-plugin-sicau-niu/backend/internal/service/grass"
	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

// Config carries the plain-value runtime configuration for the grass-social
// capability. It holds only scalar tuning values and never runtime dependencies.
type Config struct {
	// StealDailyTargets is the number of stealable targets in each player's daily
	// random list.
	StealDailyTargets int
	// StealDailyLimit is the maximum number of steal actions a player may perform
	// per natural day.
	StealDailyLimit int
	// StealMinAmount is the inclusive lower bound of the random per-steal amount.
	StealMinAmount int
	// StealMaxAmount is the inclusive upper bound of the random per-steal amount.
	StealMaxAmount int
	// GiftDailyLimit is the maximum number of gift actions a player may perform
	// per natural day.
	GiftDailyLimit int
	// GiftMinAmount is the minimum grass amount per gift action.
	GiftMinAmount int
}

// Service defines the C4 grass-social contract: daily stealable list, steal,
// gift and the player inbox.
type Service interface {
	// StealTargets returns playerID's deterministic daily stealable list (up to
	// the configured size) of other players, each with nickname. The list is
	// recomputed from a per-player, per-day seed and stable within the day. It
	// returns a query bizerr on store failure.
	StealTargets(ctx context.Context, playerID int64) (out []*StealTarget, err error)
	// Steal moves a small random grass amount from in.TargetUserId to playerID. It
	// rejects a target not in playerID's daily list and a player that reached the
	// daily steal limit. Inside one transaction it debits the target, credits the
	// player and writes a stolen-notification to the target's inbox. The stolen
	// amount never exceeds the target's balance. It returns the stolen amount and
	// the player's new balance, or the relevant bizerr on failure.
	Steal(ctx context.Context, playerID int64, in *StealInput) (out *StealResult, err error)
	// Gift moves in.Amount grass from playerID to in.ToUserId. It rejects an
	// amount below the per-gift minimum, a player that reached the daily gift
	// limit and an insufficient balance. Inside one transaction it debits the
	// giver, credits the recipient and writes a gift-received notification to the
	// recipient's inbox. It returns the giver's new balance, or the relevant
	// bizerr on failure.
	Gift(ctx context.Context, playerID int64, in *GiftInput) (out *GiftResult, err error)
	// Messages returns playerID's own inbox page (newest first) with the total
	// count. It is isolated to the current player and returns a query bizerr on
	// store failure.
	Messages(ctx context.Context, playerID int64, in *MessagesInput) (out *MessagesOutput, err error)
	// MarkRead marks one of playerID's own messages read. It rejects a message
	// that does not belong to the player with CodeMessageNotFound and returns a
	// store bizerr on failure.
	MarkRead(ctx context.Context, playerID int64, messageID int64) (err error)
	// Activities returns recent structured steal and gift events that affected
	// playerID, newest first. Actor profiles are batch-loaded to avoid N+1.
	Activities(ctx context.Context, playerID int64) (out []*Activity, err error)
}

// Interface compliance assertion for the default grass-social implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service against the plugin-owned steal, gift, inbox,
// user and grass tables. Its only runtime service dependency is the grass ledger
// service used to move grass between players inside the social transactions; it
// is injected explicitly so dependency changes surface at compile time. The
// plain-value limits are held as scalar fields.
type serviceImpl struct {
	grassSvc          grasssvc.Service // grassSvc moves grass between players inside steal/gift transactions.
	rulesSvc          rulessvc.Service // rulesSvc supplies operator-maintained steal and gift limits when injected.
	stealDailyTargets int              // stealDailyTargets is the daily stealable list size.
	stealDailyLimit   int              // stealDailyLimit is the per-day steal action cap.
	stealMinAmount    int              // stealMinAmount is the per-steal random lower bound.
	stealMaxAmount    int              // stealMaxAmount is the per-steal random upper bound.
	giftDailyLimit    int              // giftDailyLimit is the per-day gift action cap.
	giftMinAmount     int              // giftMinAmount is the per-gift minimum amount.
}

// New creates a grass-social service with explicit dependencies: the grass
// ledger service used for the two-sided transfers, the optional runtime-rule
// service used for operator-maintained social limits, and the fallback
// plain-value social limits. The fallback steal amount range is normalized so
// min<=max and both are positive, keeping the random steal amount well-defined.
func New(grassSvc grasssvc.Service, rulesSvc rulessvc.Service, config Config) Service {
	stealMin, stealMax := normalizeStealRange(config.StealMinAmount, config.StealMaxAmount)
	return &serviceImpl{
		grassSvc:          grassSvc,
		rulesSvc:          rulesSvc,
		stealDailyTargets: normalizePositive(config.StealDailyTargets, defaultStealDailyTargets),
		stealDailyLimit:   normalizePositive(config.StealDailyLimit, defaultStealDailyLimit),
		stealMinAmount:    stealMin,
		stealMaxAmount:    stealMax,
		giftDailyLimit:    normalizePositive(config.GiftDailyLimit, defaultGiftDailyLimit),
		giftMinAmount:     normalizePositive(config.GiftMinAmount, defaultGiftMinAmount),
	}
}
