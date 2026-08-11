package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	irontransportsvc "lina-plugin-sicau-niu/backend/internal/service/irontransport"
)

func (c *ControllerV1) HeartbeatTransport(ctx context.Context, req *v1.HeartbeatTransportReq) (res *v1.HeartbeatTransportRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.ironTransportSvc.Heartbeat(ctx, playerID, &irontransportsvc.HeartbeatInput{RequestID: req.RequestId, TeamID: req.TeamId, Lat: req.Lat, Lng: req.Lng})
	if err != nil {
		return nil, err
	}
	return &v1.HeartbeatTransportRes{IronTransportState: toIronTransportState(out)}, nil
}
