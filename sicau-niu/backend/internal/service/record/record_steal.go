// record_steal.go implements the read-only paged steal-record query. Both the actor
// and the target nicknames are batch-assembled in one query.

package record

import (
	"context"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// ListStealsInput is the steal-record query: optional actor/target filters and
// pagination.
type ListStealsInput struct {
	ActorUserId  int64
	TargetUserId int64
	PageNum      int
	PageSize     int
}

// ListStealsOutput is the steal-record query result.
type ListStealsOutput struct {
	List  []*StealItem
	Total int
}

// StealItem is one steal record projected for the operator console.
type StealItem struct {
	Id             int64
	ActorUserId    int64
	ActorNickname  string
	TargetUserId   int64
	TargetNickname string
	Amount         int
	StealDate      string
	CreatedAt      *int64
}

// ListSteals returns one DB-side paged steal-record page.
func (s *serviceImpl) ListSteals(ctx context.Context, in *ListStealsInput) (*ListStealsOutput, error) {
	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	model := dao.Steal.Ctx(ctx)
	if in.ActorUserId > 0 {
		model = model.Where(dao.Steal.Columns().ActorUserId, in.ActorUserId)
	}
	if in.TargetUserId > 0 {
		model = model.Where(dao.Steal.Columns().TargetUserId, in.TargetUserId)
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRecordQueryFailed)
	}

	rows := make([]*entitymodel.Steal, 0, pageSize)
	err = model.
		Page(pageNum, pageSize).
		OrderDesc(dao.Steal.Columns().CreatedAt).
		OrderDesc(dao.Steal.Columns().Id).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRecordQueryFailed)
	}

	ids := make([]int64, 0, len(rows)*2)
	for _, row := range rows {
		ids = append(ids, row.ActorUserId, row.TargetUserId)
	}
	nicknames, err := batchNicknames(ctx, dedupeInt64(ids))
	if err != nil {
		return nil, err
	}

	list := make([]*StealItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, &StealItem{
			Id:             row.Id,
			ActorUserId:    row.ActorUserId,
			ActorNickname:  nicknames[row.ActorUserId],
			TargetUserId:   row.TargetUserId,
			TargetNickname: nicknames[row.TargetUserId],
			Amount:         row.Amount,
			StealDate:      row.StealDate,
			CreatedAt:      apitime.Milli(row.CreatedAt),
		})
	}
	return &ListStealsOutput{List: list, Total: total}, nil
}
