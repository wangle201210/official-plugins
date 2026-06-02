// This file contains typed DO builders for application, relation, blacklist,
// log, token, password-rule, and SMS resources.

package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
)

func (s *serviceImpl) applicationData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.Application{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyStringFields(body, map[string]*any{
		"name": &data.Name, "alias": &data.Alias, "clientId": &data.ClientId, "secretKey": &data.SecretKey,
		"accessModel": &data.AccessModel, "callbackUrl": &data.CallbackUrl, "whitelist": &data.Whitelist,
	})
	copyIntFields(body, map[string]*any{"status": &data.Status})
	return data, nil
}

func (s *serviceImpl) accountGroupData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.AccountGroup{}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyInt64Fields(body, map[string]*any{"accountId": &data.AccountId, "groupId": &data.GroupId})
	return data, nil
}

func (s *serviceImpl) accountUnitData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.AccountUnit{}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyInt64Fields(body, map[string]*any{"accountId": &data.AccountId, "unitId": &data.UnitId})
	return data, nil
}

func (s *serviceImpl) accountAppRoleData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.AccountAppRole{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyInt64Fields(body, map[string]*any{
		"giveAccountId": &data.GiveAccountId, "empoweredAccountId": &data.EmpoweredAccountId, "appId": &data.AppId,
	})
	if value := timeField(body, "expireAt"); value != nil {
		data.ExpireAt = value
	}
	return data, nil
}

func (s *serviceImpl) accountAppBlacklistData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.AccountAppBlacklist{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyStringFields(body, map[string]*any{"name": &data.Name})
	copyInt64Fields(body, map[string]*any{"appId": &data.AppId, "accountId": &data.AccountId})
	if value := timeField(body, "effectAt"); value != nil {
		data.EffectAt = value
	}
	if value := timeField(body, "expireAt"); value != nil {
		data.ExpireAt = value
	}
	return data, nil
}

func (s *serviceImpl) groupAppBlacklistData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.GroupAppBlacklist{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyStringFields(body, map[string]*any{"name": &data.Name})
	copyInt64Fields(body, map[string]*any{"appId": &data.AppId, "groupId": &data.GroupId})
	if value := timeField(body, "effectAt"); value != nil {
		data.EffectAt = value
	}
	if value := timeField(body, "expireAt"); value != nil {
		data.ExpireAt = value
	}
	return data, nil
}

func (s *serviceImpl) passRuleData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.PassRule{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyStringFields(body, map[string]*any{"name": &data.Name})
	copyIntFields(body, map[string]*any{
		"capital": &data.Capital, "lower": &data.Lower, "number": &data.Number, "symbol": &data.Symbol,
		"length": &data.Length, "intervalDays": &data.IntervalDays, "intervalStatus": &data.IntervalStatus, "status": &data.Status,
	})
	return data, nil
}

func (s *serviceImpl) smsData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.Sms{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyStringFields(body, map[string]*any{"phone": &data.Phone, "type": &data.Type, "content": &data.Content, "respMsg": &data.RespMsg})
	copyIntFields(body, map[string]*any{"status": &data.Status})
	return data, nil
}

func (s *serviceImpl) casLoginLogData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.CasLoginLog{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyStringFields(body, map[string]*any{
		"ipaddr": &data.Ipaddr, "loginLocation": &data.LoginLocation, "browser": &data.Browser, "os": &data.Os,
		"platform": &data.Platform, "remark": &data.Remark, "msg": &data.Msg, "loginType": &data.LoginType,
	})
	copyInt64Fields(body, map[string]*any{"accountId": &data.AccountId, "choiceAccountId": &data.ChoiceAccountId, "appId": &data.AppId})
	if value := timeField(body, "loginTime"); value != nil {
		data.LoginTime = value
	}
	return data, nil
}

func (s *serviceImpl) oauthLogData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.OauthLog{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyInt64Fields(body, map[string]*any{"userId": &data.UserId, "appId": &data.AppId})
	copyStringFields(body, map[string]*any{"redirectUri": &data.RedirectUri, "scope": &data.Scope})
	return data, nil
}

func (s *serviceImpl) oauthTokenData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.OauthToken{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyStringFields(body, map[string]*any{"code": &data.Code, "access": &data.Access, "refresh": &data.Refresh, "data": &data.Data})
	if value := timeField(body, "expiredAt"); value != nil {
		data.ExpiredAt = value
	}
	return data, nil
}

func (s *serviceImpl) accountChangeLogData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.AccountChangeLog{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyInt64Fields(body, map[string]*any{"accountId": &data.AccountId})
	copyStringFields(body, map[string]*any{
		"tableName": &data.TableName, "action": &data.Action, "dataOld": &data.DataOld, "dataNew": &data.DataNew,
		"errMsg": &data.ErrMsg, "errNumber": &data.ErrNumber,
	})
	return data, nil
}

func (s *serviceImpl) accountActiveLogData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.AccountActiveLog{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyStringFields(body, map[string]*any{"number": &data.Number, "phone": &data.Phone, "wechat": &data.Wechat})
	copyIntFields(body, map[string]*any{"type": &data.Type})
	return data, nil
}

func (s *serviceImpl) sysJobData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.SysJob{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyStringFields(body, map[string]*any{
		"jobName": &data.JobName, "jobGroup": &data.JobGroup, "cronExpression": &data.CronExpression,
		"invokeTarget": &data.InvokeTarget, "args": &data.Args,
	})
	copyIntFields(body, map[string]*any{
		"jobType": &data.JobType, "misfirePolicy": &data.MisfirePolicy, "concurrent": &data.Concurrent, "status": &data.Status,
	})
	copyInt64Fields(body, map[string]*any{"entryId": &data.EntryId})
	return data, nil
}

func (s *serviceImpl) jobLogData(ctx context.Context, body map[string]any, create bool) (any, error) {
	tenantID, actorID := s.baseOwnedDO(ctx, create)
	data := do.JobLog{UpdatedBy: actorID}
	if create {
		data.TenantId = tenantID
		data.CreatedBy = actorID
	}
	copyStringFields(body, map[string]*any{"jobName": &data.JobName})
	copyInt64Fields(body, map[string]*any{
		"jobId": &data.JobId, "createNum": &data.CreateNum, "updateNum": &data.UpdateNum,
		"deleteNum": &data.DeleteNum, "errNum": &data.ErrNum,
	})
	if value := timeField(body, "startAt"); value != nil {
		data.StartAt = value
	}
	if value := timeField(body, "endAt"); value != nil {
		data.EndAt = value
	}
	return data, nil
}

func copyStringFields(body map[string]any, fields map[string]*any) {
	for field, target := range fields {
		if hasField(body, field) {
			*target = stringField(body, field)
		}
	}
}

func copyIntFields(body map[string]any, fields map[string]*any) {
	for field, target := range fields {
		if hasField(body, field) {
			*target = intField(body, field)
		}
	}
}

func copyInt64Fields(body map[string]any, fields map[string]*any) {
	for field, target := range fields {
		if hasField(body, field) {
			*target = int64Field(body, field)
		}
	}
}
