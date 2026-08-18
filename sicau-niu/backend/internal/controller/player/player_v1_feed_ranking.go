// player_v1_feed_ranking.go implements the player personal feeding leaderboard
// handler and the shared ranking service-to-DTO projection helpers.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	rankingsvc "lina-plugin-sicau-niu/backend/internal/service/ranking"
)

// FeedRanking returns the personal feeding leaderboard with the player's own rank.
func (c *ControllerV1) FeedRanking(ctx context.Context, req *v1.FeedRankingReq) (res *v1.FeedRankingRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	board, err := c.rankingSvc.FeedBoard(ctx, playerID)
	if err != nil {
		return nil, err
	}
	return &v1.FeedRankingRes{
		List:    toFeedRankItems(board.List),
		Self:    toSelfRank(board.Self),
		NiuList: toNiuRankItems(board.NiuList),
	}, nil
}

// toFeedRankItems projects the ranked player rows to their response DTOs.
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

// toNiuRankItems projects cattle leaderboard rows to response DTOs.
func toNiuRankItems(rows []*rankingsvc.NiuRank) []*v1.NiuRankItem {
	items := make([]*v1.NiuRankItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &v1.NiuRankItem{
			Rank: row.Rank, NiuId: row.NiuId, NiuCode: row.NiuCode, NiuName: row.NiuName, Total: row.Total,
		})
	}
	return items
}

// toSelfRank projects the requesting player's own rank to its response DTO.
func toSelfRank(self *rankingsvc.SelfRank) *v1.SelfRank {
	if self == nil {
		return nil
	}
	return &v1.SelfRank{Rank: self.Rank, Total: self.Total}
}
