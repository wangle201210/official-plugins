package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
)

func (c *ControllerV1) IronTransportState(ctx context.Context, req *v1.IronTransportStateReq) (res *v1.IronTransportStateRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.ironTransportSvc.State(ctx, playerID)
	if err != nil {
		return nil, err
	}
	return &v1.IronTransportStateRes{IronTransportState: toIronTransportState(out)}, nil
}
