// wall_first_activators.go implements the public first-activator wall. The
// first-activation rows (is_first) are read on the database side, ordered by
// activation time ascending and capped at 120; the player nicknames/identity
// labels and the cattle names/codes are then batch-assembled in two projected
// queries to avoid N+1. Only public fields are projected: the assembly never
// reads phone numbers, openids or device fingerprints.

package wall

import (
	"context"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// FirstActivatorBoard is the public first-activator wall result.
type FirstActivatorBoard struct {
	// List is the first activators ordered by activation time ascending, capped at
	// firstActivatorLimit.
	List []*FirstActivator
}

// FirstActivator is one first activator on the public wall. It carries only the
// player's public identity and the activated cattle, never a privacy field.
type FirstActivator struct {
	// Seq is the 1-based global position on the wall.
	Seq int
	// UserId is the player ID.
	UserId int64
	// Nickname is the player nickname; empty when unset.
	Nickname string
	// IdentityType is the player identity label; empty when unset.
	IdentityType string
	// NiuId is the activated cattle ID.
	NiuId int64
	// NiuName is the activated cattle name; empty when missing.
	NiuName string
	// NiuCode is the activated cattle code; empty when missing.
	NiuCode string
	// ActivatedAt is the activation time as Unix milliseconds; nil when unset.
	ActivatedAt *int64
}

// FirstActivators returns the bounded first-activator wall.
func (s *serviceImpl) FirstActivators(ctx context.Context) (*FirstActivatorBoard, error) {
	rows := make([]*entitymodel.Activation, 0, firstActivatorLimit)
	err := dao.Activation.Ctx(ctx).
		Fields(
			dao.Activation.Columns().UserId,
			dao.Activation.Columns().NiuId,
			dao.Activation.Columns().ActivatedAt,
		).
		Where(dao.Activation.Columns().IsFirst, 1).
		Order(dao.Activation.Columns().ActivatedAt + " ASC").
		Limit(firstActivatorLimit).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeWallQueryFailed)
	}
	list, err := s.assembleFirstActivators(ctx, rows)
	if err != nil {
		return nil, err
	}
	return &FirstActivatorBoard{List: list}, nil
}

// assembleFirstActivators projects the first-activation rows to public wall rows,
// batch-loading player public identities and cattle names/codes in one query each
// to avoid N+1 and assigning 1-based sequence numbers in the already-sorted order.
func (s *serviceImpl) assembleFirstActivators(ctx context.Context, rows []*entitymodel.Activation) ([]*FirstActivator, error) {
	if len(rows) == 0 {
		return []*FirstActivator{}, nil
	}
	userIDs := make([]int64, 0, len(rows))
	niuIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		userIDs = append(userIDs, row.UserId)
		niuIDs = append(niuIDs, row.NiuId)
	}
	players, err := s.batchPlayers(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	niu, err := s.batchNiu(ctx, niuIDs)
	if err != nil {
		return nil, err
	}

	list := make([]*FirstActivator, 0, len(rows))
	for i, row := range rows {
		item := &FirstActivator{
			Seq:         i + 1,
			UserId:      row.UserId,
			NiuId:       row.NiuId,
			ActivatedAt: apitime.Milli(row.ActivatedAt),
		}
		if player := players[row.UserId]; player != nil {
			item.Nickname = player.Nickname
			item.IdentityType = player.IdentityType
		}
		if cattle := niu[row.NiuId]; cattle != nil {
			item.NiuName = cattle.Name
			item.NiuCode = cattle.Code
		}
		list = append(list, item)
	}
	return list, nil
}
