// This file implements logged-in Wechat rebind state handling for the runtime
// user self-service API without requiring a direct external Wechat dependency.

package uidentity

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const (
	ticketKindWechatRebind       = "wechat_rebind"
	ticketCodePrefixWechatRebind = "wechat_rebind:"

	wechatRebindStatePrefix       = "rebindWechat"
	wechatRebindStatusPending     = "pending"
	wechatRebindStatusSuccess     = "success"
	wechatRebindStatusUnsupported = "unsupported"
	wechatRebindStatusFailed      = "failed"

	configKeyWechatRebindAuthorizeURL = "runtime.wechatRebindAuthorizeUrl"
	configKeyWechatRebindRedirectURL  = "runtime.wechatRebindRedirectUrl"
)

type wechatRebindStateData struct {
	Kind        string `json:"kind"`
	State       string `json:"state"`
	TenantID    int    `json:"tenantId"`
	Number      string `json:"number"`
	AccountID   int64  `json:"accountId"`
	Callback    string `json:"callback,omitempty"`
	Status      string `json:"status"`
	UnionID     string `json:"unionId,omitempty"`
	Code        string `json:"code,omitempty"`
	RedirectURL string `json:"redirectUrl,omitempty"`
	ErrorCode   string `json:"errorCode,omitempty"`
	Message     string `json:"message,omitempty"`
	ExpiredAt   int64  `json:"expiredAt,omitempty"`
}

// CreateRuntimeWechatRebindState creates one pending logged-in rebind state.
func (s *serviceImpl) CreateRuntimeWechatRebindState(ctx context.Context, in WechatRebindStateInput) (*WechatRebindStateOutput, error) {
	account, err := s.getAccountByNumber(ctx, in.Number)
	if err != nil {
		return nil, err
	}
	state, err := randomToken(wechatRebindStatePrefix)
	if err != nil {
		return nil, err
	}
	tenantID := s.tenantID(ctx)
	expiredAt := time.Now().Add(activationTTL)
	millis := expiredAt.UnixMilli()
	payload := wechatRebindStateData{
		Kind:      ticketKindWechatRebind,
		State:     state,
		TenantID:  tenantID,
		Number:    account.Number,
		AccountID: account.Id,
		Callback:  strings.TrimSpace(in.Callback),
		Status:    wechatRebindStatusPending,
		ExpiredAt: millis,
	}
	if err := s.createRuntimeToken(ctx, do.OauthToken{
		Code:      ticketCodePrefixWechatRebind + state,
		ExpiredAt: &expiredAt,
	}, payload); err != nil {
		return nil, err
	}
	baseURL, err := s.configSvc.String(ctx, configKeyWechatRebindAuthorizeURL, "")
	if err != nil {
		return nil, err
	}
	return &WechatRebindStateOutput{
		State:     state,
		Status:    payload.Status,
		URL:       wechatRebindAuthorizeURL(baseURL, state, payload.Callback),
		ExpiredAt: &millis,
	}, nil
}

// CompleteRuntimeWechatRebind records one rebind callback result.
func (s *serviceImpl) CompleteRuntimeWechatRebind(ctx context.Context, in WechatRebindCallbackInput) (*WechatRebindStateOutput, error) {
	token, payload, err := s.wechatRebindStateUnscoped(ctx, in.State)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Callback) != "" {
		payload.Callback = strings.TrimSpace(in.Callback)
	}
	payload.Code = strings.TrimSpace(in.Code)
	unionID := strings.TrimSpace(in.UnionID)
	if unionID == "" {
		payload.Status = wechatRebindStatusUnsupported
		payload.ErrorCode = CodeUnsupportedExternalFlow.RuntimeCode()
		payload.Message = CodeUnsupportedExternalFlow.Fallback()
		payload.RedirectURL = s.wechatRebindRedirectURL(ctx, payload)
		if err := s.updateRuntimePayloadForTenant(ctx, token.TenantId, token.Id, payload); err != nil {
			return nil, err
		}
		return wechatRebindResult(payload), nil
	}
	if err := s.bindUnionIDToAccountForTenant(ctx, token.TenantId, payload.AccountID, unionID); err != nil {
		payload.Status = wechatRebindStatusFailed
		if meta, ok := bizerr.As(err); ok {
			payload.ErrorCode = meta.RuntimeCode()
			payload.Message = meta.Fallback()
		} else {
			payload.Message = err.Error()
		}
		payload.RedirectURL = s.wechatRebindRedirectURL(ctx, payload)
		if updateErr := s.updateRuntimePayloadForTenant(ctx, token.TenantId, token.Id, payload); updateErr != nil {
			return nil, updateErr
		}
		return nil, err
	}
	payload.UnionID = unionID
	payload.Status = wechatRebindStatusSuccess
	payload.RedirectURL = s.wechatRebindRedirectURL(ctx, payload)
	account, err := s.getAccountByIDForTenant(ctx, token.TenantId, payload.AccountID)
	if err != nil {
		return nil, err
	}
	if err := s.recordAccountActiveLogForTenant(ctx, token.TenantId, account, unionID, accountActiveLogTypeActivation); err != nil {
		return nil, err
	}
	if err := s.updateRuntimePayloadForTenant(ctx, token.TenantId, token.Id, payload); err != nil {
		return nil, err
	}
	return wechatRebindResult(payload), nil
}

// GetRuntimeWechatRebindState returns one rebind state for the matched account.
func (s *serviceImpl) GetRuntimeWechatRebindState(ctx context.Context, in WechatRebindStateLookupInput) (*WechatRebindStateOutput, error) {
	_, payload, err := s.wechatRebindState(ctx, in.State)
	if err != nil {
		return nil, err
	}
	if payload.Number != strings.TrimSpace(in.Number) {
		return nil, bizerr.NewCode(CodeWechatRebindInvalid)
	}
	return wechatRebindResult(payload), nil
}

