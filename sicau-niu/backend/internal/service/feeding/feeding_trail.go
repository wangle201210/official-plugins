// feeding_trail.go implements the player's recent feeding trail. It loads the
// player's latest feedings in one DB-side ordered, limited query, then
// batch-assembles the fed cattle code/name in one WhereIn query to avoid N+1.

package feeding

import (
	"context"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// trailLimit bounds the recent feeding trail to the latest entries.
const trailLimit = 10

// TrailItem is one recent feeding projected for the player trail.
type TrailItem struct {
	// NiuId is the fed cattle ID.
	NiuId int64
	// NiuCode is the fed cattle serial code.
	NiuCode string
	// NiuName is the fed cattle name; empty for common cattle.
	NiuName string
	// EffectAmount is the actual feeding effect for this record.
	EffectAmount int
	// IsIronBonus reports whether an iron-cow proximity bonus applied.
	IsIronBonus bool
	// FedAt is the feed time as a Unix timestamp in milliseconds.
	FedAt *int64
}

// Trail returns the player's latest feedings with batch-assembled cattle info.
func (s *serviceImpl) Trail(ctx context.Context, playerID int64) ([]*TrailItem, error) {
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeQueryFailed)
	}

	rows := make([]*entitymodel.Feeding, 0, trailLimit)
	err := dao.Feeding.Ctx(ctx).
		Where(dao.Feeding.Columns().UserId, playerID).
		OrderDesc(dao.Feeding.Columns().FedAt).
		OrderDesc(dao.Feeding.Columns().Id).
		Limit(trailLimit).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}
	if len(rows) == 0 {
		return []*TrailItem{}, nil
	}

	niuInfo, err := s.loadNiuInfo(ctx, rows)
	if err != nil {
		return nil, err
	}

	items := make([]*TrailItem, 0, len(rows))
	for _, row := range rows {
		info := niuInfo[row.NiuId]
		items = append(items, &TrailItem{
			NiuId:        row.NiuId,
			NiuCode:      info.code,
			NiuName:      info.name,
			EffectAmount: row.EffectAmount,
			IsIronBonus:  row.IsIronBonus == 1,
			FedAt:        apitime.Milli(row.FedAt),
		})
	}
	return items, nil
}

// niuInfo holds the projected cattle fields used by the feeding trail.
type niuInfo struct {
	code string
	name string
}

// loadNiuInfo batch-loads the code and name for the distinct cattle in the trail
// page in one WhereIn query and returns them keyed by cattle ID, avoiding a
// per-row cattle lookup.
func (s *serviceImpl) loadNiuInfo(ctx context.Context, rows []*entitymodel.Feeding) (map[int64]niuInfo, error) {
	idSet := make(map[int64]struct{}, len(rows))
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		if _, seen := idSet[row.NiuId]; seen {
			continue
		}
		idSet[row.NiuId] = struct{}{}
		ids = append(ids, row.NiuId)
	}

	niuRows := make([]*entitymodel.Niu, 0, len(ids))
	err := dao.Niu.Ctx(ctx).
		Fields(
			dao.Niu.Columns().Id,
			dao.Niu.Columns().Code,
			dao.Niu.Columns().Name,
		).
		WhereIn(dao.Niu.Columns().Id, ids).
		Scan(&niuRows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}

	info := make(map[int64]niuInfo, len(niuRows))
	for _, niuRow := range niuRows {
		info[niuRow.Id] = niuInfo{code: niuRow.Code, name: niuRow.Name}
	}
	return info, nil
}
