// player_v1_gift.go implements the player gift-grass handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	grasssocialsvc "lina-plugin-sicau-niu/backend/internal/service/grasssocial"
)

// Gift gifts grass from the authenticated player to another player and returns
// the giver's resulting balance.
func (c *ControllerV1) Gift(ctx context.Context, req *v1.GiftReq) (res *v1.GiftRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.grassSocialSvc.Gift(ctx, playerID, &grasssocialsvc.GiftInput{
		ToUserId:  req.ToUserId,
		Amount:    req.Amount,
		RequestId: req.RequestId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.GiftRes{Balance: out.Balance}, nil
}
