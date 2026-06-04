// ranking_relations.go holds the batch relation-assembly helpers shared by the
// boards: player nicknames keyed by user ID and college names keyed by college
// ID. Each helper resolves a whole board page in one projected WHERE IN query to
// avoid per-row lookups.

package ranking

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// batchNicknames returns a user-ID to nickname map for ids, fetched in one
// projected query. An empty id set short-circuits without a query.
func (s *serviceImpl) batchNicknames(ctx context.Context, ids []int64) (map[int64]string, error) {
	names := make(map[int64]string, len(ids))
	if len(ids) == 0 {
		return names, nil
	}
	rows := make([]*entitymodel.User, 0, len(ids))
	err := dao.User.Ctx(ctx).
		Fields(dao.User.Columns().Id, dao.User.Columns().Nickname).
		WhereIn(dao.User.Columns().Id, ids).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRankingQueryFailed)
	}
	for _, row := range rows {
		names[row.Id] = row.Nickname
	}
	return names, nil
}

// batchCollegeNames returns a college-ID to name map for ids, fetched in one
// projected query. An empty id set short-circuits without a query.
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
		return nil, bizerr.WrapCode(err, CodeRankingQueryFailed)
	}
	for _, row := range rows {
		names[row.Id] = row.Name
	}
	return names, nil
}
