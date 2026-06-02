// This file restores legacy account/account-detail change auditing for plugin
// write paths that no longer have old GORM model hooks available.

package uidentity

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/util/gconv"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
)

const (
	accountAuditActionCreate = "create"
	accountAuditActionUpdate = "update"
	accountAuditActionDelete = "delete"
	accountAuditTableAccount = "account"
	accountAuditTableDetail  = "account_details"
)

func (s *serviceImpl) createAccountWithAudit(ctx context.Context, data any) (int64, error) {
	id, err := dao.Account.Ctx(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	if err := s.recordAccountAudit(ctx, id, accountAuditTableAccount, accountAuditActionCreate, nil, func(ctx context.Context, accountID int64) (Record, error) {
		return s.accountAuditRecord(ctx, accountID)
	}); err != nil {
		return 0, err
	}
	return id, nil
}

func (s *serviceImpl) updateAccountWithAudit(ctx context.Context, accountID int64, data any) error {
	return s.recordAccountAudit(ctx, accountID, accountAuditTableAccount, accountAuditActionUpdate, func(ctx context.Context, accountID int64) (Record, error) {
		return s.accountAuditRecord(ctx, accountID)
	}, func(ctx context.Context, accountID int64) (Record, error) {
		_, err := s.tenantFilter.Apply(ctx, dao.Account.Ctx(ctx), "").
			Where(dao.Account.Columns().Id, accountID).
			OmitNilData().
			Data(data).
			Update()
		if err != nil {
			return nil, err
		}
		return s.accountAuditRecord(ctx, accountID)
	})
}

func (s *serviceImpl) updateAccountWithAuditForTenant(ctx context.Context, tenantID int, accountID int64, data any) error {
	return s.recordAccountAuditForTenant(ctx, tenantID, accountID, accountAuditTableAccount, accountAuditActionUpdate, func(ctx context.Context, tenantID int, accountID int64) (Record, error) {
		return s.accountAuditRecordForTenant(ctx, tenantID, accountID)
	}, func(ctx context.Context, tenantID int, accountID int64) (Record, error) {
		_, err := dao.Account.Ctx(ctx).
			Where(dao.Account.Columns().TenantId, tenantID).
			Where(dao.Account.Columns().Id, accountID).
			OmitNilData().
			Data(data).
			Update()
		if err != nil {
			return nil, err
		}
		return s.accountAuditRecordForTenant(ctx, tenantID, accountID)
	})
}

func (s *serviceImpl) createAccountDetailWithAudit(ctx context.Context, data any, accountID int64) error {
	if _, err := dao.AccountDetail.Ctx(ctx).Data(data).Insert(); err != nil {
		return err
	}
	return s.recordAccountAudit(ctx, accountID, accountAuditTableDetail, accountAuditActionCreate, nil, func(ctx context.Context, accountID int64) (Record, error) {
		return s.accountDetailAuditRecord(ctx, accountID)
	})
}

func (s *serviceImpl) createAccountDetailWithAuditForTenant(ctx context.Context, tenantID int, data any, accountID int64) error {
	if _, err := dao.AccountDetail.Ctx(ctx).Data(data).Insert(); err != nil {
		return err
	}
	return s.recordAccountAuditForTenant(ctx, tenantID, accountID, accountAuditTableDetail, accountAuditActionCreate, nil, func(ctx context.Context, tenantID int, accountID int64) (Record, error) {
		return s.accountDetailAuditRecordForTenant(ctx, tenantID, accountID)
	})
}

func (s *serviceImpl) updateAccountDetailWithAudit(ctx context.Context, accountID int64, data any) error {
	return s.recordAccountAudit(ctx, accountID, accountAuditTableDetail, accountAuditActionUpdate, func(ctx context.Context, accountID int64) (Record, error) {
		return s.accountDetailAuditRecord(ctx, accountID)
	}, func(ctx context.Context, accountID int64) (Record, error) {
		_, err := s.tenantFilter.Apply(ctx, dao.AccountDetail.Ctx(ctx), "").
			Where(dao.AccountDetail.Columns().AccountId, accountID).
			OmitNilData().
			Data(data).
			Update()
		if err != nil {
			return nil, err
		}
		return s.accountDetailAuditRecord(ctx, accountID)
	})
}

func (s *serviceImpl) updateAccountDetailWithAuditForTenant(ctx context.Context, tenantID int, accountID int64, data any) error {
	return s.recordAccountAuditForTenant(ctx, tenantID, accountID, accountAuditTableDetail, accountAuditActionUpdate, func(ctx context.Context, tenantID int, accountID int64) (Record, error) {
		return s.accountDetailAuditRecordForTenant(ctx, tenantID, accountID)
	}, func(ctx context.Context, tenantID int, accountID int64) (Record, error) {
		_, err := dao.AccountDetail.Ctx(ctx).
			Where(dao.AccountDetail.Columns().TenantId, tenantID).
			Where(dao.AccountDetail.Columns().AccountId, accountID).
			OmitNilData().
			Data(data).
			Update()
		if err != nil {
			return nil, err
		}
		return s.accountDetailAuditRecordForTenant(ctx, tenantID, accountID)
	})
}

func (s *serviceImpl) deleteResourceWithAccountAudit(ctx context.Context, def *resourceDefinition, idList []int64) error {
	if def.name != "accounts" && def.name != "account-details" {
		_, err := s.tenantFilter.Apply(ctx, def.model(ctx), "").
			WhereIn(def.idColumn, idList).
			Delete()
		return err
	}
	for _, id := range idList {
		if def.name == "accounts" {
			if err := s.deleteAccountWithAudit(ctx, id); err != nil {
				return err
			}
			continue
		}
		if err := s.deleteAccountDetailWithAudit(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (s *serviceImpl) deleteAccountWithAudit(ctx context.Context, accountID int64) error {
	oldData, err := s.accountAuditRecord(ctx, accountID)
	if err != nil {
		return err
	}
	if _, err := s.tenantFilter.Apply(ctx, dao.Account.Ctx(ctx), "").
		Where(dao.Account.Columns().Id, accountID).
		Delete(); err != nil {
		return err
	}
	return s.insertAccountAudit(ctx, accountID, accountAuditTableAccount, accountAuditActionDelete, oldData, nil)
}

func (s *serviceImpl) deleteAccountDetailWithAudit(ctx context.Context, accountID int64) error {
	oldData, err := s.accountDetailAuditRecord(ctx, accountID)
	if err != nil {
		return err
	}
	if _, err := s.tenantFilter.Apply(ctx, dao.AccountDetail.Ctx(ctx), "").
		Where(dao.AccountDetail.Columns().AccountId, accountID).
		Delete(); err != nil {
		return err
	}
	return s.insertAccountAudit(ctx, accountID, accountAuditTableDetail, accountAuditActionDelete, oldData, nil)
}

func (s *serviceImpl) recordAccountAudit(ctx context.Context, accountID int64, tableName string, action string, oldData func(context.Context, int64) (Record, error), writeData func(context.Context, int64) (Record, error)) error {
	var previous Record
	if oldData != nil {
		record, err := oldData(ctx, accountID)
		if err != nil {
			return err
		}
		previous = record
	}
	current, err := writeData(ctx, accountID)
	if err != nil {
		return err
	}
	return s.insertAccountAudit(ctx, accountID, tableName, action, previous, current)
}

func (s *serviceImpl) recordAccountAuditForTenant(ctx context.Context, tenantID int, accountID int64, tableName string, action string, oldData func(context.Context, int, int64) (Record, error), writeData func(context.Context, int, int64) (Record, error)) error {
	var previous Record
	if oldData != nil {
		record, err := oldData(ctx, tenantID, accountID)
		if err != nil {
			return err
		}
		previous = record
	}
	current, err := writeData(ctx, tenantID, accountID)
	if err != nil {
		return err
	}
	return s.insertAccountAuditForTenant(ctx, tenantID, accountID, tableName, action, previous, current)
}

func (s *serviceImpl) insertAccountAudit(ctx context.Context, accountID int64, tableName string, action string, oldData Record, newData Record) error {
	return s.insertAccountAuditForTenant(ctx, s.tenantID(ctx), accountID, tableName, action, oldData, newData)
}

func (s *serviceImpl) insertAccountAuditForTenant(ctx context.Context, tenantID int, accountID int64, tableName string, action string, oldData Record, newData Record) error {
	actorID := s.actorID(ctx)
	_, err := dao.AccountChangeLog.Ctx(ctx).Data(do.AccountChangeLog{
		TenantId:  tenantID,
		AccountId: accountID,
		TableName: tableName,
		Action:    action,
		DataOld:   marshalAuditRecord(oldData),
		DataNew:   marshalAuditRecord(newData),
		CreatedBy: actorID,
		UpdatedBy: actorID,
	}).Insert()
	return err
}

func (s *serviceImpl) accountAuditRecord(ctx context.Context, accountID int64) (Record, error) {
	return s.accountAuditRecordByModel(ctx, accountID, s.tenantFilter.Apply(ctx, dao.Account.Ctx(ctx), ""))
}

func (s *serviceImpl) accountAuditRecordForTenant(ctx context.Context, tenantID int, accountID int64) (Record, error) {
	return s.accountAuditRecordByModel(ctx, accountID, dao.Account.Ctx(ctx).Where(dao.Account.Columns().TenantId, tenantID))
}

func (s *serviceImpl) accountAuditRecordByModel(ctx context.Context, accountID int64, model *gdb.Model) (Record, error) {
	def := s.accountResource()
	result, err := model.
		Fields(projectionFields(def)...).
		Where(dao.Account.Columns().Id, accountID).
		One()
	if err != nil {
		return nil, err
	}
	if result.IsEmpty() {
		return nil, bizerrResourceNotFound()
	}
	record := projectRecord(result, def)
	delete(record, "passwordHash")
	return record, nil
}

func (s *serviceImpl) accountDetailAuditRecord(ctx context.Context, accountID int64) (Record, error) {
	return s.accountDetailAuditRecordByModel(ctx, accountID, s.tenantFilter.Apply(ctx, dao.AccountDetail.Ctx(ctx), ""))
}

func (s *serviceImpl) accountDetailAuditRecordForTenant(ctx context.Context, tenantID int, accountID int64) (Record, error) {
	return s.accountDetailAuditRecordByModel(ctx, accountID, dao.AccountDetail.Ctx(ctx).Where(dao.AccountDetail.Columns().TenantId, tenantID))
}

func (s *serviceImpl) accountDetailAuditRecordByModel(ctx context.Context, accountID int64, model *gdb.Model) (Record, error) {
	def := s.accountDetailResource()
	result, err := model.
		Fields(projectionFields(def)...).
		Where(dao.AccountDetail.Columns().AccountId, accountID).
		One()
	if err != nil {
		return nil, err
	}
	if result.IsEmpty() {
		return nil, bizerrResourceNotFound()
	}
	return projectRecord(result, def), nil
}

func (s *serviceImpl) accountDetailRecordsByWechat(ctx context.Context, unionID string, excludedAccountID int64) ([]Record, error) {
	def := s.accountDetailResource()
	model := s.tenantFilter.Apply(ctx, dao.AccountDetail.Ctx(ctx), "").
		Fields(projectionFields(def)...).
		Where(dao.AccountDetail.Columns().Wechat, unionID)
	if excludedAccountID > 0 {
		model = model.WhereNot(dao.AccountDetail.Columns().AccountId, excludedAccountID)
	}
	result, err := model.All()
	if err != nil {
		return nil, err
	}
	return projectResult(result, def), nil
}

func (s *serviceImpl) insertAccountDetailUpdateAudits(ctx context.Context, oldRecords []Record) error {
	return s.insertAccountDetailUpdateAuditsForTenant(ctx, s.tenantID(ctx), oldRecords)
}

func (s *serviceImpl) insertAccountDetailUpdateAuditsForTenant(ctx context.Context, tenantID int, oldRecords []Record) error {
	for _, oldRecord := range oldRecords {
		accountID := recordAccountID(oldRecord)
		newRecord, err := s.accountDetailAuditRecordForTenant(ctx, tenantID, accountID)
		if err != nil {
			return err
		}
		if err := s.insertAccountAuditForTenant(ctx, tenantID, accountID, accountAuditTableDetail, accountAuditActionUpdate, oldRecord, newRecord); err != nil {
			return err
		}
	}
	return nil
}

func recordAccountID(record Record) int64 {
	if record == nil {
		return 0
	}
	if value, ok := record["accountId"]; ok {
		return gconv.Int64(value)
	}
	if value, ok := record["id"]; ok {
		return gconv.Int64(value)
	}
	return 0
}

func marshalAuditRecord(record Record) string {
	if record == nil {
		return ""
	}
	content, err := json.Marshal(record)
	if err != nil {
		return ""
	}
	return string(content)
}

func bizerrResourceNotFound() error {
	return bizerr.NewCode(CodeResourceNotFound)
}
