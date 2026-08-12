// admin_v1_list_transport_reports.go handles operator report audit queries.
package admin

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/admin/v1"
	irontransportsvc "lina-plugin-sicau-niu/backend/internal/service/irontransport"
)

func (c *ControllerV1) ListTransportReports(ctx context.Context, req *v1.ListTransportReportsReq) (res *v1.ListTransportReportsRes, err error) {
	out, err := c.ironTransportSvc.ListAdminReports(ctx, &irontransportsvc.AdminReportListInput{PageNum: req.PageNum, PageSize: req.PageSize, TeamID: req.TeamId, UserID: req.UserId, ActivityDate: req.ActivityDate})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.AdminTransportReport, 0, len(out.List))
	for _, report := range out.List {
		list = append(list, toAdminTransportReport(report))
	}
	return &v1.ListTransportReportsRes{List: list, Total: out.Total}, nil
}
