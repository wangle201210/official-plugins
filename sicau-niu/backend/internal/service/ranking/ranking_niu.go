// ranking_niu.go implements the cattle feeding leaderboard. Feeding rows are
// grouped by cattle ID, summed and capped in the database; cattle labels are
// then assembled with one bounded relation query. Equal totals use competition
// ranks and cattle ID provides deterministic ordering.

package ranking

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
)

// NiuRank is one ranked cattle row on the feeding leaderboard.
type NiuRank struct {
	// Rank is the 1-based competition rank; tied totals share a rank.
	Rank int
	// NiuId is the cattle ID.
	NiuId int64
	// NiuCode is the cattle serial code.
	NiuCode string
	// NiuName is the configured cattle display name.
	NiuName string
	// Total is the cattle's accumulated feeding effect.
	Total int64
}

// niuTotalRow is the projected result of one grouped cattle aggregate.
type niuTotalRow struct {
	NiuId int64 `json:"niu_id"`
	Total int64 `json:"total"`
}

// topNiuRanks returns the database-capped cattle leaderboard with batched labels.
func (s *serviceImpl) topNiuRanks(ctx context.Context, topN int) ([]*NiuRank, error) {
	rows := make([]*niuTotalRow, 0, topN)
	err := dao.Feeding.Ctx(ctx).
		Fields(dao.Feeding.Columns().NiuId, "SUM("+dao.Feeding.Columns().EffectAmount+") AS total").
		Group(dao.Feeding.Columns().NiuId).
		Order("total DESC").
		OrderAsc(dao.Feeding.Columns().NiuId).
		Limit(topN).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRankingQueryFailed)
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.NiuId)
	}
	labels, err := s.batchNiuLabels(ctx, ids)
	if err != nil {
		return nil, err
	}

	list := make([]*NiuRank, 0, len(rows))
	rank := 0
	var previousTotal int64
	for i, row := range rows {
		if i == 0 || row.Total != previousTotal {
			rank = i + 1
		}
		label := labels[row.NiuId]
		list = append(list, &NiuRank{
			Rank: rank, NiuId: row.NiuId, NiuCode: label.Code, NiuName: label.Name, Total: row.Total,
		})
		previousTotal = row.Total
	}
	return list, nil
}
