// wall_v1_first_activators.go implements the public first-activator wall handler
// and its service-to-DTO projection.

package wall

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/wall/v1"
	wallsvc "lina-plugin-sicau-niu/backend/internal/service/wall"
)

// FirstActivators returns the public first-activator wall.
func (c *ControllerV1) FirstActivators(ctx context.Context, req *v1.FirstActivatorsReq) (res *v1.FirstActivatorsRes, err error) {
	board, err := c.wallSvc.FirstActivators(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.FirstActivatorsRes{List: toFirstActivatorItems(board.List)}, nil
}

// toFirstActivatorItems projects the first-activator rows to their response DTOs.
func toFirstActivatorItems(rows []*wallsvc.FirstActivator) []*v1.FirstActivatorItem {
	items := make([]*v1.FirstActivatorItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &v1.FirstActivatorItem{
			Seq:          row.Seq,
			UserId:       row.UserId,
			Nickname:     row.Nickname,
			IdentityType: row.IdentityType,
			NiuId:        row.NiuId,
			NiuName:      row.NiuName,
			NiuCode:      row.NiuCode,
			ActivatedAt:  row.ActivatedAt,
		})
	}
	return items
}
