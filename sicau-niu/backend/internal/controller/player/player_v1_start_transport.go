package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

func (c *ControllerV1) StartTransport(ctx context.Context, req *v1.StartTransportReq) (res *v1.StartTransportRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.ironTransportSvc.Start(ctx, playerID, req.TeamId, req.RequestId)
	if err != nil {
		return nil, err
	}
	return &v1.StartTransportRes{IronTransportState: toIronTransportState(out)}, nil
}
