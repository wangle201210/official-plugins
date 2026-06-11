// player_v1_activate.go implements the player LBS activation handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	activationsvc "lina-plugin-sicau-niu/backend/internal/service/activation"
)

// Activate matches and activates a nearby cattle for the authenticated player,
// returning the matched cattle ID, first-activator flag, arrival order,
// activation time and issued main card.
func (c *ControllerV1) Activate(ctx context.Context, req *v1.ActivateReq) (res *v1.ActivateRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.activationSvc.Activate(ctx, playerID, &activationsvc.ActivateInput{
		Lat:       req.Lat,
		Lng:       req.Lng,
		PhotoPath: req.PhotoPath,
	})
	if err != nil {
		return nil, err
	}
	res = &v1.ActivateRes{
		NiuId:       out.NiuId,
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
