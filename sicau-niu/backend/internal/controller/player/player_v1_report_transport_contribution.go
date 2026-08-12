// player_v1_report_transport_contribution.go handles immutable location contribution reports.
package player

import (
	"context"
	"time"

	"lina-plugin-sicau-niu/backend/api/player/v1"
	irontransportsvc "lina-plugin-sicau-niu/backend/internal/service/irontransport"
)

func (c *ControllerV1) ReportTransportContribution(ctx context.Context, req *v1.ReportTransportContributionReq) (res *v1.ReportTransportContributionRes, err error) {
	playerID, err := currentPlayerID(ctx)
	if err != nil {
		return nil, err
	}
	if req.Lat == nil || req.Lng == nil {
		return nil, irontransportsvc.InvalidInputError()
	}
	out, err := c.ironTransportSvc.Report(ctx, playerID, &irontransportsvc.ReportInput{
		RequestID: req.RequestId, TeamID: req.TeamId, Lat: *req.Lat, Lng: *req.Lng,
		SampledAt: time.UnixMilli(req.SampledAt),
	})
	if err != nil {
		return nil, err
	}
	return &v1.ReportTransportContributionRes{Report: toIronTransportReportResult(out)}, nil
}
