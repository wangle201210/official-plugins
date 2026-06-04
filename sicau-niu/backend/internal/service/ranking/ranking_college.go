// ranking_college.go implements the college leaderboard. Feeding rows are joined
// to the player table on the database side, restricted to enrolled students with
// a positive college ID, grouped by college and summed, then ordered descending
// with a LIMIT. The college names are batch-assembled in one query. All filtering,
// grouping, ordering and the Top-N cap run in the database so the full set is
// never loaded into memory.

package ranking

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
)

// CollegeBoard is the college leaderboard result.
type CollegeBoard struct {
	// List is the Top-N colleges ordered by total feeding effect descending.
	List []*CollegeRank
}

// CollegeRank is one ranked college on the college board.
type CollegeRank struct {
	// Rank is the 1-based position on the board.
	Rank int
	// CollegeId is the college ID.
	CollegeId int64
	// CollegeName is the college name; empty when the college is missing.
	CollegeName string
	// Total is the total feeding effect of the college's enrolled students.
	Total int64
}

// collegeTotalRow is the temporary projection for one grouped college/total
// aggregate.
type collegeTotalRow struct {
	CollegeId int64 `json:"college_id"`
	Total     int64 `json:"total"`
}

// userTable is the qualified player table name used to build the inner join in
// the college aggregation; it mirrors the DAO-owned table name.
const userTable = "plugin_sicau_niu_user"

// CollegeBoard returns the Top-N college board aggregated from enrolled students.
func (s *serviceImpl) CollegeBoard(ctx context.Context) (*CollegeBoard, error) {
	rows := make([]*collegeTotalRow, 0, s.topN)
	err := dao.Feeding.Ctx(ctx).
		As("f").
		InnerJoin(
			userTable+" AS u",
			"u."+dao.User.Columns().Id+" = f."+dao.Feeding.Columns().UserId+
				" AND u."+dao.User.Columns().DeletedAt+" IS NULL",
		).
		Where("u."+dao.User.Columns().IdentityType, studentIdentity).
		Where("u."+dao.User.Columns().CollegeId+" > ?", 0).
		Fields(
			"u."+dao.User.Columns().CollegeId+" AS college_id",
			"SUM(f."+dao.Feeding.Columns().EffectAmount+") AS total",
		).
		Group("u." + dao.User.Columns().CollegeId).
		Order("total DESC").
		Limit(s.topN).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRankingQueryFailed)
	}

	list, err := s.assembleCollegeRanks(ctx, rows)
	if err != nil {
		return nil, err
	}
	return &CollegeBoard{List: list}, nil
}

// assembleCollegeRanks projects the grouped college totals to ranked rows,
// batch-loading the college names in one query to avoid N+1 and assigning 1-based
// ranks in the already-sorted order.
func (s *serviceImpl) assembleCollegeRanks(ctx context.Context, rows []*collegeTotalRow) ([]*CollegeRank, error) {
	if len(rows) == 0 {
		return []*CollegeRank{}, nil
	}
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.CollegeId)
	}
	names, err := s.batchCollegeNames(ctx, ids)
	if err != nil {
		return nil, err
	}
	list := make([]*CollegeRank, 0, len(rows))
	for i, row := range rows {
		list = append(list, &CollegeRank{
			Rank:        i + 1,
			CollegeId:   row.CollegeId,
			CollegeName: names[row.CollegeId],
			Total:       row.Total,
		})
	}
	return list, nil
}
