// player_v1_steal_targets.go implements the player daily stealable-target list
// handler.

package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

// StealTargets returns the authenticated player's daily stealable-target list.
func (c *ControllerV1) StealTargets(ctx context.Context, req *v1.StealTargetsReq) (res *v1.StealTargetsRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	targets, err := c.grassSocialSvc.StealTargets(ctx, playerID)
	if err != nil {
		return nil, err
	}
	list := make([]*v1.StealTargetItem, 0, len(targets))
	for _, target := range targets {
		list = append(list, &v1.StealTargetItem{
			UserId:   target.UserId,
			Nickname: target.Nickname,
		})
	}
	return &v1.StealTargetsRes{List: list}, nil
}
