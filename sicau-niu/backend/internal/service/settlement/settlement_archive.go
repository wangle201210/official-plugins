// settlement_archive.go implements the settlement archive: creating a frozen
// dashboard snapshot and listing the archives. The create path computes the current
// dashboard once, serializes it to JSON text and persists one row; the list path
// returns a bounded set ordered by archive time descending. The archive time is a
// business field written explicitly by the service.

package settlement

import (
	"context"
	"encoding/json"
	"time"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// ArchiveList is the settlement archive list result.
type ArchiveList struct {
	// List is the settlement archives ordered by archive time descending, bounded.
	List []*Archive
}

// Archive is one settlement archive record.
type Archive struct {
	// Id is the settlement archive ID.
	Id int64
	// Title is the archive title.
	Title string
	// Snapshot is the frozen dashboard metrics serialized as JSON text.
	Snapshot string
	// OperatorId is the operator user ID who created the archive; 0 when
	// unattributed (the host does not currently expose the current operator ID to
	// plugin handlers).
	OperatorId int64
	// ArchivedAt is the archive time as Unix milliseconds; nil when unset.
	ArchivedAt *int64
}

// CreateArchive freezes the current dashboard into a persisted snapshot and returns
// its ID. The archive time is set explicitly to the current time.
func (s *serviceImpl) CreateArchive(ctx context.Context, title string) (int64, error) {
	dashboard, err := s.Dashboard(ctx)
	if err != nil {
		return 0, err
	}
	raw, err := json.Marshal(dashboard)
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeSettlementArchiveFailed)
	}
	now := time.Now()
	id, err := dao.Settlement.Ctx(ctx).Data(do.Settlement{
		Title:      title,
		Snapshot:   string(raw),
		OperatorId: 0,
		ArchivedAt: &now,
	}).InsertAndGetId()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeSettlementArchiveFailed)
	}
	return id, nil
}

// ListArchives returns the bounded settlement archives ordered by archive time
// descending.
func (s *serviceImpl) ListArchives(ctx context.Context) (*ArchiveList, error) {
	rows := make([]*entitymodel.Settlement, 0, archiveListLimit)
	err := dao.Settlement.Ctx(ctx).
		Order(dao.Settlement.Columns().ArchivedAt + " DESC").
		Limit(archiveListLimit).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	list := make([]*Archive, 0, len(rows))
	for _, row := range rows {
		list = append(list, &Archive{
			Id:         row.Id,
			Title:      row.Title,
			Snapshot:   row.Snapshot,
			OperatorId: row.OperatorId,
			ArchivedAt: apitime.Milli(row.ArchivedAt),
		})
	}
	return &ArchiveList{List: list}, nil
}
