// player_v1_friend_ranking.go implements the player SICAU-friend leaderboard
// handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

// FriendRanking returns the SICAU-friend leaderboard with the player's own rank.
func (c *ControllerV1) FriendRanking(ctx context.Context, req *v1.FriendRankingReq) (res *v1.FriendRankingRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	board, err := c.rankingSvc.FriendBoard(ctx, playerID)
	if err != nil {
		return nil, err
	}
	return &v1.FriendRankingRes{
		List: toFeedRankItems(board.List),
		Self: toSelfRank(board.Self),
	}, nil
}
