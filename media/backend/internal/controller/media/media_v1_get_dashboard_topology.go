// This file implements the dashboard topology controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// GetDashboardTopology returns the gateway monitoring topology in the frontend data shape.
func (c *ControllerV1) GetDashboardTopology(ctx context.Context, req *v1.GetDashboardTopologyReq) (res *v1.GetDashboardTopologyRes, err error) {
	out, err := c.mediaSvc.GetDashboardTopology(ctx, mediasvc.DashboardTopologyInput{
		DeviceId: req.DeviceId,
	})
	if err != nil {
		return nil, err
	}
	return dashboardTopologyToDTO(out), nil
}
