package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
	mediasvc "lina-plugin-media/backend/internal/service/media"
)

// ListDashboardSessions returns dashboard sessions grouped by protocol in the frontend data shape.
func (c *ControllerV1) ListDashboardSessions(ctx context.Context, req *v1.ListDashboardSessionsReq) (res *v1.ListDashboardSessionsRes, err error) {
	out, err := c.mediaSvc.ListDashboardSessions(ctx, mediasvc.ListDashboardSessionsInput{
		StreamId:     req.StreamId,
		TenantId:     req.TenantId,
		DeviceId:     req.DeviceId,
		ProtocolType: req.ProtocolType,
		NodeId:       req.NodeId,
		InstanceId:   req.InstanceId,
		Keyword:      req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ListDashboardSessionsRes{
		StreamInfo:   dashboardSessionStreamInfoToDTO(out.StreamInfo),
		ProtocolList: dashboardSessionProtocolListToDTO(out.Protocols),
	}, nil
}
