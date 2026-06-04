// settlement_v1_archive.go implements the settlement archive create and list
// handlers and the archive service-to-DTO projection.

package settlement

import (
	"context"

	v1 "lina-plugin-sicau-niu/backend/api/settlement/v1"
	settlementsvc "lina-plugin-sicau-niu/backend/internal/service/settlement"
)

// CreateArchive freezes the current dashboard into a persisted settlement snapshot.
func (c *ControllerV1) CreateArchive(ctx context.Context, req *v1.CreateArchiveReq) (res *v1.CreateArchiveRes, err error) {
	id, err := c.settlementSvc.CreateArchive(ctx, req.Title)
	if err != nil {
		return nil, err
	}
	return &v1.CreateArchiveRes{Id: id}, nil
}

// ListArchives returns the bounded settlement archive list.
func (c *ControllerV1) ListArchives(ctx context.Context, req *v1.ListArchivesReq) (res *v1.ListArchivesRes, err error) {
	archives, err := c.settlementSvc.ListArchives(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.ListArchivesRes{List: toArchiveItems(archives.List)}, nil
}

// toArchiveItems projects the archive records to their response DTOs.
func toArchiveItems(rows []*settlementsvc.Archive) []*v1.ArchiveItem {
	items := make([]*v1.ArchiveItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &v1.ArchiveItem{
			Id:         row.Id,
			Title:      row.Title,
			Snapshot:   row.Snapshot,
			OperatorId: row.OperatorId,
			ArchivedAt: row.ArchivedAt,
		})
	}
	return items
}
