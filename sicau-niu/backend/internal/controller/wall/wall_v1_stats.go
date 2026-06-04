// wall_v1_stats.go implements the public activity stats handler.

package wall

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/wall/v1"
)

// Stats returns the public activity statistics.
func (c *ControllerV1) Stats(ctx context.Context, req *v1.StatsReq) (res *v1.StatsRes, err error) {
	stats, err := c.wallSvc.Stats(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.StatsRes{
		ActivatedNiuCount:   stats.ActivatedNiuCount,
		TotalNiuCount:       stats.TotalNiuCount,
		FirstActivatorCount: stats.FirstActivatorCount,
		PlayerCount:         stats.PlayerCount,
	}, nil
}
