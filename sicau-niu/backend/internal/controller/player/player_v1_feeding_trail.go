// player_v1_feeding_trail.go implements the player recent feeding trail handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

// FeedingTrail returns the authenticated player's most recent feeding records.
func (c *ControllerV1) FeedingTrail(ctx context.Context, req *v1.FeedingTrailReq) (res *v1.FeedingTrailRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.feedingSvc.Trail(ctx, playerID)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.FeedingTrailItem, 0, len(items))
	for _, item := range items {
		list = append(list, &v1.FeedingTrailItem{
			NiuId:        item.NiuId,
			NiuCode:      item.NiuCode,
			NiuName:      item.NiuName,
			EffectAmount: item.EffectAmount,
			IsIronBonus:  item.IsIronBonus,
			FedAt:        item.FedAt,
		})
	}
	return &v1.FeedingTrailRes{List: list}, nil
}
