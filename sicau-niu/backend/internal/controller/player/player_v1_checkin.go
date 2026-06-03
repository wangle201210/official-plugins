// player_v1_checkin.go implements the player daily check-in handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

// Checkin performs the authenticated player's daily check-in and returns the
// granted grass amount and the resulting balance.
func (c *ControllerV1) Checkin(ctx context.Context, req *v1.CheckinReq) (res *v1.CheckinRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.grassSvc.Checkin(ctx, playerID)
	if err != nil {
		return nil, err
	}
	return &v1.CheckinRes{Amount: out.Amount, Balance: out.Balance}, nil
}
