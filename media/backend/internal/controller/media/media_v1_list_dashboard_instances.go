package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListDashboardInstances returns dashboard instances in the frontend data shape.
func (c *ControllerV1) ListDashboardInstances(ctx context.Context, req *v1.ListDashboardInstancesReq) (res *v1.ListDashboardInstancesRes, err error) {
	out, err := c.mediaSvc.ListDashboardInstances(ctx, mediasvc.ListDashboardInstancesInput{
		NodeId:  req.NodeId,
		Status:  req.Status,
		Keyword: req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListDashboardInstancesRes{
		NodeInfo:     dashboardInstanceNodeInfoToDTO(out.NodeInfo),
		InstanceList: dashboardInstanceListToDTO(out.List),
	}, nil
}
