// player_v1_collection.go implements the player personal card collection handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

// Collection returns the authenticated player's personal card collection,
// optionally filtered by card category.
func (c *ControllerV1) Collection(ctx context.Context, req *v1.CollectionReq) (res *v1.CollectionRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.activationSvc.Collection(ctx, playerID, req.Category)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.CollectionCardItem, 0, len(items))
	for _, item := range items {
		list = append(list, &v1.CollectionCardItem{
			NiuId:     item.NiuId,
			NiuCode:   item.NiuCode,
			Category:  item.Category,
			Title:     item.Title,
			Content:   item.Content,
			ImagePath: item.ImagePath,
		})
	}
	return &v1.CollectionRes{List: list}, nil
}
