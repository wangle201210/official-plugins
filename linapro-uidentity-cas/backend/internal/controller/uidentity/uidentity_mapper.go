// This file maps service-layer generic records and statistics into API DTOs.

package uidentity

import (
	v1 "lina-plugin-linapro-uidentity-cas/backend/api/uidentity/v1"
	uidentitysvc "lina-plugin-linapro-uidentity-cas/backend/internal/service/uidentity"
)

func toAPIRecords(records []uidentitysvc.Record) []v1.ResourceRecord {
	result := make([]v1.ResourceRecord, 0, len(records))
	for _, record := range records {
		result = append(result, v1.ResourceRecord(record))
	}
	return result
}

func toAPIStatItems(items []*uidentitysvc.StatItem) []*v1.StatItem {
	result := make([]*v1.StatItem, 0, len(items))
	for _, item := range items {
		result = append(result, &v1.StatItem{Name: item.Name, Total: item.Total})
	}
	return result
}

func toAPIRuntimeApplication(app *uidentitysvc.RuntimeApplication) *v1.RuntimeApplication {
	if app == nil {
		return nil
	}
	return &v1.RuntimeApplication{
		Id:          app.ID,
		Name:        app.Name,
		Alias:       app.Alias,
		ClientId:    app.ClientID,
		AccessModel: app.AccessModel,
		CallbackUrl: app.CallbackURL,
	}
}

func toAPIRuntimeApplications(apps []*uidentitysvc.RuntimeApplication) []*v1.RuntimeApplication {
	result := make([]*v1.RuntimeApplication, 0, len(apps))
	for _, app := range apps {
		result = append(result, toAPIRuntimeApplication(app))
	}
	return result
}

func toAPIRuntimeAccount(account *uidentitysvc.RuntimeAccount) *v1.RuntimeAccount {
	if account == nil {
		return nil
	}
	var detail *v1.RuntimeAccountDetail
	if account.Detail != nil {
		detail = &v1.RuntimeAccountDetail{
			Birthday: account.Detail.Birthday,
			Email:    account.Detail.Email,
			Gender:   account.Detail.Gender,
			Qq:       account.Detail.QQ,
			Wechat:   account.Detail.Wechat,
			Idcard:   account.Detail.Idcard,
			Avatar:   account.Detail.Avatar,
			Face:     account.Detail.Face,
		}
	}
	return &v1.RuntimeAccount{
		Id:            account.ID,
		Number:        account.Number,
		Name:          account.Name,
		Phone:         account.Phone,
		Status:        account.Status,
		PassLevel:     account.PassLevel,
		ContainerId:   account.ContainerID,
		ContainerName: account.ContainerName,
		UnitId:        account.UnitID,
		UnitName:      account.UnitName,
		ExpireAt:      account.ExpireAt,
		Groups:        account.Groups,
		Detail:        detail,
	}
}

func toAPIRuntimeAccounts(accounts []*uidentitysvc.RuntimeAccount) []*v1.RuntimeAccount {
	result := make([]*v1.RuntimeAccount, 0, len(accounts))
	for _, account := range accounts {
		result = append(result, toAPIRuntimeAccount(account))
	}
	return result
}

func toAPIRuntimeLogin(out *uidentitysvc.RuntimeLoginOutput) *v1.CasRuntimeLoginRes {
	if out == nil {
		return nil
	}
	return &v1.CasRuntimeLoginRes{
		CallbackUrl: out.CallbackURL,
		Tgt:         out.TGT,
		St:          out.ST,
		User:        toAPIRuntimeAccount(out.User),
		Users:       toAPIRuntimeAccounts(out.Users),
	}
}

func toAPIWechatLoginResult(out *uidentitysvc.WechatLoginQRResultOutput) *v1.WechatLoginQRResultRes {
	if out == nil {
		return nil
	}
	return &v1.WechatLoginQRResultRes{
		State:       out.State,
		Status:      out.Status,
		RedirectUrl: out.RedirectURL,
		ChallengeId: out.ChallengeID,
		CallbackUrl: out.CallbackURL,
		ErrorCode:   out.ErrorCode,
		Message:     out.Message,
		Login:       toAPIRuntimeLogin(out.Login),
	}
}

func toAPIWechatRebindState(out *uidentitysvc.WechatRebindStateOutput) *v1.UserWechatRebindStateRes {
	if out == nil {
		return nil
	}
	return &v1.UserWechatRebindStateRes{
		State:       out.State,
		Status:      out.Status,
		Success:     out.Success,
		RedirectUrl: out.RedirectURL,
		ErrorCode:   out.ErrorCode,
		Message:     out.Message,
		ExpiredAt:   out.ExpiredAt,
	}
}

func toAPIActivationWechatState(out *uidentitysvc.ActivationWechatStateOutput) *v1.ActivationWechatCallbackRes {
	if out == nil {
		return nil
	}
	return &v1.ActivationWechatCallbackRes{
		State:       out.State,
		Status:      out.Status,
		Success:     out.Success,
		RedirectUrl: out.RedirectURL,
		ErrorCode:   out.ErrorCode,
		Message:     out.Message,
	}
}
