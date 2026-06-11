// record_activation_attempt.go implements the read-only paged photo check-in
// attempt audit query. It keeps failed location-matching attempts separate from
// successful activation records while sharing the same operator record permission.

package record

import (
	"context"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// Activation attempt result constants persisted by the activation service.
const (
	ActivationAttemptResultSuccess    = "success"
	ActivationAttemptResultNoNearby   = "no_nearby"
	ActivationAttemptResultOutOfRange = "out_of_range"
)

// ListActivationAttemptsInput is the activation-attempt audit query.
type ListActivationAttemptsInput struct {
	UserId   int64
	NiuId    int64
	Result   string
	PageNum  int
	PageSize int
}

// ListActivationAttemptsOutput is the activation-attempt audit query result.
type ListActivationAttemptsOutput struct {
	List  []*ActivationAttemptItem
	Total int
}

// ActivationAttemptItem is one photo check-in attempt row projected for operators.
type ActivationAttemptItem struct {
	Id             int64
	UserId         int64
	Nickname       string
	NiuId          int64
	NiuName        string
	NiuCode        string
	NearestNiuId   int64
	NearestNiuName string
	NearestNiuCode string
	Result         string
	Lat            float64
	Lng            float64
	DistanceM      float64
	ThresholdM     float64
	PhotoPath      string
	AttemptedAt    *int64
	CreatedAt      *int64
}

// ListActivationAttempts returns one DB-side paged photo check-in attempt page.
func (s *serviceImpl) ListActivationAttempts(ctx context.Context, in *ListActivationAttemptsInput) (*ListActivationAttemptsOutput, error) {
	pageNum, pageSize := normalizePagination(in.PageNum, in.PageSize)
	model := dao.ActivationAttempt.Ctx(ctx)
	columns := dao.ActivationAttempt.Columns()
	if in.UserId > 0 {
		model = model.Where(columns.UserId, in.UserId)
	}
	if in.NiuId > 0 {
		niuFilter := model.Builder().
			Where(columns.NiuId, in.NiuId).
			WhereOr(columns.NearestNiuId, in.NiuId)
		model = model.Where(niuFilter)
	}
	if isActivationAttemptResult(in.Result) {
		model = model.Where(columns.Result, in.Result)
	}

	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRecordQueryFailed)
	}

	rows := make([]*entitymodel.ActivationAttempt, 0, pageSize)
	err = model.
		Fields(
			columns.Id,
			columns.UserId,
			columns.NiuId,
			columns.NearestNiuId,
			columns.Result,
			columns.Lat,
			columns.Lng,
			columns.DistanceM,
			columns.ThresholdM,
			columns.PhotoPath,
			columns.AttemptedAt,
			columns.CreatedAt,
		).
		Page(pageNum, pageSize).
		OrderDesc(columns.AttemptedAt).
		OrderDesc(columns.Id).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRecordQueryFailed)
	}

	userIDs := make([]int64, 0, len(rows))
	niuIDs := make([]int64, 0, len(rows)*2)
	for _, row := range rows {
		userIDs = append(userIDs, row.UserId)
		niuIDs = append(niuIDs, row.NiuId, row.NearestNiuId)
	}
	nicknames, err := batchNicknames(ctx, dedupeInt64(userIDs))
	if err != nil {
		return nil, err
	}
	niu, err := batchNiu(ctx, dedupeInt64(niuIDs))
	if err != nil {
		return nil, err
	}

	list := make([]*ActivationAttemptItem, 0, len(rows))
	for _, row := range rows {
		item := &ActivationAttemptItem{
			Id:           row.Id,
			UserId:       row.UserId,
			Nickname:     nicknames[row.UserId],
			NiuId:        row.NiuId,
			NearestNiuId: row.NearestNiuId,
			Result:       row.Result,
			Lat:          row.Lat,
			Lng:          row.Lng,
			DistanceM:    row.DistanceM,
			ThresholdM:   row.ThresholdM,
			PhotoPath:    row.PhotoPath,
			AttemptedAt:  apitime.Milli(row.AttemptedAt),
			CreatedAt:    apitime.Milli(row.CreatedAt),
		}
		if cattle := niu[row.NiuId]; cattle != nil {
			item.NiuName = cattle.Name
			item.NiuCode = cattle.Code
		}
		if cattle := niu[row.NearestNiuId]; cattle != nil {
			item.NearestNiuName = cattle.Name
			item.NearestNiuCode = cattle.Code
		}
		list = append(list, item)
	}
	return &ListActivationAttemptsOutput{List: list, Total: total}, nil
}

func isActivationAttemptResult(result string) bool {
	switch result {
	case ActivationAttemptResultSuccess, ActivationAttemptResultNoNearby, ActivationAttemptResultOutOfRange:
		return true
	default:
		return false
	}
}
