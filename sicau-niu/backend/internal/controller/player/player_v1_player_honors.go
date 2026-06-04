// player_v1_player_honors.go implements the player honor unlock list handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

// PlayerHonors returns the player's honor list with each honor's unlock status.
func (c *ControllerV1) PlayerHonors(ctx context.Context, req *v1.PlayerHonorsReq) (res *v1.PlayerHonorsRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.honorSvc.PlayerHonors(ctx, playerID)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.PlayerHonorItem, 0, len(items))
	for _, item := range items {
		list = append(list, &v1.PlayerHonorItem{
			Id:         item.Id,
			HonorType:  item.HonorType,
			Code:       item.Code,
			Name:       item.Name,
			UnlockType: item.UnlockType,
			Threshold:  item.Threshold,
			Category:   item.Category,
			ImagePath:  item.ImagePath,
			Unlocked:   item.Unlocked,
		})
	}
	return &v1.PlayerHonorsRes{List: list}, nil
}
