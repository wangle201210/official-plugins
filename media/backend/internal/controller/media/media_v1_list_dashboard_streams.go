package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListDashboardStreams returns dashboard streams in the frontend data shape.
func (c *ControllerV1) ListDashboardStreams(ctx context.Context, req *v1.ListDashboardStreamsReq) (res *v1.ListDashboardStreamsRes, err error) {
	out, err := c.mediaSvc.ListDashboardStreams(ctx, mediasvc.ListDashboardStreamsInput{
		SourceType: req.SourceType,
		SourceId:   req.SourceId,
		TenantId:   req.TenantId,
		DeviceId:   req.DeviceId,
		NodeId:     req.NodeId,
		InstanceId: req.InstanceId,
		Status:     req.Status,
		Keyword:    req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListDashboardStreamsRes{
		SourceType: out.SourceType,
		SourceId:   out.SourceId,
		StreamList: dashboardStreamListToDTO(out.List),
	}, nil
}
