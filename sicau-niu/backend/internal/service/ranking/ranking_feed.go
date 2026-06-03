// ranking_feed.go implements the personal feeding leaderboard. The Top-N is
// aggregated on the database side by grouping feeding rows by user, summing the
// effect and ordering descending with a LIMIT; nicknames are batch-assembled in
// one query. The requesting player's own rank is computed without scanning the
// whole board: the player's total is summed once and the number of players
// strictly ahead is counted with a grouped HAVING query.

package ranking

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
)

// FeedBoard is the personal feeding leaderboard result: the Top-N rows and the
// requesting player's own rank.
type FeedBoard struct {
	// List is the Top-N players ordered by total feeding effect descending.
	List []*PlayerRank
	// Self is the requesting player's own rank and total.
	Self *SelfRank
}

// PlayerRank is one ranked player on a personal board.
type PlayerRank struct {
	// Rank is the 1-based position on the board.
	Rank int
	// UserId is the player ID.
	UserId int64
	// Nickname is the player nickname; empty when unset.
	Nickname string
	// Total is the player's total feeding effect.
	Total int64
}

// SelfRank is the requesting player's own rank and total on a board.
type SelfRank struct {
	// Rank is the 1-based rank of the player; 0 when the player is off the board.
	Rank int
	// Total is the player's total feeding effect.
	Total int64
}

// userTotalRow is the temporary projection for one grouped user/total aggregate.
type userTotalRow struct {
	UserId int64 `json:"user_id"`
	Total  int64 `json:"total"`
}

// FeedBoard returns the Top-N personal feeding board and the player's own rank.
func (s *serviceImpl) FeedBoard(ctx context.Context, playerID int64) (*FeedBoard, error) {
	rows, err := s.topUserTotals(ctx, nil)
	if err != nil {
		return nil, err
	}
	list, err := s.assemblePlayerRanks(ctx, rows)
	if err != nil {
		return nil, err
	}
	self, err := s.selfRank(ctx, playerID, nil)
	if err != nil {
		return nil, err
	}
	return &FeedBoard{List: list, Self: self}, nil
}

// topUserTotals aggregates the Top-N user totals on the database side. When
// userIDs is non-nil the aggregation is restricted to those players (used by the
// friend board); a nil filter aggregates all players. The returned rows are
// ordered by total descending and capped at the configured Top-N.
func (s *serviceImpl) topUserTotals(ctx context.Context, userIDs []int64) ([]*userTotalRow, error) {
	model := dao.Feeding.Ctx(ctx)
	if userIDs != nil {
		if len(userIDs) == 0 {
			return []*userTotalRow{}, nil
		}
		model = model.WhereIn(dao.Feeding.Columns().UserId, userIDs)
	}
	rows := make([]*userTotalRow, 0, s.topN)
	err := model.
		Fields(dao.Feeding.Columns().UserId, "SUM("+dao.Feeding.Columns().EffectAmount+") AS total").
		Group(dao.Feeding.Columns().UserId).
		Order("total DESC").
		Limit(s.topN).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRankingQueryFailed)
	}
	return rows, nil
}

// assemblePlayerRanks projects the grouped user totals to ranked player rows,
// batch-loading the nicknames for the listed users in one query to avoid N+1 and
// assigning 1-based ranks in the already-sorted order.
func (s *serviceImpl) assemblePlayerRanks(ctx context.Context, rows []*userTotalRow) ([]*PlayerRank, error) {
	if len(rows) == 0 {
		return []*PlayerRank{}, nil
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.UserId)
	}
	nicknames, err := s.batchNicknames(ctx, ids)
	if err != nil {
		return nil, err
	}
	list := make([]*PlayerRank, 0, len(rows))
	for i, row := range rows {
		list = append(list, &PlayerRank{
			Rank:     i + 1,
			UserId:   row.UserId,
			Nickname: nicknames[row.UserId],
			Total:    row.Total,
		})
	}
	return list, nil
}

// selfRank computes the requesting player's own rank and total. The player's
// total is summed once; the rank is the number of players strictly ahead of the
// player plus one, counted with a grouped HAVING query so the whole board is
// never scanned. When restrictIDs is non-nil only those players (the friend
// cohort) participate in the ahead count and the player must be in the cohort. A
// non-positive playerID or a zero total yields rank 0.
func (s *serviceImpl) selfRank(ctx context.Context, playerID int64, restrictIDs []int64) (*SelfRank, error) {
	if playerID <= 0 {
		return &SelfRank{Rank: 0, Total: 0}, nil
	}
	if restrictIDs != nil && !containsID(restrictIDs, playerID) {
		return &SelfRank{Rank: 0, Total: 0}, nil
	}

	total, err := s.userTotal(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if total <= 0 {
		return &SelfRank{Rank: 0, Total: total}, nil
	}

	ahead, err := s.countAhead(ctx, total, restrictIDs)
	if err != nil {
		return nil, err
	}
	return &SelfRank{Rank: ahead + 1, Total: total}, nil
}

// userTotal returns the player's total feeding effect using a database-side SUM.
func (s *serviceImpl) userTotal(ctx context.Context, playerID int64) (int64, error) {
	value, err := dao.Feeding.Ctx(ctx).
		Where(dao.Feeding.Columns().UserId, playerID).
		Sum(dao.Feeding.Columns().EffectAmount)
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeRankingQueryFailed)
	}
	return int64(value), nil
}

// countAhead returns the number of players whose total feeding effect is strictly
// greater than total. It groups feeding rows by user, keeps groups whose summed
// effect exceeds the reference total with a HAVING clause, and counts the
// resulting groups, all on the database side. When restrictIDs is non-nil only
// those players participate.
func (s *serviceImpl) countAhead(ctx context.Context, total int64, restrictIDs []int64) (int, error) {
	model := dao.Feeding.Ctx(ctx)
	if restrictIDs != nil {
		if len(restrictIDs) == 0 {
			return 0, nil
		}
		model = model.WhereIn(dao.Feeding.Columns().UserId, restrictIDs)
	}
	count, err := model.
		Fields(dao.Feeding.Columns().UserId).
		Group(dao.Feeding.Columns().UserId).
		Having("SUM("+dao.Feeding.Columns().EffectAmount+") > ?", total).
		Count()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeRankingQueryFailed)
	}
	return count, nil
}

// containsID reports whether id is present in ids.
func containsID(ids []int64, id int64) bool {
	for _, candidate := range ids {
		if candidate == id {
			return true
		}
	}
	return false
}
