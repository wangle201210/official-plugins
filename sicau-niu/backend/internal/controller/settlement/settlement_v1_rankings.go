// settlement_v1_rankings.go implements operator-facing leaderboard view handlers.
// The handlers reuse the ranking service's database-side aggregation and only
// project bounded ranked rows to settlement DTOs.

package settlement

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/settlement/v1"
	rankingsvc "lina-plugin-sicau-niu/backend/internal/service/ranking"
)

// FeedRanking returns the operator personal feeding leaderboard.
func (c *ControllerV1) FeedRanking(ctx context.Context, req *v1.FeedRankingReq) (res *v1.FeedRankingRes, err error) {
	board, err := c.rankingSvc.FeedBoard(ctx, 0)
	if err != nil {
		return nil, err
	}
	return &v1.FeedRankingRes{List: toFeedRankItems(board.List)}, nil
}

// FriendRanking returns the operator SICAU-friend leaderboard.
func (c *ControllerV1) FriendRanking(ctx context.Context, req *v1.FriendRankingReq) (res *v1.FriendRankingRes, err error) {
	board, err := c.rankingSvc.FriendBoard(ctx, 0)
	if err != nil {
		return nil, err
	}
	return &v1.FriendRankingRes{List: toFeedRankItems(board.List)}, nil
}

// CollegeRanking returns the operator college leaderboard.
func (c *ControllerV1) CollegeRanking(ctx context.Context, req *v1.CollegeRankingReq) (res *v1.CollegeRankingRes, err error) {
	board, err := c.rankingSvc.CollegeBoard(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]*v1.CollegeRankItem, 0, len(board.List))
	for _, row := range board.List {
		items = append(items, &v1.CollegeRankItem{
			Rank:        row.Rank,
			CollegeId:   row.CollegeId,
			CollegeName: row.CollegeName,
			Total:       row.Total,
		})
	}
	return &v1.CollegeRankingRes{List: items}, nil
}

// toFeedRankItems projects ranked player rows to settlement response DTOs.
func toFeedRankItems(rows []*rankingsvc.PlayerRank) []*v1.FeedRankItem {
	items := make([]*v1.FeedRankItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &v1.FeedRankItem{
			Rank:     row.Rank,
			UserId:   row.UserId,
			Nickname: row.Nickname,
			Total:    row.Total,
		})
	}
	return items
}
