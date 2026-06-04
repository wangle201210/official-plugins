// settlement_v1_export.go implements the player roster export handler and its
// service-to-DTO projection.

package settlement

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/settlement/v1"
	settlementsvc "lina-plugin-sicau-niu/backend/internal/service/settlement"
)

// ExportPlayers returns the bounded player roster for export.
func (c *ControllerV1) ExportPlayers(ctx context.Context, req *v1.ExportPlayersReq) (res *v1.ExportPlayersRes, err error) {
	export, err := c.settlementSvc.ExportPlayers(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ExportPlayersRes{
		List:      toPlayerExportRows(export.List),
		Total:     export.Total,
		Truncated: export.Truncated,
	}, nil
}

// toPlayerExportRows projects the roster rows to their response DTOs.
func toPlayerExportRows(rows []*settlementsvc.PlayerExportRow) []*v1.PlayerExportRow {
	items := make([]*v1.PlayerExportRow, 0, len(rows))
	for _, row := range rows {
		items = append(items, &v1.PlayerExportRow{
			UserId:          row.UserId,
			Nickname:        row.Nickname,
			IdentityType:    row.IdentityType,
			CollegeName:     row.CollegeName,
			Grade:           row.Grade,
			ActivationCount: row.ActivationCount,
			FeedTotalEffect: row.FeedTotalEffect,
		})
	}
	return items
}
