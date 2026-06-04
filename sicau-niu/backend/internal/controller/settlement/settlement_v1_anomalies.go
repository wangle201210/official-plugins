// settlement_v1_anomalies.go implements the risk anomaly alert handler and its
// service-to-DTO projection.

package settlement

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/settlement/v1"
	settlementsvc "lina-plugin-sicau-niu/backend/internal/service/settlement"
)

// RiskAnomalies returns the risk anomaly alerts.
func (c *ControllerV1) RiskAnomalies(ctx context.Context, req *v1.RiskAnomaliesReq) (res *v1.RiskAnomaliesRes, err error) {
	alerts, err := c.settlementSvc.Anomalies(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.RiskAnomaliesRes{List: toAnomalyAlerts(alerts.List)}, nil
}

// toAnomalyAlerts projects the anomaly alerts to their response DTOs.
func toAnomalyAlerts(rows []*settlementsvc.AnomalyAlert) []*v1.AnomalyAlert {
	alerts := make([]*v1.AnomalyAlert, 0, len(rows))
	for _, row := range rows {
		alerts = append(alerts, &v1.AnomalyAlert{
			UserId:    row.UserId,
			Nickname:  row.Nickname,
			Type:      row.Type,
			Date:      row.Date,
			Count:     row.Count,
			Threshold: row.Threshold,
		})
	}
	return alerts
}
