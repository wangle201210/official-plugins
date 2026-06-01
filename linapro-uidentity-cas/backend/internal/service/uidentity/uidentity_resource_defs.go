// This file defines non-account resource metadata and typed DO builders for
// all plugin-owned UIdentity CAS tables.

package uidentity

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
)

func (s *serviceImpl) applicationResource() *resourceDefinition {
	cols := dao.Application.Columns()
	return &resourceDefinition{
		name:          "applications",
		table:         dao.Application.Table(),
		idColumn:      cols.Id,
		defaultOrder:  cols.Id,
		keywordFields: []string{cols.Name, cols.Alias, cols.ClientId},
		apiToColumn: map[string]string{
			"id": cols.Id, "tenantId": cols.TenantId, "name": cols.Name, "alias": cols.Alias, "clientId": cols.ClientId,
			"secretKey": cols.SecretKey, "accessModel": cols.AccessModel, "status": cols.Status, "callbackUrl": cols.CallbackUrl,
			"whitelist": cols.Whitelist, "createdBy": cols.CreatedBy, "updatedBy": cols.UpdatedBy,
			"createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
		},
		timeFields: commonTimeFields(),
		model:      func(ctx context.Context) *gdb.Model { return dao.Application.Ctx(ctx) },
		data:       s.applicationData,
	}
}

func (s *serviceImpl) accountGroupResource() *resourceDefinition {
	cols := dao.AccountGroup.Columns()
	return relationResource("account-groups", dao.AccountGroup.Table(), cols.Id, map[string]string{
		"id": cols.Id, "tenantId": cols.TenantId, "accountId": cols.AccountId, "groupId": cols.GroupId,
		"createdBy": cols.CreatedBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt,
	}, func(ctx context.Context) *gdb.Model { return dao.AccountGroup.Ctx(ctx) }, s.accountGroupData)
}

func (s *serviceImpl) accountUnitResource() *resourceDefinition {
	cols := dao.AccountUnit.Columns()
	return relationResource("account-units", dao.AccountUnit.Table(), cols.Id, map[string]string{
		"id": cols.Id, "tenantId": cols.TenantId, "accountId": cols.AccountId, "unitId": cols.UnitId,
		"createdBy": cols.CreatedBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt,
	}, func(ctx context.Context) *gdb.Model { return dao.AccountUnit.Ctx(ctx) }, s.accountUnitData)
}