func (s *serviceImpl) wechatRebindState(ctx context.Context, state string) (*entity.OauthToken, *wechatRebindStateData, error) {
	model := s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "")
	return s.wechatRebindStateByModel(ctx, model, state)
}

func (s *serviceImpl) wechatRebindStateUnscoped(ctx context.Context, state string) (*entity.OauthToken, *wechatRebindStateData, error) {
	return s.wechatRebindStateByModel(ctx, dao.OauthToken.Ctx(ctx), state)
}

func (s *serviceImpl) wechatRebindStateByModel(ctx context.Context, model *gdb.Model, state string) (*entity.OauthToken, *wechatRebindStateData, error) {
	trimmed := strings.TrimSpace(state)
	if trimmed == "" {
		return nil, nil, bizerr.NewCode(CodeWechatRebindInvalid)
	}
	var token *entity.OauthToken
	err := model.
		Where(dao.OauthToken.Columns().Code, ticketCodePrefixWechatRebind+trimmed).
		Scan(&token)
	if err != nil {
		return nil, nil, err
	}
	if token == nil || token.ExpiredAt == nil || token.ExpiredAt.Before(time.Now()) {
		return nil, nil, bizerr.NewCode(CodeWechatRebindInvalid)
	}
	payload := &wechatRebindStateData{}
	if err := json.Unmarshal([]byte(token.Data), payload); err != nil {
		return nil, nil, bizerr.NewCode(CodeWechatRebindInvalid)
	}
	if payload.Kind != ticketKindWechatRebind || payload.State != trimmed || payload.TenantID != token.TenantId {
		return nil, nil, bizerr.NewCode(CodeWechatRebindInvalid)
	}
	payload.ExpiredAt = token.ExpiredAt.UnixMilli()
	return token, payload, nil
}

func (s *serviceImpl) updateRuntimePayloadForTenant(ctx context.Context, tenantID int, tokenID int64, payload any) error {
	content, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = dao.OauthToken.Ctx(ctx).
		Where(dao.OauthToken.Columns().TenantId, tenantID).
		Where(dao.OauthToken.Columns().Id, tokenID).
		Data(do.OauthToken{Data: string(content), UpdatedBy: s.actorID(ctx)}).
		Update()
	return err
}

func (s *serviceImpl) bindUnionIDToAccountForTenant(ctx context.Context, tenantID int, accountID int64, unionID string) error {
	trimmed := strings.TrimSpace(unionID)
	if trimmed == "" {
		return bizerr.NewCode(CodeUnionIDChallengeInvalid)
	}
	count, err := dao.AccountDetail.Ctx(ctx).
		Where(dao.AccountDetail.Columns().TenantId, tenantID).
		Where(dao.AccountDetail.Columns().Wechat, trimmed).
		WhereNot(dao.AccountDetail.Columns().AccountId, accountID).
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewCode(CodeContactConflict)
	}
	existing, err := dao.AccountDetail.Ctx(ctx).
		Where(dao.AccountDetail.Columns().TenantId, tenantID).
		Where(dao.AccountDetail.Columns().AccountId, accountID).
		Count()
	if err != nil {
		return err
	}
	if existing == 0 {
		_, err = dao.AccountDetail.Ctx(ctx).Data(do.AccountDetail{
			TenantId:  tenantID,
			AccountId: accountID,
			Wechat:    trimmed,
			CreatedBy: s.actorID(ctx),
			UpdatedBy: s.actorID(ctx),
		}).Insert()
		return err
	}
	_, err = dao.AccountDetail.Ctx(ctx).
		Where(dao.AccountDetail.Columns().TenantId, tenantID).
		Where(dao.AccountDetail.Columns().AccountId, accountID).
		Data(do.AccountDetail{Wechat: trimmed, UpdatedBy: s.actorID(ctx)}).
		Update()
	return err
}

func (s *serviceImpl) wechatRebindRedirectURL(ctx context.Context, payload *wechatRebindStateData) string {
	baseURL, err := s.configSvc.String(ctx, configKeyWechatRebindRedirectURL, "")
	if err != nil || payload == nil || strings.TrimSpace(baseURL) == "" {
		return ""
	}
	result := callbackWithQuery(baseURL, "state", payload.State)
	result = callbackWithQuery(result, "status", payload.Status)
	callback := payload.Callback
	if strings.TrimSpace(callback) == "" {
		callback = "rebind"
	}
	result = callbackWithQuery(result, "cascallback", callback)
	if payload.Message != "" {
		result = callbackWithQuery(result, "msg", payload.Message)
	}
	return result
}

func wechatRebindAuthorizeURL(baseURL string, state string, callback string) string {
	result := callbackWithQuery(baseURL, "state", state)
	if strings.TrimSpace(callback) != "" {
		result = callbackWithQuery(result, "cascallback", callback)
	}
	return result
}

func wechatRebindResult(payload *wechatRebindStateData) *WechatRebindStateOutput {
	if payload == nil {
		return nil
	}
	millis := payload.ExpiredAt
	return &WechatRebindStateOutput{
		State:       payload.State,
		Status:      payload.Status,
		Success:     payload.Status == wechatRebindStatusSuccess,
		RedirectURL: payload.RedirectURL,
		ExpiredAt:   &millis,
		ErrorCode:   payload.ErrorCode,
		Message:     payload.Message,
	}
}
