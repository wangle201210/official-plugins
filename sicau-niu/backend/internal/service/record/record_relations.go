// record_relations.go holds the batch relation-assembly helpers shared by the
// record queries: player nicknames keyed by user ID and cattle (name + code) keyed
// by cattle ID. Each helper resolves a whole page in one projected WHERE IN query
// to avoid per-row lookups.

package record

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// batchNicknames returns a user-ID to nickname map for ids, fetched in one
// projected query. An empty id set short-circuits without a query.
func batchNicknames(ctx context.Context, ids []int64) (map[int64]string, error) {
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
		return nil, bizerr.WrapCode(err, CodeRecordQueryFailed)
	}
	for _, row := range rows {
		names[row.Id] = row.Nickname
	}
	return names, nil
}

// batchNiu returns a cattle-ID to cattle (name + code) map for ids, fetched in one
// projected query. An empty id set short-circuits without a query.
func batchNiu(ctx context.Context, ids []int64) (map[int64]*entitymodel.Niu, error) {
	niu := make(map[int64]*entitymodel.Niu, len(ids))
	if len(ids) == 0 {
		return niu, nil
	}
	rows := make([]*entitymodel.Niu, 0, len(ids))
	err := dao.Niu.Ctx(ctx).
		Fields(dao.Niu.Columns().Id, dao.Niu.Columns().Name, dao.Niu.Columns().Code).
		WhereIn(dao.Niu.Columns().Id, ids).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeRecordQueryFailed)
	}
	for _, row := range rows {
		niu[row.Id] = row
	}
	return niu, nil
}

// dedupeInt64 returns the distinct non-zero values of ids preserving first-seen
// order, used to collect the user/cattle keys before batch assembly.
func dedupeInt64(ids []int64) []int64 {
	seen := make(map[int64]struct{}, len(ids))
	out := make([]int64, 0, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}
