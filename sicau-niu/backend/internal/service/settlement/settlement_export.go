// settlement_export.go implements the player roster export. A bounded page of
// players is read, then the per-player activation count, feeding total and college
// name are batch-assembled in one grouped query each, so the export never issues a
// per-row lookup. When more players exist than the export cap the page is truncated
// and the result flags it rather than silently dropping rows.

package settlement

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// PlayerExport is the player roster export result.
type PlayerExport struct {
	// List is the bounded player roster rows.
	List []*PlayerExportRow
	// Total is the total player count before the export cap.
	Total int64
	// Truncated reports whether the roster was truncated at the export cap.
	Truncated bool
}

// PlayerExportRow is one player row in the roster export.
type PlayerExportRow struct {
	UserId          int64
	Nickname        string
	IdentityType    string
	CollegeName     string
	Grade           int
	ActivationCount int64
	FeedTotalEffect int64
}

// countRow is the temporary projection for one grouped id/value aggregate.
type countRow struct {
	Id  int64 `json:"id"`
	Val int64 `json:"val"`
}

// ExportPlayers returns the bounded player roster with batch-assembled aggregates.
func (s *serviceImpl) ExportPlayers(ctx context.Context) (*PlayerExport, error) {
	total, err := dao.User.Ctx(ctx).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}

	// Fetch one more than the cap so truncation can be detected precisely.
	pageRows := make([]*entitymodel.User, 0, exportMaxRows)
	err = dao.User.Ctx(ctx).
		Fields(
			dao.User.Columns().Id,
			dao.User.Columns().Nickname,
			dao.User.Columns().IdentityType,
			dao.User.Columns().CollegeId,
			dao.User.Columns().Grade,
		).
		Order(dao.User.Columns().Id + " ASC").
		Limit(exportMaxRows + 1).
		Scan(&pageRows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	truncated := false
	if len(pageRows) > exportMaxRows {
		pageRows = pageRows[:exportMaxRows]
		truncated = true
	}
	if len(pageRows) == 0 {
		return &PlayerExport{List: []*PlayerExportRow{}, Total: int64(total), Truncated: truncated}, nil
	}

	userIDs := make([]int64, 0, len(pageRows))
	collegeIDs := make([]int64, 0, len(pageRows))
	for _, row := range pageRows {
		userIDs = append(userIDs, row.Id)
		if row.CollegeId > 0 {
			collegeIDs = append(collegeIDs, row.CollegeId)
		}
	}

	activationCounts, err := s.batchUserCounts(ctx, dao.Activation.Ctx(ctx).
		WhereIn(dao.Activation.Columns().UserId, userIDs).
		Fields(dao.Activation.Columns().UserId+" AS id", "COUNT(*) AS val").
		Group(dao.Activation.Columns().UserId))
	if err != nil {
		return nil, err
	}
	feedTotals, err := s.batchUserCounts(ctx, dao.Feeding.Ctx(ctx).
		WhereIn(dao.Feeding.Columns().UserId, userIDs).
		Fields(dao.Feeding.Columns().UserId+" AS id", "SUM("+dao.Feeding.Columns().EffectAmount+") AS val").
		Group(dao.Feeding.Columns().UserId))
	if err != nil {
		return nil, err
	}
	collegeNames, err := s.batchCollegeNames(ctx, collegeIDs)
	if err != nil {
		return nil, err
	}

	list := make([]*PlayerExportRow, 0, len(pageRows))
	for _, row := range pageRows {
		list = append(list, &PlayerExportRow{
			UserId:          row.Id,
			Nickname:        row.Nickname,
			IdentityType:    row.IdentityType,
			CollegeName:     collegeNames[row.CollegeId],
			Grade:           row.Grade,
			ActivationCount: activationCounts[row.Id],
			FeedTotalEffect: feedTotals[row.Id],
		})
	}
	return &PlayerExport{List: list, Total: int64(total), Truncated: truncated}, nil
}

// batchUserCounts runs the prepared grouped id/val projection and returns a
// user-ID to value map.
func (s *serviceImpl) batchUserCounts(ctx context.Context, model *gdb.Model) (map[int64]int64, error) {
	_ = ctx
	var rows []*countRow
	if err := model.Scan(&rows); err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	out := make(map[int64]int64, len(rows))
	for _, row := range rows {
		out[row.Id] = row.Val
	}
	return out, nil
}

// batchCollegeNames returns a college-ID to name map for ids in one projected
// query. An empty id set short-circuits without a query.
func (s *serviceImpl) batchCollegeNames(ctx context.Context, ids []int64) (map[int64]string, error) {
	names := make(map[int64]string, len(ids))
	if len(ids) == 0 {
		return names, nil
	}
	rows := make([]*entitymodel.College, 0, len(ids))
	err := dao.College.Ctx(ctx).
		Fields(dao.College.Columns().Id, dao.College.Columns().Name).
		WhereIn(dao.College.Columns().Id, ids).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeSettlementQueryFailed)
	}
	for _, row := range rows {
		names[row.Id] = row.Name
	}
	return names, nil
}
