package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

func (c *ControllerV1) EndTransport(ctx context.Context, req *v1.EndTransportReq) (res *v1.EndTransportRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.ironTransportSvc.End(ctx, playerID, req.TeamId, req.RequestId)
	if err != nil {
		return nil, err
	}
	return &v1.EndTransportRes{IronTransportState: toIronTransportState(out)}, nil
}
