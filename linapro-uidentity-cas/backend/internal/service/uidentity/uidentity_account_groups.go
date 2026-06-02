// This file restores legacy account request groupIds handling by synchronizing
// account_group rows from account create/update resource payloads.

package uidentity

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
)

// syncAccountGroupsFromBody applies legacy account payload groupIds semantics.
func (s *serviceImpl) syncAccountGroupsFromBody(ctx context.Context, accountID int64, body map[string]any) error {
	groupIDs, ok := accountGroupIDsFromBody(body)
	if !ok {
		return nil
	}
	return s.replaceAccountGroups(ctx, accountID, groupIDs)
}

// accountGroupIDsFromBody extracts normalized groupIds when present.
func accountGroupIDsFromBody(body map[string]any) ([]int64, bool) {
	if !hasField(body, "groupIds") {
		return nil, false
	}
	return uniquePositiveInt64s(gconv.SliceInt64(body["groupIds"])), true
}

// validateAccountGroups rejects missing or cross-tenant group IDs before writes.
func (s *serviceImpl) validateAccountGroups(ctx context.Context, groupIDs []int64) error {
	if len(groupIDs) == 0 {
		return nil
	}
	groupCount, err := s.tenantFilter.Apply(ctx, dao.Group.Ctx(ctx), "").
		WhereIn(dao.Group.Columns().Id, groupIDs).
		Count()
	if err != nil {
		return err
	}
	if groupCount != len(groupIDs) {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	return nil
}

// replaceAccountGroups replaces all tenant-visible account group relations.
func (s *serviceImpl) replaceAccountGroups(ctx context.Context, accountID int64, groupIDs []int64) error {
	if accountID <= 0 {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	tenantID, actorID := s.baseOwnedDO(ctx, true)
	return dao.AccountGroup.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if err := s.validateAccountGroups(ctx, groupIDs); err != nil {
			return err
		}
		if _, err := s.tenantFilter.Apply(ctx, dao.AccountGroup.Ctx(ctx), "").
			Where(dao.AccountGroup.Columns().AccountId, accountID).
			Delete(); err != nil {
			return err
		}
		for _, groupID := range groupIDs {
			if _, err := dao.AccountGroup.Ctx(ctx).Data(do.AccountGroup{
				TenantId:  tenantID,
				AccountId: accountID,
				GroupId:   groupID,
				CreatedBy: actorID,
			}).Insert(); err != nil {
				return err
			}
		}
		return nil
	})
}

// uniquePositiveInt64s removes duplicate, zero, and negative IDs.
func uniquePositiveInt64s(values []int64) []int64 {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
