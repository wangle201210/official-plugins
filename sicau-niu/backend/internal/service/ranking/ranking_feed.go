// ranking_feed.go implements the personal feeding leaderboard. The Top-N is
// aggregated on the database side by grouping feeding rows by user, summing the
// effect and ordering descending with a stable user-ID tie-breaker and LIMIT;
// nicknames are batch-assembled in
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
	// NiuList is the Top-N cattle ordered by accumulated feeding effect.
	NiuList []*NiuRank
}

// PlayerRank is one ranked player on a personal board.
type PlayerRank struct {
	// Rank is the 1-based competition rank; tied totals share a rank.
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
	// Rank is the 1-based competition rank; 0 when the player is off the board.
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
	topN, err := s.currentTopN(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := s.topUserTotals(ctx, topN)
	if err != nil {
		return nil, err
	}
	list, err := s.assemblePlayerRanks(ctx, rows)
	if err != nil {
		return nil, err
	}
	self, err := s.selfRank(ctx, playerID)
	if err != nil {
		return nil, err
	}
	niuList, err := s.topNiuRanks(ctx, topN)
	if err != nil {
		return nil, err
	}
	return &FeedBoard{List: list, Self: self, NiuList: niuList}, nil
}

// topUserTotals aggregates all-player Top-N totals on the database side. The
// returned rows are ordered by total descending and capped at the configured
// Top-N.
func (s *serviceImpl) topUserTotals(ctx context.Context, topN int) ([]*userTotalRow, error) {
	rows := make([]*userTotalRow, 0, topN)
	err := dao.Feeding.Ctx(ctx).
		Fields(dao.Feeding.Columns().UserId, "SUM("+dao.Feeding.Columns().EffectAmount+") AS total").
		Group(dao.Feeding.Columns().UserId).
		Order("total DESC").
		OrderAsc(dao.Feeding.Columns().UserId).
		Limit(topN).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRankingQueryFailed)
	}
	return rows, nil
}

// currentTopN returns the operator-maintained leaderboard cap when rules are
// injected, otherwise the constructor fallback.
func (s *serviceImpl) currentTopN(ctx context.Context) (int, error) {
	if s.rulesSvc == nil {
		return s.topN, nil
	}
	return s.rulesSvc.RankingTopN(ctx)
}

// assemblePlayerRanks projects the grouped user totals to ranked player rows,
// batch-loading the nicknames for the listed users in one query to avoid N+1 and
// assigning competition ranks (1, 1, 3) in the already-sorted order so list and
// self projections use the same strictly-ahead semantics.
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
	rank := 0
	var previousTotal int64
	for i, row := range rows {
		if i == 0 || row.Total != previousTotal {
			rank = i + 1
		}
		list = append(list, &PlayerRank{
			Rank:     rank,
			UserId:   row.UserId,
			Nickname: nicknames[row.UserId],
			Total:    row.Total,
		})
		previousTotal = row.Total
	}
	return list, nil
}

// selfRank computes the requesting player's own rank and total. The player's
// total is summed once; the rank is the number of players strictly ahead of the
// player plus one, counted with a grouped HAVING query so the whole board is
// never scanned. A non-positive playerID or a zero total yields rank 0.
func (s *serviceImpl) selfRank(ctx context.Context, playerID int64) (*SelfRank, error) {
	if playerID <= 0 {
		return &SelfRank{Rank: 0, Total: 0}, nil
	}

	total, err := s.userTotal(ctx, playerID)
	if err != nil {
		return nil, err
	}
	if total <= 0 {
		return &SelfRank{Rank: 0, Total: total}, nil
	}

	ahead, err := s.countAhead(ctx, total)
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
// resulting groups, all on the database side.
func (s *serviceImpl) countAhead(ctx context.Context, total int64) (int, error) {
	count, err := dao.Feeding.Ctx(ctx).
		Fields(dao.Feeding.Columns().UserId).
		Group(dao.Feeding.Columns().UserId).
		Having("SUM("+dao.Feeding.Columns().EffectAmount+") > ?", total).
		Count()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeRankingQueryFailed)
	}
	return count, nil
}
