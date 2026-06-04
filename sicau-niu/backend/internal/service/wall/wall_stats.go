// wall_stats.go implements the public activity stats. Every figure is a
// database-side count restricted by a single indexed predicate, so the endpoint
// never loads rows into memory. The activated-cattle filter reuses the cattle
// package's stable status enum so the wall figure and the cattle contract never
// drift.

package wall

import (
	"context"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// activeNiuStatus is the persisted cattle status string the activated-cattle
// count filters on. It reuses the cattle package's stable enum constant so the
// wall stat and the cattle contract never drift.
const activeNiuStatus = string(cattlesvc.NiuStatusActive)

// Stats is the public activity statistics result.
type Stats struct {
	// ActivatedNiuCount is the number of cattle whose status is active.
	ActivatedNiuCount int64
	// TotalNiuCount is the total cattle count.
	TotalNiuCount int64
	// FirstActivatorCount is the number of first activators (is_first activations).
	FirstActivatorCount int64
	// PlayerCount is the number of participating players.
	PlayerCount int64
}

// Stats returns the public activity statistics, each counted on the database side.
func (s *serviceImpl) Stats(ctx context.Context) (*Stats, error) {
	activatedNiu, err := dao.Niu.Ctx(ctx).Where(dao.Niu.Columns().Status, activeNiuStatus).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWallQueryFailed)
	}
	totalNiu, err := dao.Niu.Ctx(ctx).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWallQueryFailed)
	}
	firstActivators, err := dao.Activation.Ctx(ctx).Where(dao.Activation.Columns().IsFirst, 1).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWallQueryFailed)
	}
	players, err := dao.User.Ctx(ctx).Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWallQueryFailed)
	}
	return &Stats{
		ActivatedNiuCount:   int64(activatedNiu),
		TotalNiuCount:       int64(totalNiu),
		FirstActivatorCount: int64(firstActivators),
		PlayerCount:         int64(players),
	}, nil
}
