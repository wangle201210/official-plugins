// honor_type.go defines the honor stable enums as Go named types with constants
// and validation: the honor type (badge/avatar_frame/certificate) and the unlock
// rule (participation/feed_count/activation_count/category_complete/full_complete).
// These are small, plugin-private, single-language enums governed in code per the
// program design decision to keep them as named constants rather than host
// dictionary entries.

package honor

// HonorType is the honor type enum. Its string value is persisted in the
// honor_def table honor_type column.
type HonorType string

const (
	// HonorTypeBadge marks a badge honor (徽章).
	HonorTypeBadge HonorType = "badge"
	// HonorTypeAvatarFrame marks an avatar-frame honor (头像框).
	HonorTypeAvatarFrame HonorType = "avatar_frame"
	// HonorTypeCertificate marks a certificate honor (证书).
	HonorTypeCertificate HonorType = "certificate"
)

// String returns the persisted string form of the honor type.
func (t HonorType) String() string {
	return string(t)
}

// valid reports whether t is one of the allowed honor types.
func (t HonorType) valid() bool {
	switch t {
	case HonorTypeBadge, HonorTypeAvatarFrame, HonorTypeCertificate:
		return true
	default:
		return false
	}
}

// UnlockType is the honor unlock-rule enum. Its string value is persisted in the
// honor_def table unlock_type column and drives the player honor unlock
// computation.
type UnlockType string

const (
	// UnlockTypeParticipation marks a participation honor unlocked on registration.
	UnlockTypeParticipation UnlockType = "participation"
	// UnlockTypeFeedCount marks an honor unlocked when the player's feeding count
	// reaches the threshold.
	UnlockTypeFeedCount UnlockType = "feed_count"
	// UnlockTypeActivationCount marks an honor unlocked when the player's
	// activation count reaches the threshold.
	UnlockTypeActivationCount UnlockType = "activation_count"
	// UnlockTypeCategoryComplete marks an honor unlocked when the player has
	// collected every active card of the configured category.
	UnlockTypeCategoryComplete UnlockType = "category_complete"
	// UnlockTypeFullComplete marks an honor unlocked when the player has collected
	// every active main card.
	UnlockTypeFullComplete UnlockType = "full_complete"
)

// String returns the persisted string form of the unlock rule.
func (t UnlockType) String() string {
	return string(t)
}

// valid reports whether t is one of the allowed unlock rules.
func (t UnlockType) valid() bool {
	switch t {
	case UnlockTypeParticipation,
		UnlockTypeFeedCount,
		UnlockTypeActivationCount,
		UnlockTypeCategoryComplete,
		UnlockTypeFullComplete:
		return true
	default:
		return false
	}
}

// requiresThreshold reports whether the unlock rule uses the threshold value.
// Only the count-based rules (feed_count/activation_count) need a positive
// threshold; the other rules ignore it.
func (t UnlockType) requiresThreshold() bool {
	return t == UnlockTypeFeedCount || t == UnlockTypeActivationCount
}

// requiresCategory reports whether the unlock rule uses the category value. Only
// the category-complete rule binds a card category; the other rules leave it
// empty.
func (t UnlockType) requiresCategory() bool {
	return t == UnlockTypeCategoryComplete
}
