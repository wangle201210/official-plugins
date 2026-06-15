package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// GetDashboardNodeOverview returns one dashboard node tree object.
func (c *ControllerV1) GetDashboardNodeOverview(ctx context.Context, req *v1.GetDashboardNodeOverviewReq) (res *v1.GetDashboardNodeOverviewRes, err error) {
	out, err := c.mediaSvc.GetDashboardNodeOverview(ctx, mediasvc.DashboardNodeOverviewInput{
		RootNodeId:   req.RootNodeId,
		IncludeEmpty: req.IncludeEmpty,
	})
	if err != nil {
		return nil, err
	}
	return dashboardNodeOverviewToDTO(out.Node), nil
}