func (s *serviceImpl) accountAppRoleResource() *resourceDefinition {
	cols := dao.AccountAppRole.Columns()
	return relationResource("account-app-roles", dao.AccountAppRole.Table(), cols.Id, map[string]string{
		"id": cols.Id, "tenantId": cols.TenantId, "giveAccountId": cols.GiveAccountId, "empoweredAccountId": cols.EmpoweredAccountId,
		"appId": cols.AppId, "expireAt": cols.ExpireAt, "createdBy": cols.CreatedBy, "updatedBy": cols.UpdatedBy,
		"createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, func(ctx context.Context) *gdb.Model { return dao.AccountAppRole.Ctx(ctx) }, s.accountAppRoleData)
}

func (s *serviceImpl) accountAppBlacklistResource() *resourceDefinition {
	cols := dao.AccountAppBlacklist.Columns()
	return blacklistResource("account-app-blacklists", dao.AccountAppBlacklist.Table(), cols.Id, cols.Name, cols.AppId, cols.AccountId, "", cols.EffectAt, cols.ExpireAt, cols.CreatedBy, cols.UpdatedBy, cols.CreatedAt, cols.UpdatedAt, cols.DeletedAt, func(ctx context.Context) *gdb.Model { return dao.AccountAppBlacklist.Ctx(ctx) }, s.accountAppBlacklistData)
}

func (s *serviceImpl) groupAppBlacklistResource() *resourceDefinition {
	cols := dao.GroupAppBlacklist.Columns()
	return blacklistResource("group-app-blacklists", dao.GroupAppBlacklist.Table(), cols.Id, cols.Name, cols.AppId, "", cols.GroupId, cols.EffectAt, cols.ExpireAt, cols.CreatedBy, cols.UpdatedBy, cols.CreatedAt, cols.UpdatedAt, cols.DeletedAt, func(ctx context.Context) *gdb.Model { return dao.GroupAppBlacklist.Ctx(ctx) }, s.groupAppBlacklistData)
}

func (s *serviceImpl) passRuleResource() *resourceDefinition {
	cols := dao.PassRule.Columns()
	return &resourceDefinition{
		name:          "pass-rules",
		table:         dao.PassRule.Table(),
		idColumn:      cols.Id,
		defaultOrder:  cols.Id,
		keywordFields: []string{cols.Name},
		apiToColumn: map[string]string{
			"id": cols.Id, "tenantId": cols.TenantId, "name": cols.Name, "capital": cols.Capital, "lower": cols.Lower,
			"number": cols.Number, "symbol": cols.Symbol, "length": cols.Length, "intervalDays": cols.IntervalDays,
			"intervalStatus": cols.IntervalStatus, "status": cols.Status, "createdBy": cols.CreatedBy, "updatedBy": cols.UpdatedBy,
			"createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
		},
		timeFields: commonTimeFields(),
		model:      func(ctx context.Context) *gdb.Model { return dao.PassRule.Ctx(ctx) },
		data:       s.passRuleData,
	}
}

func (s *serviceImpl) smsResource() *resourceDefinition {
	cols := dao.Sms.Columns()
	return &resourceDefinition{
		name:          "sms-records",
		table:         dao.Sms.Table(),
		idColumn:      cols.Id,
		defaultOrder:  cols.Id,
		keywordFields: []string{cols.Phone, cols.Type, cols.Content},
		apiToColumn: map[string]string{
			"id": cols.Id, "tenantId": cols.TenantId, "phone": cols.Phone, "type": cols.Type, "content": cols.Content,
			"status": cols.Status, "respMsg": cols.RespMsg, "createdBy": cols.CreatedBy, "updatedBy": cols.UpdatedBy,
			"createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
		},
		timeFields: commonTimeFields(),
		model:      func(ctx context.Context) *gdb.Model { return dao.Sms.Ctx(ctx) },
		data:       s.smsData,
	}
}

func (s *serviceImpl) casLoginLogResource() *resourceDefinition {
	cols := dao.CasLoginLog.Columns()
	return relationResource("cas-login-logs", dao.CasLoginLog.Table(), cols.Id, map[string]string{
		"id": cols.Id, "tenantId": cols.TenantId, "accountId": cols.AccountId, "choiceAccountId": cols.ChoiceAccountId,
		"appId": cols.AppId, "ipaddr": cols.Ipaddr, "loginLocation": cols.LoginLocation, "browser": cols.Browser,
		"os": cols.Os, "platform": cols.Platform, "loginTime": cols.LoginTime, "remark": cols.Remark, "msg": cols.Msg,
		"loginType": cols.LoginType, "createdBy": cols.CreatedBy, "updatedBy": cols.UpdatedBy,
		"createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, func(ctx context.Context) *gdb.Model { return dao.CasLoginLog.Ctx(ctx) }, s.casLoginLogData)
}

func (s *serviceImpl) oauthLogResource() *resourceDefinition {
	cols := dao.OauthLog.Columns()
	return relationResource("oauth-logs", dao.OauthLog.Table(), cols.Id, map[string]string{
		"id": cols.Id, "tenantId": cols.TenantId, "userId": cols.UserId, "appId": cols.AppId, "redirectUri": cols.RedirectUri,
		"scope": cols.Scope, "createdBy": cols.CreatedBy, "updatedBy": cols.UpdatedBy,
		"createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, func(ctx context.Context) *gdb.Model { return dao.OauthLog.Ctx(ctx) }, s.oauthLogData)
}

func (s *serviceImpl) oauthTokenResource() *resourceDefinition {
	cols := dao.OauthToken.Columns()
	return relationResource("oauth-tokens", dao.OauthToken.Table(), cols.Id, map[string]string{
		"id": cols.Id, "tenantId": cols.TenantId, "expiredAt": cols.ExpiredAt, "code": cols.Code, "access": cols.Access,
		"refresh": cols.Refresh, "data": cols.Data, "createdBy": cols.CreatedBy, "updatedBy": cols.UpdatedBy,
		"createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt, "deletedAt": cols.DeletedAt,
	}, func(ctx context.Context) *gdb.Model { return dao.OauthToken.Ctx(ctx) }, s.oauthTokenData)
}

func (s *serviceImpl) accountChangeLogResource() *resourceDefinition {
	cols := dao.AccountChangeLog.Columns()
	return relationResource("account-change-logs", dao.AccountChangeLog.Table(), cols.Id, map[string]string{
		"id": cols.Id, "tenantId": cols.TenantId, "accountId": cols.AccountId, "tableName": cols.TableName, "action": cols.Action,
		"dataOld": cols.DataOld, "dataNew": cols.DataNew, "errMsg": cols.ErrMsg, "errNumber": cols.ErrNumber,
		"createdBy": cols.CreatedBy, "updatedBy": cols.UpdatedBy, "createdAt": cols.CreatedAt, "updatedAt": cols.UpdatedAt,
		"deletedAt": cols.DeletedAt,
	}, func(ctx context.Context) *gdb.Model { return dao.AccountChangeLog.Ctx(ctx) }, s.accountChangeLogData)
}

func relationResource(name, table, id string, apiToColumn map[string]string, model func(context.Context) *gdb.Model, data func(context.Context, map[string]any, bool) (any, error)) *resourceDefinition {
	return &resourceDefinition{
		name:          name,
		table:         table,
		idColumn:      id,
		defaultOrder:  id,
		keywordFields: relationKeywordFields(apiToColumn),
		apiToColumn:   apiToColumn,
		timeFields:    commonTimeFields(),
		model:         model,
		data:          data,
	}
}

func blacklistResource(name, table, id, nameCol, appIDCol, accountIDCol, groupIDCol, effectAtCol, expireAtCol, createdByCol, updatedByCol, createdAtCol, updatedAtCol, deletedAtCol string, model func(context.Context) *gdb.Model, data func(context.Context, map[string]any, bool) (any, error)) *resourceDefinition {
	apiToColumn := map[string]string{
		"id": id, "name": nameCol, "appId": appIDCol, "effectAt": effectAtCol, "expireAt": expireAtCol,
		"createdBy": createdByCol, "updatedBy": updatedByCol, "createdAt": createdAtCol, "updatedAt": updatedAtCol, "deletedAt": deletedAtCol,
	}
	if accountIDCol != "" {
		apiToColumn["accountId"] = accountIDCol
	}
	if groupIDCol != "" {
		apiToColumn["groupId"] = groupIDCol
	}
	return &resourceDefinition{
		name:          name,
		table:         table,
		idColumn:      id,
		defaultOrder:  id,
		keywordFields: []string{nameCol},
		apiToColumn:   apiToColumn,
		timeFields:    commonTimeFields(),
		model:         model,
		data:          data,
	}
}

func relationKeywordFields(apiToColumn map[string]string) []string {
	fields := make([]string, 0, 3)
	for _, apiName := range []string{"name", "type", "scope", "action", "errNumber", "phone"} {
		if column := apiToColumn[apiName]; column != "" {
			fields = append(fields, column)
		}
	}
	return fields
}

func (s *serviceImpl) accountData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.Account{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	if hasField(body, "number") {
		data.Number = stringField(body, "number")
	}
	if hasField(body, "name") {
		data.Name = stringField(body, "name")
	}
	if hasField(body, "phone") {
		data.Phone = stringField(body, "phone")
	}
	if hasField(body, "passwordHash") {
		data.PasswordHash = stringField(body, "passwordHash")
	}
	if value := timeField(body, "effectAt"); value != nil {
		data.EffectAt = value
	}
	if value := timeField(body, "expireAt"); value != nil {
		data.ExpireAt = value
	}
	if hasField(body, "passLevel") {
		data.PassLevel = intField(body, "passLevel")
	}
	if hasField(body, "containerId") {
		data.ContainerId = int64Field(body, "containerId")
	}
	if hasField(body, "unitId") {
		data.UnitId = int64Field(body, "unitId")
	}
	if hasField(body, "status") {
		data.Status = intField(body, "status")
	}
	return data, nil
}

func (s *serviceImpl) accountDetailData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.AccountDetail{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyStringFields(body, map[string]*any{
		"birthday": &data.Birthday, "email": &data.Email, "qq": &data.Qq, "wechat": &data.Wechat,
		"idcard": &data.Idcard, "avatar": &data.Avatar, "source": &data.Source, "grade": &data.Grade,
		"college": &data.College, "collegeCode": &data.CollegeCode, "campus": &data.Campus,
		"schoolSystem": &data.SchoolSystem, "graduatedAt": &data.GraduatedAt, "major": &data.Major,
		"className": &data.ClassName, "face": &data.Face,
	})
	if hasField(body, "accountId") {
		data.AccountId = int64Field(body, "accountId")
	}
	if hasField(body, "gender") {
		data.Gender = intField(body, "gender")
	}
	return data, nil
}

func (s *serviceImpl) groupData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.Group{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyStringFields(body, map[string]*any{"name": &data.Name, "alias": &data.Alias})
	return data, nil
}

func (s *serviceImpl) unitData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.Unit{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyStringFields(body, map[string]*any{"name": &data.Name, "alias": &data.Alias, "code": &data.Code})
	return data, nil
}

func (s *serviceImpl) containerData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.Container{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyStringFields(body, map[string]*any{"name": &data.Name, "alias": &data.Alias})
	copyIntFields(body, map[string]*any{"accountCount": &data.AccountCount, "adminCount": &data.AdminCount})
	return data, nil
}
