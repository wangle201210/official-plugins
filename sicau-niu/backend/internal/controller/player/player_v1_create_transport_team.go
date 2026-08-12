// player_v1_create_transport_team.go handles player cloud-moving team creation.
package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	irontransportsvc "lina-plugin-sicau-niu/backend/internal/service/irontransport"
)

func (c *ControllerV1) CreateTransportTeam(ctx context.Context, req *v1.CreateTransportTeamReq) (res *v1.CreateTransportTeamRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.ironTransportSvc.CreateTeam(ctx, playerID, &irontransportsvc.CreateTeamInput{RequestID: req.RequestId, Name: req.Name})
	if err != nil {
		return nil, err
	}
	return &v1.CreateTransportTeamRes{IronTransportState: toIronTransportState(out)}, nil
}
