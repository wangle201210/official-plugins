// record_checkin.go implements the read-only paged check-in-record query with
// player nicknames batch-assembled.

package record

import (
	"context"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// ListCheckinsInput is the check-in-record query: optional player filter and
// pagination.
type ListCheckinsInput struct {
	UserId   int64
	PageNum  int
	PageSize int
}

// ListCheckinsOutput is the check-in-record query result.
type ListCheckinsOutput struct {
	List  []*CheckinItem
	Total int
}

// CheckinItem is one check-in record projected for the operator console.
type CheckinItem struct {
	Id          int64
	UserId      int64
	Nickname    string
	CheckinDate string
	Amount      int
	CreatedAt   *int64
}

// ListCheckins returns one DB-side paged check-in-record page.
func (s *serviceImpl) ListCheckins(ctx context.Context, in *ListCheckinsInput) (*ListCheckinsOutput, error) {
	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	model := dao.Checkin.Ctx(ctx)
	if in.UserId > 0 {
		model = model.Where(dao.Checkin.Columns().UserId, in.UserId)
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRecordQueryFailed)
	}

	rows := make([]*entitymodel.Checkin, 0, pageSize)
	err = model.
		Page(pageNum, pageSize).
		OrderDesc(dao.Checkin.Columns().CreatedAt).
		OrderDesc(dao.Checkin.Columns().Id).
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

	list := make([]*CheckinItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, &CheckinItem{
			Id:          row.Id,
			UserId:      row.UserId,
			Nickname:    nicknames[row.UserId],
			CheckinDate: row.CheckinDate,
			Amount:      row.Amount,
			CreatedAt:   apitime.Milli(row.CreatedAt),
		})
	}
	return &ListCheckinsOutput{List: list, Total: total}, nil
}
