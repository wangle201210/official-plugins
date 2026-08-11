package player

import (
	"context"
	"strconv"

	"lina-core/pkg/apitime"
	"lina-plugin-sicau-niu/backend/api/player/v1"
)

func (c *ControllerV1) Activities(ctx context.Context, req *v1.ActivitiesReq) (res *v1.ActivitiesRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	items, err := c.grassSocialSvc.Activities(ctx, playerID)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.ActivityItem, 0, len(items))
	for _, item := range items {
		action, target := "送给了你", "苜蓿草"
		if item.Type == "steal" {
			action = "领走了你的"
		}
		name := item.ActorName
		if name == "" {
			name = "川农同学"
		}
		list = append(list, &v1.ActivityItem{Id: item.Type + ":" + strconv.FormatInt(item.ID, 10), Name: name, Action: action, Amount: strconv.Itoa(item.Amount) + "捆", Target: target, OccurredAt: apitime.Milli(item.OccurredAt), Type: item.Type})
	}
	return &v1.ActivitiesRes{List: list}, nil
}
