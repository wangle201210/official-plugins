// This file contains typed DO builders for application, relation, blacklist,
// log, token, password-rule, and SMS resources.

package uidentity

import (
	"context"

	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
)

func (s *serviceImpl) applicationData(ctx context.Context, body map[string]any, create bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.Applications{UpdateBy: actorID}
	if create {
		data.CreateBy = actorID
	}
	copyStringFields(body, map[string]*any{
		"name": &data.Name, "alias": &data.Alias, "clientId": &data.ClientId, "secretKey": &data.SecretKey,
		"accessModel": &data.AccessModel, "callbackUrl": &data.CallbackUrl, "whitelist": &data.Whitelist,
	})
	copyIntFields(body, map[string]*any{"status": &data.Status})
	return data, nil
}

func (s *serviceImpl) accountGroupData(ctx context.Context, body map[string]any, create bool) (any, error) {
	data := do.AccountGroup{}
	copyInt64Fields(body, map[string]*any{"accountId": &data.AccountId, "groupId": &data.GroupsId, "groupsId": &data.GroupsId})
	return data, nil
}

func (s *serviceImpl) accountUnitData(ctx context.Context, body map[string]any, create bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.AccountUnit{}
	if create {
		data.CreateBy = actorID
	}
	copyInt64Fields(body, map[string]*any{"accountId": &data.AccountId, "unitId": &data.UnitsId, "unitsId": &data.UnitsId})
	return data, nil
}

func (s *serviceImpl) accountAppRoleData(ctx context.Context, body map[string]any, create bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.AccountAppRole{UpdateBy: actorID}
	if create {
		data.CreateBy = actorID
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
	actorID := s.actorID(ctx)
	data := do.AccountAppBlacklist{UpdateBy: actorID}
	if create {
		data.CreateBy = actorID
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
	actorID := s.actorID(ctx)
	data := do.GroupAppBlacklist{UpdateBy: actorID}
	if create {
		data.CreateBy = actorID
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
	actorID := s.actorID(ctx)
	data := do.PassRuler{UpdateBy: actorID}
	if create {
		data.CreateBy = actorID
	}
	copyStringFields(body, map[string]*any{"name": &data.Name})
	copyIntFields(body, map[string]*any{
		"capital": &data.Capital, "lower": &data.Lower, "number": &data.Number, "symbol": &data.Symbol,
		"length": &data.Length, "interval": &data.Interval, "intervalDays": &data.Interval, "intervalStatus": &data.IntervalStatus, "status": &data.Status,
	})
	return data, nil
}

func (s *serviceImpl) smsData(ctx context.Context, body map[string]any, create bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.Sms{UpdateBy: actorID}
	if create {
		data.CreateBy = actorID
	}
	copyStringFields(body, map[string]*any{"phone": &data.Phone, "type": &data.Type, "content": &data.Content, "respMsg": &data.RespMsg})
	copyIntFields(body, map[string]*any{"status": &data.Status})
	return data, nil
}

func (s *serviceImpl) casLoginLogData(ctx context.Context, body map[string]any, create bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.CasLoginLog{UpdateBy: actorID}
	if create {
		data.CreateBy = actorID
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
	actorID := s.actorID(ctx)
	data := do.OauthLog{UpdateBy: actorID}
	if create {
		data.CreateBy = actorID
	}
	copyInt64Fields(body, map[string]*any{"userId": &data.UserId, "appId": &data.AppId})
	copyStringFields(body, map[string]*any{"redirectUri": &data.RedirectUri, "scope": &data.Scope})
	return data, nil
}

func (s *serviceImpl) oauthTokenData(ctx context.Context, body map[string]any, create bool) (any, error) {
	data := do.Oauth2Token{}
	copyStringFields(body, map[string]*any{"code": &data.Code, "access": &data.Access, "refresh": &data.Refresh, "data": &data.Data})
	if hasField(body, "expiredAt") {
		data.ExpiredAt = int64Field(body, "expiredAt")
	}
	return data, nil
}

func (s *serviceImpl) accountChangeLogData(ctx context.Context, body map[string]any, create bool) (any, error) {
	actorID := s.actorID(ctx)
	data := do.AccountChangeLog{UpdateBy: actorID}
	if create {
		data.CreateBy = actorID
	}
	copyInt64Fields(body, map[string]*any{"accountId": &data.AccountId})
	copyStringFields(body, map[string]*any{
		"tableName": &data.TableName, "action": &data.Action, "dataOld": &data.DataOld, "dataNew": &data.DataNew,
		"errMsg": &data.ErrMsg, "errNumber": &data.ErrNumber,
	})
	return data, nil
}

func (s *serviceImpl) accountActiveLogData(ctx context.Context, body map[string]any, create bool) (any, error) {
	data := do.AccountActiveLog{}
	copyStringFields(body, map[string]*any{"number": &data.Number, "phone": &data.Phone, "wechat": &data.Wechat})
	copyIntFields(body, map[string]*any{"type": &data.Type})
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
