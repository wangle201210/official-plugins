// This file implements the HotGo-compatible route-memory endpoints.

package mediaopen

import (
	"context"

	"lina-plugin-media/backend/api/mediaopen/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// SetRouteData stores route memory for one HotGo-compatible device channel.
func (c *ControllerV1) SetRouteData(ctx context.Context, req *v1.SetRouteDataReq) (res *v1.SetRouteDataRes, err error) {
	if err = c.mediaSvc.SetRouteMemory(ctx, mediasvc.RouteMemoryInput{
		RouteMemoryKeyInput: mediasvc.RouteMemoryKeyInput{
			DeviceCode:  req.DeviceCode,
			ChannelCode: req.ChannelCode,
		},
		Data: req.Data,
	}); err != nil {
		return nil, err
	}
	return &v1.SetRouteDataRes{}, nil
}

// GetRouteData reads route memory for one HotGo-compatible device channel.
func (c *ControllerV1) GetRouteData(ctx context.Context, req *v1.GetRouteDataReq) (res *v1.GetRouteDataRes, err error) {
	out, err := c.mediaSvc.GetRouteMemory(ctx, mediasvc.RouteMemoryKeyInput{
		DeviceCode:  req.DeviceCode,
		ChannelCode: req.ChannelCode,
	})
	if err != nil {
		return nil, err
	}
	return &v1.GetRouteDataRes{Data: out.Data}, nil
}

// DelRouteData deletes route memory for one HotGo-compatible device channel.
func (c *ControllerV1) DelRouteData(ctx context.Context, req *v1.DelRouteDataReq) (res *v1.DelRouteDataRes, err error) {
	if err = c.mediaSvc.DeleteRouteMemory(ctx, mediasvc.RouteMemoryKeyInput{
		DeviceCode:  req.DeviceCode,
		ChannelCode: req.ChannelCode,
	}); err != nil {
		return nil, err
	}
	return &v1.DelRouteDataRes{}, nil
}
