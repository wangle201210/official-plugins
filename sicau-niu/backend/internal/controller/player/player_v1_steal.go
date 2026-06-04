// player_v1_steal.go implements the player steal-grass handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	grasssocialsvc "lina-plugin-sicau-niu/backend/internal/service/grasssocial"
)

// Steal steals grass from a target in the player's daily stealable list and
// returns the stolen amount and the resulting balance.
func (c *ControllerV1) Steal(ctx context.Context, req *v1.StealReq) (res *v1.StealRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.grassSocialSvc.Steal(ctx, playerID, &grasssocialsvc.StealInput{
		TargetUserId: req.TargetUserId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.StealRes{Amount: out.Amount, Balance: out.Balance}, nil
}
