// wall_relations.go holds the batch relation-assembly helpers shared by the wall
// views: player public identities keyed by user ID and cattle name/code keyed by
// cattle ID. Each helper resolves a whole wall page in one projected WHERE IN
// query to avoid per-row lookups. The player projection selects only public
// columns (nickname and identity label); it never reads phone numbers, openids or
// device fingerprints.

package wall

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// batchPlayers returns a user-ID to public-identity map for ids, fetched in one
// projected query that selects only the nickname and identity label. An empty id
// set short-circuits without a query.
func (s *serviceImpl) batchPlayers(ctx context.Context, ids []int64) (map[int64]*entitymodel.User, error) {
	players := make(map[int64]*entitymodel.User, len(ids))
	if len(ids) == 0 {
		return players, nil
	}
	rows := make([]*entitymodel.User, 0, len(ids))
	err := dao.User.Ctx(ctx).
		Fields(
			dao.User.Columns().Id,
			dao.User.Columns().Nickname,
			dao.User.Columns().IdentityType,
		).
		WhereIn(dao.User.Columns().Id, ids).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWallQueryFailed)
	}
	for _, row := range rows {
		players[row.Id] = row
	}
	return players, nil
}

// batchNiu returns a cattle-ID to cattle map (name and code only) for ids,
// fetched in one projected query. An empty id set short-circuits without a query.
func (s *serviceImpl) batchNiu(ctx context.Context, ids []int64) (map[int64]*entitymodel.Niu, error) {
	niu := make(map[int64]*entitymodel.Niu, len(ids))
	if len(ids) == 0 {
		return niu, nil
	}
	rows := make([]*entitymodel.Niu, 0, len(ids))
	err := dao.Niu.Ctx(ctx).
		Fields(
			dao.Niu.Columns().Id,
			dao.Niu.Columns().Name,
			dao.Niu.Columns().Code,
		).
		WhereIn(dao.Niu.Columns().Id, ids).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWallQueryFailed)
	}
	for _, row := range rows {
		niu[row.Id] = row
	}
	return niu, nil
}
