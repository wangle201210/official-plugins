// admin_v1_get_transport_stats.go handles operator cloud-moving aggregate queries.
package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
)

func (c *ControllerV1) GetTransportStats(ctx context.Context, req *v1.GetTransportStatsReq) (res *v1.GetTransportStatsRes, err error) {
	out, err := c.ironTransportSvc.AdminStats(ctx, req.ActivityDate)
	if err != nil {
		return nil, err
	}
	return &v1.GetTransportStatsRes{
		EffectiveTeamCount: out.EffectiveTeamCount, InvalidTeamCount: out.InvalidTeamCount,
		ActiveMemberCount: out.ActiveMemberCount, ReportCount: out.ReportCount,
		ContributionMeters: out.ContributionMeters,
	}, nil
}
