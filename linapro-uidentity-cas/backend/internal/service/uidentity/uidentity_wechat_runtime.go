// This file implements legacy Wechat QR login state handling without making a
// real external Wechat dependency mandatory for the plugin runtime.

package uidentity

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const (
	ticketKindWechatLogin       = "wechat_login"
	ticketCodePrefixWechatLogin = "wechat_login:"

	wechatLoginStatePrefix       = "loginByQr"
	wechatLoginStatusPending     = "pending"
	wechatLoginStatusSuccess     = "success"
	wechatLoginStatusBindNeeded  = "bind_required"
	wechatLoginStatusUnsupported = "unsupported"
	wechatLoginStatusFailed      = "failed"

	wechatLoginTTL = 5 * time.Minute

	configKeyWechatLoginAuthorizeURL = "runtime.wechatLoginAuthorizeUrl"
	configKeyWechatLoginRedirectURL  = "runtime.wechatLoginRedirectUrl"
)

type wechatLoginStateData struct {
	Kind        string              `json:"kind"`
	State       string              `json:"state"`
	ClientID    string              `json:"clientId"`
	Callback    string              `json:"callback"`
	Status      string              `json:"status"`
	UnionID     string              `json:"unionId,omitempty"`
	Code        string              `json:"code,omitempty"`
	RedirectURL string              `json:"redirectUrl,omitempty"`
	ChallengeID string              `json:"challengeId,omitempty"`
	CallbackURL string              `json:"callbackUrl,omitempty"`
	ErrorCode   string              `json:"errorCode,omitempty"`
	Message     string              `json:"message,omitempty"`
	Login       *RuntimeLoginOutput `json:"login,omitempty"`
}

// CreateWechatLoginQR creates one pending QR login state.
func (s *serviceImpl) CreateWechatLoginQR(ctx context.Context, in WechatLoginQRInput) (*WechatLoginQROutput, error) {
	app, err := s.runtimeApplicationByClientID(ctx, in.ClientID)
	if err != nil {
		return nil, err
	}
	state, err := randomToken(wechatLoginStatePrefix)
	if err != nil {
		return nil, err
	}
	expiredAt := time.Now().Add(wechatLoginTTL)
	payload := wechatLoginStateData{
		Kind:     ticketKindWechatLogin,
		State:    state,
		ClientID: app.ClientId,
		Callback: strings.TrimSpace(in.Callback),
		Status:   wechatLoginStatusPending,
	}
	if err := s.createRuntimeToken(ctx, do.OauthToken{
		Code:      ticketCodePrefixWechatLogin + state,
		ExpiredAt: &expiredAt,
	}, payload); err != nil {
		return nil, err
	}
	baseURL, err := s.configSvc.String(ctx, configKeyWechatLoginAuthorizeURL, "")
	if err != nil {
		return nil, err
	}
	millis := expiredAt.UnixMilli()
	return &WechatLoginQROutput{
		State:     state,
		URL:       wechatLoginAuthorizeURL(baseURL, app.ClientId, state, payload.Callback),
		ExpiredAt: &millis,
	}, nil
}

// CompleteWechatLoginQR records one callback result for later polling.
func (s *serviceImpl) CompleteWechatLoginQR(ctx context.Context, in WechatLoginCallbackInput) (*WechatLoginQRResultOutput, error) {
	token, payload, err := s.wechatLoginState(ctx, in.State)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.ClientID) != "" && strings.TrimSpace(in.ClientID) != payload.ClientID {
		return nil, bizerr.NewCode(CodeWechatLoginInvalid)
	}
	if strings.TrimSpace(in.Callback) != "" {
		payload.Callback = strings.TrimSpace(in.Callback)
	}
	payload.Code = strings.TrimSpace(in.Code)
	unionID := strings.TrimSpace(in.UnionID)
	if unionID == "" {
		payload.Status = wechatLoginStatusUnsupported
		payload.ErrorCode = CodeUnsupportedExternalFlow.RuntimeCode()
		payload.Message = CodeUnsupportedExternalFlow.Fallback()
		payload.RedirectURL = s.wechatLoginRedirectURL(ctx, payload)
		if err := s.updateRuntimePayload(ctx, token.Id, payload); err != nil {
			return nil, err
		}
		return wechatLoginResult(payload), nil
	}
	payload.UnionID = unionID
	detail, err := s.accountDetailByUnionID(ctx, unionID)
	if err != nil {
		if bizerr.Is(err, CodeResourceNotFound) {
			return s.completeUnboundWechatLogin(ctx, token, payload, unionID)
		}
		return nil, err
	}
	account, err := s.getAccountByID(ctx, detail.AccountId)
	if err != nil {
		return nil, err
	}
	app, err := s.runtimeApplicationByClientID(ctx, payload.ClientID)
	if err != nil {
		return nil, err
	}
	login, err := s.issueRuntimeLogin(ctx, account, app, LoginTypeWechat)
	if err != nil {
		payload.Status = wechatLoginStatusFailed
		if meta, ok := bizerr.As(err); ok {
			payload.ErrorCode = meta.RuntimeCode()
			payload.Message = meta.Fallback()
		} else {
			payload.Message = err.Error()
		}
		payload.RedirectURL = s.wechatLoginRedirectURL(ctx, payload)
		if updateErr := s.updateRuntimePayload(ctx, token.Id, payload); updateErr != nil {
			return nil, updateErr
		}
		return nil, err
	}
	payload.Status = wechatLoginStatusSuccess
	payload.Login = login
	payload.RedirectURL = s.wechatLoginRedirectURL(ctx, payload)
	if err := s.updateRuntimePayload(ctx, token.Id, payload); err != nil {
		return nil, err
	}
	return wechatLoginResult(payload), nil
}

