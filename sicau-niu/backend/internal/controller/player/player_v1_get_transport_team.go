// player_v1_get_transport_team.go handles one effective team query.
package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

func (c *ControllerV1) GetTransportTeam(ctx context.Context, req *v1.GetTransportTeamReq) (res *v1.GetTransportTeamRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	team, err := c.ironTransportSvc.GetTeam(ctx, playerID, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.GetTransportTeamRes{Team: toIronTransportTeam(team)}, nil
}
