// player_v1_visible_niu.go implements the player visible-cattle map list handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	activationsvc "lina-plugin-sicau-niu/backend/internal/service/activation"
)

// VisibleNiu returns the cattle currently visible to the authenticated player
// with shared-pool status and the player's own activation flag.
func (c *ControllerV1) VisibleNiu(ctx context.Context, req *v1.VisibleNiuReq) (res *v1.VisibleNiuRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.activationSvc.VisibleNiu(ctx, playerID)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.VisibleNiuItem, 0, len(items))
	for _, item := range items {
		list = append(list, toVisibleNiuItem(item))
	}
	return &v1.VisibleNiuRes{List: list}, nil
}

func toVisibleNiuItem(item *activationsvc.VisibleNiuItem) *v1.VisibleNiuItem {
	if item == nil {
		return nil
	}
	out := &v1.VisibleNiuItem{Id: item.Id, Code: item.Code, NiuType: item.NiuType, Name: item.Name, CampusId: item.CampusId, Skin: item.Skin, Lat: item.Lat, Lng: item.Lng, Status: item.Status, ActivatedByMe: item.ActivatedByMe, ActivatedBy: item.ActivatedBy, FeedCount: item.FeedCount, IronBoost: item.IronBoost}
	if item.Area != nil {
		out.Area = &v1.NiuArea{Name: item.Area.Name, Lat: item.Area.Lat, Lng: item.Area.Lng, RadiusM: item.Area.RadiusM}
	}
	return out
}
