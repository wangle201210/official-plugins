// This file serves platform-wide alive node, active stream and session totals.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
)

// GetDashboardSummary returns all three counts through the existing management gate.
func (c *ControllerV1) GetDashboardSummary(ctx context.Context, _ *v1.GetDashboardSummaryReq) (res *v1.GetDashboardSummaryRes, err error) {
	out, err := c.mediaSvc.GetDashboardSummary(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetDashboardSummaryRes{
		ActiveStreamCount:  out.ActiveStreamCount,
		ActiveSessionCount: out.ActiveSessionCount,
		AliveNodeCount:     out.AliveNodeCount,
	}, nil
}