// GetWechatLoginQRResult returns pending or terminal QR login state.
func (s *serviceImpl) GetWechatLoginQRResult(ctx context.Context, state string) (*WechatLoginQRResultOutput, error) {
	token, payload, err := s.wechatLoginState(ctx, state)
	if err != nil {
		return nil, err
	}
	result := wechatLoginResult(payload)
	if payload.Status != wechatLoginStatusPending {
		_, err = s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "").
			Where(dao.OauthToken.Columns().Id, token.Id).
			Delete()
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func (s *serviceImpl) completeUnboundWechatLogin(
	ctx context.Context,
	token *entity.OauthToken,
	payload *wechatLoginStateData,
	unionID string,
) (*WechatLoginQRResultOutput, error) {
	lookup, err := s.LookupUnionID(ctx, unionID)
	if err != nil {
		return nil, err
	}
	payload.Status = wechatLoginStatusBindNeeded
	payload.ChallengeID = lookup.ChallengeID
	payload.CallbackURL = lookup.CallbackURL
	payload.RedirectURL = s.wechatLoginRedirectURL(ctx, payload)
	if err := s.updateRuntimePayload(ctx, token.Id, payload); err != nil {
		return nil, err
	}
	return wechatLoginResult(payload), nil
}

func (s *serviceImpl) wechatLoginState(ctx context.Context, state string) (*entity.OauthToken, *wechatLoginStateData, error) {
	trimmed := strings.TrimSpace(state)
	if trimmed == "" {
		return nil, nil, bizerr.NewCode(CodeWechatLoginInvalid)
	}
	var token *entity.OauthToken
	err := s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "").
		Where(dao.OauthToken.Columns().Code, ticketCodePrefixWechatLogin+trimmed).
		Scan(&token)
	if err != nil {
		return nil, nil, err
	}
	if token == nil || token.ExpiredAt == nil || token.ExpiredAt.Before(time.Now()) {
		return nil, nil, bizerr.NewCode(CodeWechatLoginInvalid)
	}
	payload := &wechatLoginStateData{}
	if err := json.Unmarshal([]byte(token.Data), payload); err != nil {
		return nil, nil, bizerr.NewCode(CodeWechatLoginInvalid)
	}
	if payload.Kind != ticketKindWechatLogin || payload.State != trimmed {
		return nil, nil, bizerr.NewCode(CodeWechatLoginInvalid)
	}
	return token, payload, nil
}

func (s *serviceImpl) wechatLoginRedirectURL(ctx context.Context, payload *wechatLoginStateData) string {
	baseURL, err := s.configSvc.String(ctx, configKeyWechatLoginRedirectURL, "")
	if err != nil || payload == nil || strings.TrimSpace(baseURL) == "" {
		return ""
	}
	result := callbackWithQuery(baseURL, "state", payload.State)
	result = callbackWithQuery(result, "appid", payload.ClientID)
	result = callbackWithQuery(result, "status", payload.Status)
	callback := payload.Callback
	if strings.TrimSpace(callback) == "" {
		callback = payload.Status
	}
	result = callbackWithQuery(result, "cascallback", callback)
	if payload.ChallengeID != "" {
		result = callbackWithQuery(result, "uuid", payload.ChallengeID)
		result = callbackWithQuery(result, "challengeId", payload.ChallengeID)
	}
	if payload.Message != "" {
		result = callbackWithQuery(result, "msg", payload.Message)
	}
	return result
}

func wechatLoginAuthorizeURL(baseURL string, clientID string, state string, callback string) string {
	result := callbackWithQuery(baseURL, "appid", clientID)
	result = callbackWithQuery(result, "state", state)
	if strings.TrimSpace(callback) != "" {
		result = callbackWithQuery(result, "cascallback", callback)
	}
	return result
}

func wechatLoginResult(payload *wechatLoginStateData) *WechatLoginQRResultOutput {
	if payload == nil {
		return nil
	}
	return &WechatLoginQRResultOutput{
		State:       payload.State,
		Status:      payload.Status,
		RedirectURL: payload.RedirectURL,
		ChallengeID: payload.ChallengeID,
		CallbackURL: payload.CallbackURL,
		ErrorCode:   payload.ErrorCode,
		Message:     payload.Message,
		Login:       payload.Login,
	}
}
