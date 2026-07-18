// This file implements the tenant media-node topology controller endpoint.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// GetDashboardTenantTopology returns tenant-level and media-node aggregates.
func (c *ControllerV1) GetDashboardTenantTopology(ctx context.Context, req *v1.GetDashboardTenantTopologyReq) (res *v1.GetDashboardTenantTopologyRes, err error) {
	out, err := c.mediaSvc.GetDashboardTenantTopology(ctx, mediasvc.DashboardTenantTopologyInput{
		TenantId: req.TenantId,
	})
	if err != nil {
		return nil, err
	}
	return dashboardTenantTopologyToDTO(out), nil
}
