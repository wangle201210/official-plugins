// player_v1_college_ranking.go implements the player college leaderboard handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

// CollegeRanking returns the college leaderboard.
func (c *ControllerV1) CollegeRanking(ctx context.Context, req *v1.CollegeRankingReq) (res *v1.CollegeRankingRes, err error) {
	if _, err = currentPlayerID(ctx); err != nil {
		return nil, err
	}
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
