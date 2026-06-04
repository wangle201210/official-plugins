// settlement_v1_activity.go implements the dashboard activity metrics handler and
// its service-to-DTO projections.

package settlement

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/settlement/v1"
	settlementsvc "lina-plugin-sicau-niu/backend/internal/service/settlement"
)

// Activity returns the dashboard activity metrics (DAU series and retention).
func (c *ControllerV1) Activity(ctx context.Context, req *v1.ActivityReq) (res *v1.ActivityRes, err error) {
	activity, err := c.settlementSvc.Activity(ctx, req.Days)
	if err != nil {
		return nil, err
	}
	return &v1.ActivityRes{
		Dau:         toDauPoints(activity.Dau),
		RetentionD1: toRetentionStat(activity.RetentionD1),
		RetentionD7: toRetentionStat(activity.RetentionD7),
	}, nil
}

// toDauPoints projects the DAU series to its response DTOs.
func toDauPoints(rows []*settlementsvc.DauPoint) []*v1.DauPoint {
	points := make([]*v1.DauPoint, 0, len(rows))
	for _, row := range rows {
		points = append(points, &v1.DauPoint{Date: row.Date, ActiveUsers: row.ActiveUsers})
	}
	return points
}

// toRetentionStat projects one retention bucket to its response DTO.
func toRetentionStat(stat *settlementsvc.RetentionStat) *v1.RetentionStat {
	if stat == nil {
		return nil
	}
	return &v1.RetentionStat{
		CohortUsers:   stat.CohortUsers,
		ReturnedUsers: stat.ReturnedUsers,
		Rate:          stat.Rate,
	}
}
