// player_v1_join_transport_team.go handles joining one effective cloud-moving team.
package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

func (c *ControllerV1) JoinTransportTeam(ctx context.Context, req *v1.JoinTransportTeamReq) (res *v1.JoinTransportTeamRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.ironTransportSvc.JoinTeam(ctx, playerID, req.Id, req.RequestId)
	if err != nil {
		return nil, err
	}
	return &v1.JoinTransportTeamRes{IronTransportState: toIronTransportState(out)}, nil
}
