// player_v1_activate.go implements the player LBS activation handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	activationsvc "lina-plugin-sicau-niu/backend/internal/service/activation"
)

// Activate activates the target cattle for the authenticated player and returns
// the first-activator flag, arrival order, activation time and issued main card.
func (c *ControllerV1) Activate(ctx context.Context, req *v1.ActivateReq) (res *v1.ActivateRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.activationSvc.Activate(ctx, playerID, &activationsvc.ActivateInput{
		NiuId:     req.NiuId,
		Lat:       req.Lat,
		Lng:       req.Lng,
		PhotoPath: req.PhotoPath,
	})
	if err != nil {
		return nil, err
	}
	res = &v1.ActivateRes{
		IsFirst:     out.IsFirst,
		OrderNo:     out.OrderNo,
		ActivatedAt: out.ActivatedAt,
	}
	if out.Card != nil {
		res.Card = &v1.ActivationCard{
			Category:  out.Card.Category,
			Title:     out.Card.Title,
			Content:   out.Card.Content,
			ImagePath: out.Card.ImagePath,
		}
	}
	return res, nil
}
