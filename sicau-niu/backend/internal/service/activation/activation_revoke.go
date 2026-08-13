// activation_revoke.go implements the operator-only erroneous activation repair.
package activation

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// Revoke soft-deletes one activation and deterministically repairs first state.
func (s *serviceImpl) Revoke(ctx context.Context, activationID int64) error {
	if activationID <= 0 {
		return bizerr.NewCode(CodeActivationNotFound)
	}
	return dao.Activation.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		var record *entitymodel.Activation
		if err := dao.Activation.Ctx(ctx).Where(do.Activation{Id: activationID}).LockUpdate().Scan(&record); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if record == nil {
			return bizerr.NewCode(CodeActivationNotFound)
		}
		var niu *entitymodel.Niu
		if err := dao.Niu.Ctx(ctx).Where(do.Niu{Id: record.NiuId}).LockUpdate().Scan(&niu); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if _, err := dao.Activation.Ctx(ctx).Where(do.Activation{Id: activationID}).Delete(); err != nil {
			return bizerr.WrapCode(err, CodeWriteFailed)
		}
		if record.IsFirst != firstActivatorFlag {
			return nil
		}
		var next *entitymodel.Activation
		if err := dao.Activation.Ctx(ctx).
			Where(dao.Activation.Columns().NiuId, record.NiuId).
			OrderAsc(dao.Activation.Columns().ActivatedAt).
			OrderAsc(dao.Activation.Columns().Id).
			Limit(1).LockUpdate().Scan(&next); err != nil {
			return bizerr.WrapCode(err, CodeQueryFailed)
		}
		if next != nil {
			if _, err = dao.Activation.Ctx(ctx).Where(do.Activation{Id: next.Id}).Data(do.Activation{
				IsFirst: firstActivatorFlag,
			}).Update(); err != nil {
				return bizerr.WrapCode(err, CodeWriteFailed)
			}
			return nil
		}
		if niu != nil {
			if _, err := dao.Niu.Ctx(ctx).Where(do.Niu{Id: record.NiuId}).Data(do.Niu{Status: cattlesvc.NiuStatusInactive.String()}).Update(); err != nil {
				return bizerr.WrapCode(err, CodeWriteFailed)
			}
		}
		return nil
	})
}
