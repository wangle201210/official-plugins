// ranking_friend.go implements the SICAU-friend leaderboard. The SICAU-friend
// player cohort is resolved once on the database side, then the personal feeding
// aggregation, nickname assembly and self-rank computation are all restricted to
// that cohort and reuse the personal-board helpers. The self rank is only
// meaningful when the requesting player belongs to the cohort.

package ranking

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// FriendBoard is the SICAU-friend leaderboard result: the Top-N rows and the
// requesting player's own rank within the cohort.
type FriendBoard struct {
	// List is the Top-N SICAU-friend players ordered by total feeding effect
	// descending.
	List []*PlayerRank
	// Self is the requesting player's own rank and total; rank 0 when the player
	// is not a SICAU-friend or has no feeding record.
	Self *SelfRank
}

// FriendBoard returns the Top-N SICAU-friend board and the player's own rank. The
// cohort is resolved once and reused for the aggregation, nickname assembly and
// self-rank so the friend filter stays consistent across the response.
func (s *serviceImpl) FriendBoard(ctx context.Context, playerID int64) (*FriendBoard, error) {
	friendIDs, err := s.friendUserIDs(ctx)
	if err != nil {
		return nil, err
	}
	if len(friendIDs) == 0 {
		return &FriendBoard{List: []*PlayerRank{}, Self: &SelfRank{Rank: 0, Total: 0}}, nil
	}

	rows, err := s.topUserTotals(ctx, friendIDs)
	if err != nil {
		return nil, err
	}
	list, err := s.assemblePlayerRanks(ctx, rows)
	if err != nil {
		return nil, err
	}
	self, err := s.selfRank(ctx, playerID, friendIDs)
	if err != nil {
		return nil, err
	}
	return &FriendBoard{List: list, Self: self}, nil
}

// friendUserIDs returns the IDs of all active SICAU-friend players in one
// projected query. The cohort is small and bounded by the friend population, so a
// single set read is the authoritative input for the friend aggregation.
func (s *serviceImpl) friendUserIDs(ctx context.Context) ([]int64, error) {
	rows := make([]*entitymodel.User, 0)
	err := dao.User.Ctx(ctx).
		Fields(dao.User.Columns().Id).
		Where(dao.User.Columns().IdentityType, friendIdentity).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRankingQueryFailed)
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.Id)
	}
	return ids, nil
}
