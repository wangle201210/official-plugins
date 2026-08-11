package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

func (c *ControllerV1) LeaveTransportTeam(ctx context.Context, req *v1.LeaveTransportTeamReq) (res *v1.LeaveTransportTeamRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.ironTransportSvc.LeaveTeam(ctx, playerID, req.Id, req.RequestId)
	if err != nil {
		return nil, err
	}
	return &v1.LeaveTransportTeamRes{IronTransportState: toIronTransportState(out)}, nil
}
