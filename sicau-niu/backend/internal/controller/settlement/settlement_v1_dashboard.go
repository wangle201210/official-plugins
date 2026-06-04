// settlement_v1_dashboard.go implements the operations dashboard handler.

package settlement

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/settlement/v1"
)

// Dashboard returns the operations dashboard.
func (c *ControllerV1) Dashboard(ctx context.Context, req *v1.DashboardReq) (res *v1.DashboardRes, err error) {
	dashboard, err := c.settlementSvc.Dashboard(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.DashboardRes{
		PlayerCount:             dashboard.PlayerCount,
		ActivatedNiuCount:       dashboard.ActivatedNiuCount,
		TotalNiuCount:           dashboard.TotalNiuCount,
		FirstActivatorCount:     dashboard.FirstActivatorCount,
		FeedingCount:            dashboard.FeedingCount,
		FeedTotalEffect:         dashboard.FeedTotalEffect,
		StealCount:              dashboard.StealCount,
		GiftCount:               dashboard.GiftCount,
		CheckinCount:            dashboard.CheckinCount,
		CertificateGrantedCount: dashboard.CertificateGrantedCount,
	}, nil
}
