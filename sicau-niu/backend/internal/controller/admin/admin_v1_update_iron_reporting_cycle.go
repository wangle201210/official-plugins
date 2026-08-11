// admin_v1_update_iron_reporting_cycle.go handles the operator action that sends
// the fixed 10-second reporting-cycle command to one external IOT locator.

package admin

import (
	"context"
	"time"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

const ironReportingCycle = 10 * time.Second

// UpdateIronReportingCycle sends the fixed activity reporting cycle to the
// locator identified by the current new-iron form without creating a local row.
func (c *ControllerV1) UpdateIronReportingCycle(ctx context.Context, req *v1.UpdateIronReportingCycleReq) (res *v1.UpdateIronReportingCycleRes, err error) {
	if err = c.reportingCycleUpdater.SetReportingCycle(ctx, req.Code, ironReportingCycle); err != nil {
		return nil, err
	}
	return &v1.UpdateIronReportingCycleRes{}, nil
}
