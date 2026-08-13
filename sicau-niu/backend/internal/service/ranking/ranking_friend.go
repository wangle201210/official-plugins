// ranking_friend.go implements the SICAU-friend leaderboard. The SICAU-friend
// player cohort is joined directly into bounded database aggregations so the
// full friend-ID set is never loaded into application memory. The self rank is
// only meaningful when the requesting player belongs to the cohort.

package ranking

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
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
// cohort filter is applied consistently by the Top-N and self-rank queries.
func (s *serviceImpl) FriendBoard(ctx context.Context, playerID int64) (*FriendBoard, error) {
	rows, err := s.topFriendTotals(ctx)
	if err != nil {
		return nil, err
	}
	list, err := s.assemblePlayerRanks(ctx, rows)
	if err != nil {
		return nil, err
	}
	self, err := s.friendSelfRank(ctx, playerID)
	if err != nil {
		return nil, err
	}
	return &FriendBoard{List: list, Self: self}, nil
}

func (s *serviceImpl) topFriendTotals(ctx context.Context) ([]*userTotalRow, error) {
	topN, err := s.currentTopN(ctx)
	if err != nil {
		return nil, err
	}
	rows := make([]*userTotalRow, 0, topN)
	err = friendFeedingModel(ctx).
		Fields(
			"f."+dao.Feeding.Columns().UserId+" AS user_id",
			"SUM(f."+dao.Feeding.Columns().EffectAmount+") AS total",
		).
		Group("f." + dao.Feeding.Columns().UserId).
		Order("total DESC").
		OrderAsc("f." + dao.Feeding.Columns().UserId).
		Limit(topN).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRankingQueryFailed)
	}
	return rows, nil
}

func (s *serviceImpl) friendSelfRank(ctx context.Context, playerID int64) (*SelfRank, error) {
	if playerID <= 0 {
		return &SelfRank{}, nil
	}
	count, err := dao.User.Ctx(ctx).
		Where(dao.User.Columns().Id, playerID).
		Where(dao.User.Columns().IdentityType, friendIdentity).
		Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRankingQueryFailed)
	}
	if count == 0 {
		return &SelfRank{}, nil
	}
	total, err := s.userTotal(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if total <= 0 {
		return &SelfRank{Total: total}, nil
	}
	ahead, err := friendFeedingModel(ctx).
		Fields("f."+dao.Feeding.Columns().UserId).
		Group("f."+dao.Feeding.Columns().UserId).
		Having("SUM(f."+dao.Feeding.Columns().EffectAmount+") > ?", total).
		Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRankingQueryFailed)
	}
	return &SelfRank{Rank: ahead + 1, Total: total}, nil
}

func friendFeedingModel(ctx context.Context) *gdb.Model {
	return dao.Feeding.Ctx(ctx).
		As("f").
		InnerJoin(
			userTable+" AS u",
			"u."+dao.User.Columns().Id+" = f."+dao.Feeding.Columns().UserId+
				" AND u."+dao.User.Columns().DeletedAt+" IS NULL",
		).
		Where("u."+dao.User.Columns().IdentityType, friendIdentity)
}
