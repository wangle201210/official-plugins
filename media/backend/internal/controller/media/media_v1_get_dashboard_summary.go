// This file serves platform-wide active stream and session totals.

package media

import (
	"context"

	"lina-plugin-media/backend/api/media/v1"
)

// GetDashboardSummary returns both exact counts through the existing management gate.
func (c *ControllerV1) GetDashboardSummary(ctx context.Context, _ *v1.GetDashboardSummaryReq) (res *v1.GetDashboardSummaryRes, err error) {
	out, err := c.mediaSvc.GetDashboardSummary(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.GetDashboardSummaryRes{
		ActiveStreamCount:  out.ActiveStreamCount,
		ActiveSessionCount: out.ActiveSessionCount,
	}, nil
}
