// record_feeding.go implements the read-only paged feeding-record query. Rows are
// filtered, counted and paginated on the database side, then the player nicknames
// and cattle names are batch-assembled in one query each to avoid N+1.

package record

import (
	"context"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// ListFeedingsInput is the feeding-record query: optional cattle/player filters and
// pagination.
type ListFeedingsInput struct {
	NiuId    int64
	UserId   int64
	PageNum  int
	PageSize int
}

// ListFeedingsOutput is the feeding-record query result.
type ListFeedingsOutput struct {
	List  []*FeedingItem
	Total int
}

// FeedingItem is one feeding record projected for the operator console.
type FeedingItem struct {
	Id               int64
	UserId           int64
	Nickname         string
	NiuId            int64
	NiuName          string
	NiuCode          string
	BaseAmount       int
	CoefficientBasis int
	EffectAmount     int
	IsIronBonus      int
	FedAt            *int64
	CreatedAt        *int64
}

// ListFeedings returns one DB-side paged feeding-record page.
func (s *serviceImpl) ListFeedings(ctx context.Context, in *ListFeedingsInput) (*ListFeedingsOutput, error) {
	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	model := dao.Feeding.Ctx(ctx)
	if in.NiuId > 0 {
		model = model.Where(dao.Feeding.Columns().NiuId, in.NiuId)
	}
	if in.UserId > 0 {
		model = model.Where(dao.Feeding.Columns().UserId, in.UserId)
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRecordQueryFailed)
	}

	rows := make([]*entitymodel.Feeding, 0, pageSize)
	err = model.
		Page(pageNum, pageSize).
		OrderDesc(dao.Feeding.Columns().CreatedAt).
		OrderDesc(dao.Feeding.Columns().Id).
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

	list := make([]*FeedingItem, 0, len(rows))
	for _, row := range rows {
		item := &FeedingItem{
			Id:               row.Id,
			UserId:           row.UserId,
			Nickname:         nicknames[row.UserId],
			NiuId:            row.NiuId,
			BaseAmount:       row.BaseAmount,
			CoefficientBasis: row.CoefficientBasis,
			EffectAmount:     row.EffectAmount,
			IsIronBonus:      row.IsIronBonus,
			FedAt:            apitime.Milli(row.FedAt),
			CreatedAt:        apitime.Milli(row.CreatedAt),
		}
		if cattle := niu[row.NiuId]; cattle != nil {
			item.NiuName = cattle.Name
			item.NiuCode = cattle.Code
		}
		list = append(list, item)
	}
	return &ListFeedingsOutput{List: list, Total: total}, nil
}
