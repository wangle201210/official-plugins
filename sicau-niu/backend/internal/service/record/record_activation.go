// record_activation.go implements the read-only paged activation-record query with
// player nicknames and cattle names batch-assembled.

package record

import (
	"context"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// ListActivationsInput is the activation-record query: optional player/cattle
// filters and pagination.
type ListActivationsInput struct {
	UserId   int64
	NiuId    int64
	PageNum  int
	PageSize int
}

// ListActivationsOutput is the activation-record query result.
type ListActivationsOutput struct {
	List  []*ActivationItem
	Total int
}

// ActivationItem is one activation record projected for the operator console.
type ActivationItem struct {
	Id           int64
	UserId       int64
	Nickname     string
	NiuId        int64
	NiuName      string
	NiuCode      string
	ActivityDate string
	IsFirst      int
	OrderNo      int
	ActivatedAt  *int64
	CreatedAt    *int64
}

// ListActivations returns one DB-side paged activation-record page.
func (s *serviceImpl) ListActivations(ctx context.Context, in *ListActivationsInput) (*ListActivationsOutput, error) {
	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	model := dao.Activation.Ctx(ctx)
	if in.UserId > 0 {
		model = model.Where(dao.Activation.Columns().UserId, in.UserId)
	}
	if in.NiuId > 0 {
		model = model.Where(dao.Activation.Columns().NiuId, in.NiuId)
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRecordQueryFailed)
	}

	rows := make([]*entitymodel.Activation, 0, pageSize)
	err = model.
		Page(pageNum, pageSize).
		OrderDesc(dao.Activation.Columns().CreatedAt).
		OrderDesc(dao.Activation.Columns().Id).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRecordQueryFailed)
	}

	userIDs := make([]int64, 0, len(rows))
	niuIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.UserId)
		niuIDs = append(niuIDs, row.NiuId)
	}
	nicknames, err := batchNicknames(ctx, dedupeInt64(userIDs))
	if err != nil {
		return nil, err
	}
	niu, err := batchNiu(ctx, dedupeInt64(niuIDs))
	if err != nil {
		return nil, err
	}

	list := make([]*ActivationItem, 0, len(rows))
	for _, row := range rows {
		item := &ActivationItem{
			Id:           row.Id,
			UserId:       row.UserId,
			Nickname:     nicknames[row.UserId],
			NiuId:        row.NiuId,
			ActivityDate: row.ActivityDate,
			IsFirst:      row.IsFirst,
			OrderNo:      row.OrderNo,
			ActivatedAt:  apitime.Milli(row.ActivatedAt),
			CreatedAt:    apitime.Milli(row.CreatedAt),
		}
		if cattle := niu[row.NiuId]; cattle != nil {
			item.NiuName = cattle.Name
			item.NiuCode = cattle.Code
		}
		list = append(list, item)
	}
	return &ListActivationsOutput{List: list, Total: total}, nil
}
