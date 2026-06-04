// record_grasstxn.go implements the read-only paged grass-ledger query with player
// nicknames batch-assembled. Each row is one signed grass delta with its
// transaction type and the reference ID of the action that produced it.

package record

import (
	"context"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// ListGrassTxnsInput is the grass-ledger query: optional player filter and
// pagination.
type ListGrassTxnsInput struct {
	UserId   int64
	PageNum  int
	PageSize int
}

// ListGrassTxnsOutput is the grass-ledger query result.
type ListGrassTxnsOutput struct {
	List  []*GrassTxnItem
	Total int
}

// GrassTxnItem is one grass-ledger entry projected for the operator console.
type GrassTxnItem struct {
	Id        int64
	UserId    int64
	Nickname  string
	Delta     int64
	TxnType   string
	RefId     int64
	CreatedAt *int64
}

// ListGrassTxns returns one DB-side paged grass-ledger page.
func (s *serviceImpl) ListGrassTxns(ctx context.Context, in *ListGrassTxnsInput) (*ListGrassTxnsOutput, error) {
	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	model := dao.GrassTxn.Ctx(ctx)
	if in.UserId > 0 {
		model = model.Where(dao.GrassTxn.Columns().UserId, in.UserId)
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRecordQueryFailed)
	}

	rows := make([]*entitymodel.GrassTxn, 0, pageSize)
	err = model.
		Page(pageNum, pageSize).
		OrderDesc(dao.GrassTxn.Columns().CreatedAt).
		OrderDesc(dao.GrassTxn.Columns().Id).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRecordQueryFailed)
	}

	userIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.UserId)
	}
	nicknames, err := batchNicknames(ctx, dedupeInt64(userIDs))
	if err != nil {
		return nil, err
	}

	list := make([]*GrassTxnItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, &GrassTxnItem{
			Id:        row.Id,
			UserId:    row.UserId,
			Nickname:  nicknames[row.UserId],
			Delta:     row.Delta,
			TxnType:   row.TxnType,
			RefId:     row.RefId,
			CreatedAt: apitime.Milli(row.CreatedAt),
		})
	}
	return &ListGrassTxnsOutput{List: list, Total: total}, nil
}
