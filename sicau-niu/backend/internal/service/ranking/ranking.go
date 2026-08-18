// Package ranking implements the sicau-niu C5 leaderboard capability: the
// player-and-cattle feeding board, the college board and the SICAU-friend board.
// All boards are aggregated views derived from the C4 feeding effect (no
// authoritative ranking table). The feeding board ranks both individual players
// and cattle; the college board ranks colleges by enrolled-student contribution;
// the SICAU-friend board ranks eligible players. Each board is computed entirely on the
// database side with a Top-N cap, so the data assembly never loads the full set
// into memory. The personal and friend boards additionally compute the requesting
// player's own rank without scanning the whole board: the player's own total is
// summed, then the number of contributors strictly ahead of the player is counted
// with a grouped HAVING query, giving rank = ahead + 1. The boards are a public
// activity view (nicknames and college aggregates only); the self-rank query is
// constrained to the requesting player. All store access uses the generated
// DAO/DO objects so GoFrame manages soft-delete and timestamp columns.
package ranking

import (
	"context"

	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
	rulessvc "lina-plugin-sicau-niu/backend/internal/service/rules"
)

// Config carries the plain-value runtime configuration for the leaderboard
// capability. It holds only scalar tuning values and never runtime dependencies.
type Config struct {
	// TopN is the maximum number of rows returned by each board. A non-positive
	// value falls back to defaultTopN so the board always has a bounded size.
	TopN int
}

// Service defines the C5 leaderboard contract: the player-and-cattle feeding
// board, the college board and the SICAU-friend board.
type Service interface {
	// FeedBoard returns Top-N player and cattle feeding leaderboards, plus the
	// requesting player's own rank and total. Both aggregations run on the
	// database side with a Top-N cap; player and cattle display fields are
	// batch-assembled without N+1 queries. The self rank is 0 when the player has
	// no feeding record. It returns a query bizerr on store failure.
	FeedBoard(ctx context.Context, playerID int64) (out *FeedBoard, err error)
	// CollegeBoard returns the college leaderboard: the Top-N colleges ranked by
	// the total feeding effect of their enrolled students descending. Enrolled
	// students are players whose identity is student with a positive college ID.
	// Aggregation runs on the database side with a Top-N cap and the college names
	// are batch-assembled in one query to avoid N+1. It returns a query bizerr on
	// store failure.
	CollegeBoard(ctx context.Context) (out *CollegeBoard, err error)
	// FriendBoard returns the SICAU-friend leaderboard: the Top-N SICAU-friend
	// players ranked by their personal total feeding effect descending, plus the
	// requesting player's own rank and total when the player is a SICAU-friend.
	// Aggregation runs on the database side with a Top-N cap and nicknames are
	// batch-assembled in one query to avoid N+1. The self rank is 0 when the player
	// is not a SICAU-friend or has no feeding record. It returns a query bizerr on
	// store failure.
	FriendBoard(ctx context.Context, playerID int64) (out *FriendBoard, err error)
}

// Interface compliance assertion for the default ranking service implementation.
var _ Service = (*serviceImpl)(nil)

// serviceImpl implements Service against the plugin-owned feeding, user and
// college tables. The optional rules service supplies the operator-maintained
// Top-N cap; the constructor scalar remains the fallback.
type serviceImpl struct {
	rulesSvc rulessvc.Service // rulesSvc supplies the current leaderboard Top-N when injected.
	topN     int              // topN bounds the size of every board when rules are absent.
}

// New creates a ranking service with an optional runtime-rule service and
// fallback plain-value board configuration. The component reads the plugin's own
// feeding/user/college tables through the generated DAO; the Top-N cap is always
// bounded by either rules or the fallback config.
func New(rulesSvc rulessvc.Service, config Config) Service {
	topN := config.TopN
	if topN <= 0 {
		topN = defaultTopN
	}
	return &serviceImpl{rulesSvc: rulesSvc, topN: topN}
}

// studentIdentity and friendIdentity are the persisted identity-type strings the
// college and friend boards filter on. They reuse the identity package's stable
// enum constants so the ranking filter and the identity contract never drift.
const (
	studentIdentity = string(identitysvc.IdentityTypeStudent)
	friendIdentity  = string(identitysvc.IdentityTypeFriend)
)
