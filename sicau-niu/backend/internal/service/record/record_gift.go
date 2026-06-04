// record_gift.go implements the read-only paged gift-record query. Both the sender
// and the receiver nicknames are batch-assembled in one query.

package record

import (
	"context"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// ListGiftsInput is the gift-record query: optional sender/receiver filters and
// pagination.
type ListGiftsInput struct {
	FromUserId int64
	ToUserId   int64
	PageNum    int
	PageSize   int
}

// ListGiftsOutput is the gift-record query result.
type ListGiftsOutput struct {
	List  []*GiftItem
	Total int
}

// GiftItem is one gift record projected for the operator console.
type GiftItem struct {
	Id           int64
	FromUserId   int64
	FromNickname string
	ToUserId     int64
	ToNickname   string
	Amount       int
	GiftDate     string
	CreatedAt    *int64
}

// ListGifts returns one DB-side paged gift-record page.
func (s *serviceImpl) ListGifts(ctx context.Context, in *ListGiftsInput) (*ListGiftsOutput, error) {
	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	model := dao.Gift.Ctx(ctx)
	if in.FromUserId > 0 {
		model = model.Where(dao.Gift.Columns().FromUserId, in.FromUserId)
	}
	if in.ToUserId > 0 {
		model = model.Where(dao.Gift.Columns().ToUserId, in.ToUserId)
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRecordQueryFailed)
	}

	rows := make([]*entitymodel.Gift, 0, pageSize)
	err = model.
		Page(pageNum, pageSize).
		OrderDesc(dao.Gift.Columns().CreatedAt).
		OrderDesc(dao.Gift.Columns().Id).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRecordQueryFailed)
	}

	ids := make([]int64, 0, len(rows)*2)
	for _, row := range rows {
		ids = append(ids, row.FromUserId, row.ToUserId)
	}
	nicknames, err := batchNicknames(ctx, dedupeInt64(ids))
	if err != nil {
		return nil, err
	}

	list := make([]*GiftItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, &GiftItem{
			Id:           row.Id,
			FromUserId:   row.FromUserId,
			FromNickname: nicknames[row.FromUserId],
			ToUserId:     row.ToUserId,
			ToNickname:   nicknames[row.ToUserId],
			Amount:       row.Amount,
			GiftDate:     row.GiftDate,
			CreatedAt:    apitime.Milli(row.CreatedAt),
		})
	}
	return &ListGiftsOutput{List: list, Total: total}, nil
}
