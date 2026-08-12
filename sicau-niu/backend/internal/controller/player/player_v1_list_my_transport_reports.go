// player_v1_list_my_transport_reports.go handles the caller's report history query.
package player

import (
	"context"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	irontransportsvc "lina-plugin-sicau-niu/backend/internal/service/irontransport"
)

func (c *ControllerV1) ListMyTransportReports(ctx context.Context, req *v1.ListMyTransportReportsReq) (res *v1.ListMyTransportReportsRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	out, err := c.ironTransportSvc.ListMyReports(ctx, playerID, &irontransportsvc.PageInput{PageNum: req.PageNum, PageSize: req.PageSize})
	if err != nil {
		return nil, err
	}
	list := make([]*v1.IronTransportReport, 0, len(out.List))
	for _, report := range out.List {
		list = append(list, toIronTransportReport(report))
	}
	return &v1.ListMyTransportReportsRes{List: list, Total: out.Total}, nil
}
