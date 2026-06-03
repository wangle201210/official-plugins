// grasssocial_type.go defines the inbox message-type enum, the social limit
// defaults and the small normalization helpers shared across the grass-social
// capability. The message type is a Go named type persisted in the inbox_msg
// msg_type column.

package grasssocial

// MsgType is the inbox message-type enum. Its string value is persisted in the
// inbox_msg msg_type column and identifies the social event that produced it.
type MsgType string

const (
	// MsgTypeStolen marks a notification that the player's grass was stolen.
	MsgTypeStolen MsgType = "stolen"
	// MsgTypeGiftReceived marks a notification that the player received gifted grass.
	MsgTypeGiftReceived MsgType = "gift_received"
)

// String returns the persisted string form of the message type.
func (t MsgType) String() string {
	return string(t)
}

// socialDateLayout is the YYYY-MM-DD natural-day key used by the steal and gift
// daily counters and by the deterministic daily stealable-list seed.
const socialDateLayout = "2006-01-02"

// Social limit fallbacks used when configuration is absent or non-positive.
const (
	// defaultStealDailyTargets is the fallback daily stealable list size.
	defaultStealDailyTargets = 12
	// defaultStealDailyLimit is the fallback per-day steal action cap.
	defaultStealDailyLimit = 5
	// defaultStealMinAmount is the fallback per-steal random lower bound.
	defaultStealMinAmount = 5
	// defaultStealMaxAmount is the fallback per-steal random upper bound.
	defaultStealMaxAmount = 20
	// defaultGiftDailyLimit is the fallback per-day gift action cap.
	defaultGiftDailyLimit = 12
	// defaultGiftMinAmount is the fallback per-gift minimum amount.
	defaultGiftMinAmount = 12
)

// normalizePositive returns value when it is positive, otherwise the fallback,
// so a missing or non-positive configured limit degrades to a safe default.
func normalizePositive(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

// normalizeStealRange clamps the configured per-steal amount range so the random
// steal amount is always well-defined: both bounds become at least 1 and the
// lower bound never exceeds the upper bound regardless of configuration order.
func normalizeStealRange(min, max int) (int, int) {
	if min <= 0 {
		min = defaultStealMinAmount
	}
	if max <= 0 {
		max = defaultStealMaxAmount
	}
	if max < min {
		max = min
	}
	return min, max
}
