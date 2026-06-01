// This file implements account-specific default detail creation and batch
// related-name projection for account list/detail responses.

package uidentity

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
)

// ensureAccountDetail creates an empty detail row for a newly created account.
func (s *serviceImpl) ensureAccountDetail(ctx context.Context, accountID int64) error {
	tenantID, actorID := s.baseOwnedDO(ctx, true)
	_, err := dao.AccountDetail.Ctx(ctx).Data(do.AccountDetail{
		AccountId: accountID,
		TenantId:  tenantID,
		CreatedBy: actorID,
		UpdatedBy: actorID,
	}).Insert()
	return err
}

// decorateAccountRecords adds related container/unit/group projections in batch.
func (s *serviceImpl) decorateAccountRecords(ctx context.Context, records []Record) error {
	accountIDs := make([]int64, 0, len(records))
	containerIDs := make([]int64, 0, len(records))
	unitIDs := make([]int64, 0, len(records))
	seenAccounts := make(map[int64]struct{}, len(records))
	seenContainers := make(map[int64]struct{}, len(records))
	seenUnits := make(map[int64]struct{}, len(records))
	for _, record := range records {
		accountID := gconv.Int64(record["id"])
		if accountID > 0 {
			if _, ok := seenAccounts[accountID]; !ok {
				accountIDs = append(accountIDs, accountID)
				seenAccounts[accountID] = struct{}{}
			}
		}
		containerID := gconv.Int64(record["containerId"])
		if containerID > 0 {
			if _, ok := seenContainers[containerID]; !ok {
				containerIDs = append(containerIDs, containerID)
				seenContainers[containerID] = struct{}{}
			}
		}
		unitID := gconv.Int64(record["unitId"])
		if unitID > 0 {
			if _, ok := seenUnits[unitID]; !ok {
				unitIDs = append(unitIDs, unitID)
				seenUnits[unitID] = struct{}{}
			}
		}
	}
	containerNames, err := s.nameMap(ctx, dao.Container.Ctx(ctx), dao.Container.Columns().Id, dao.Container.Columns().Alias, containerIDs)
	if err != nil {
		return err
	}
	unitNames, err := s.nameMap(ctx, dao.Unit.Ctx(ctx), dao.Unit.Columns().Id, dao.Unit.Columns().Alias, unitIDs)
	if err != nil {
		return err
	}
	groupNames, err := s.accountGroupNames(ctx, accountIDs)
	if err != nil {
		return err
	}
	for _, record := range records {
		record["containerName"] = containerNames[gconv.Int64(record["containerId"])]
		record["unitName"] = unitNames[gconv.Int64(record["unitId"])]
		record["groupNames"] = groupNames[gconv.Int64(record["id"])]
	}
	return nil
}

func (s *serviceImpl) nameMap(ctx context.Context, model *gdb.Model, idColumn string, nameColumn string, ids []int64) (map[int64]string, error) {
	names := make(map[int64]string)
	if len(ids) == 0 {
		return names, nil
	}
	result, err := s.tenantFilter.Apply(ctx, model, "").
		Fields(idColumn, nameColumn).
		WhereIn(idColumn, ids).
		All()
	if err != nil {
		return nil, err
	}
	for _, row := range result {
		names[row[idColumn].Int64()] = row[nameColumn].String()
	}
	return names, nil
}

func (s *serviceImpl) accountGroupNames(ctx context.Context, accountIDs []int64) (map[int64][]string, error) {
	groupNames := make(map[int64][]string)
	if len(accountIDs) == 0 {
		return groupNames, nil
	}
	accountGroupColumns := dao.AccountGroup.Columns()
	groupColumns := dao.Group.Columns()
	relations, err := s.tenantFilter.Apply(ctx, dao.AccountGroup.Ctx(ctx), "").
		Fields(accountGroupColumns.AccountId, accountGroupColumns.GroupId).
		WhereIn(accountGroupColumns.AccountId, accountIDs).
		All()
	if err != nil {
		return nil, err
	}
	groupIDs := make([]int64, 0, len(relations))
	seen := make(map[int64]struct{}, len(relations))
	for _, relation := range relations {
		groupID := relation[accountGroupColumns.GroupId].Int64()
		if groupID > 0 {
			if _, ok := seen[groupID]; !ok {
				groupIDs = append(groupIDs, groupID)
				seen[groupID] = struct{}{}
			}
		}
	}
	names, err := s.nameMap(ctx, dao.Group.Ctx(ctx), groupColumns.Id, groupColumns.Alias, groupIDs)
	if err != nil {
		return nil, err
	}
	for _, relation := range relations {
		accountID := relation[accountGroupColumns.AccountId].Int64()
		groupID := relation[accountGroupColumns.GroupId].Int64()
		if name := strings.TrimSpace(names[groupID]); name != "" {
			groupNames[accountID] = append(groupNames[accountID], name)
		}
	}
	return groupNames, nil
}
