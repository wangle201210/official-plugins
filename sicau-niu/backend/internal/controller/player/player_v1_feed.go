// player_v1_feed.go implements the player feed-grass handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	feedingsvc "lina-plugin-sicau-niu/backend/internal/service/feeding"
)

// Feed feeds grass from the authenticated player to an activated cattle and
// returns the cattle info, a random quote and the bonus breakdown.
func (c *ControllerV1) Feed(ctx context.Context, req *v1.FeedReq) (res *v1.FeedRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.feedingSvc.Feed(ctx, playerID, &feedingsvc.FeedInput{
		NiuId:      req.NiuId,
		BaseAmount: req.BaseAmount,
		RequestId:  req.RequestId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.FeedRes{
		NiuId:            out.NiuId,
		NiuCode:          out.NiuCode,
		NiuName:          out.NiuName,
		Quote:            out.Quote,
		BaseAmount:       out.BaseAmount,
		CoefficientBasis: out.CoefficientBasis,
		EffectAmount:     out.EffectAmount,
		IsIronBonus:      out.IsIronBonus,
		Balance:          out.Balance,
	}, nil
}
