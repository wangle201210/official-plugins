// This file implements CAS ticket validation, application access checks,
// blacklist checks, and OAuth token issuing for the UIdentity CAS plugin.

package uidentity

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"strconv"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const (
	configKeyCASValidateURL = "cas.validateUrl"
)

type casServiceResponse struct {
	AuthenticationSuccess *casAuthenticationSuccess `xml:"authenticationSuccess"`
	AuthenticationFailure *casAuthenticationFailure `xml:"authenticationFailure"`
}

type casAuthenticationSuccess struct {
	User       string        `xml:"user"`
	Attributes casAttributes `xml:"attributes"`
}

type casAuthenticationFailure struct {
	Message string `xml:",chardata"`
}

type casAttributes struct {
	WorkCode string `xml:"workCode"`
	Number   string `xml:"number"`
}

// LoginByCASTicket validates CAS ticket and records login result.
func (s *serviceImpl) LoginByCASTicket(ctx context.Context, in CASLoginInput) (*CASLoginOutput, error) {
	validateURL, err := s.configSvc.String(ctx, configKeyCASValidateURL, "")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(validateURL) == "" {
		return nil, bizerr.NewCode(CodeCASValidateURLMissing)
	}
	number, err := s.validateCASTicket(ctx, validateURL, in)
	if err != nil {
		return nil, err
	}
	account, err := s.getAccountByNumber(ctx, number)
	if err != nil {
		return nil, err
	}
	app, err := s.runtimeApplication(ctx, in.AppID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureRuntimeAccess(ctx, account, app); err != nil {
		return nil, err
	}
	if err := s.recordCASLogin(ctx, account.Id, in.AppID, LoginTypeCAS, "CAS ticket login"); err != nil {
		return nil, err
	}
	return &CASLoginOutput{Number: account.Number, AccountID: account.Id, AppID: in.AppID}, nil
}

// ValidateLegacyAdminCASTicket resolves the old GoAdmin CAS callback account.
func (s *serviceImpl) ValidateLegacyAdminCASTicket(ctx context.Context, in CASLoginInput) (*CASLoginOutput, error) {
	validateURL, err := s.configSvc.String(ctx, configKeyCASValidateURL, "")
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(validateURL) == "" {
		return nil, bizerr.NewCode(CodeCASValidateURLMissing)
	}
	number, err := s.validateCASTicket(ctx, validateURL, in)
	if err != nil {
		return nil, err
	}
	account, err := s.getAccountByNumber(ctx, number)
	if err != nil {
		return nil, err
	}
	if err := s.recordCASLogin(ctx, account.Id, in.AppID, LoginTypeCAS, "Legacy admin CAS callback"); err != nil {
		return nil, err
	}
	out := &CASLoginOutput{Number: account.Number, AccountID: account.Id, AppID: in.AppID}
	if in.AppID <= 0 {
		return out, nil
	}
	app, err := s.runtimeApplication(ctx, in.AppID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureRuntimeAccess(ctx, account, app); err != nil {
		return nil, err
	}
	token, err := s.issueRuntimeAccessToken(ctx, account, app)
	if err != nil {
		return nil, err
	}
	out.AccessToken = token.AccessToken
	out.ExpiredAt = token.ExpiredAt
	return out, nil
}

// IssueOAuthToken issues one OAuth token record and auth log.
func (s *serviceImpl) IssueOAuthToken(ctx context.Context, in OAuthIssueInput) (*OAuthIssueOutput, error) {
	account, err := s.getAccountByID(ctx, in.AccountID)
	if err != nil {
		return nil, err
	}
	app, err := s.runtimeApplication(ctx, in.AppID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureRuntimeAccess(ctx, account, app); err != nil {
		return nil, err
	}
	if in.TtlSeconds <= 0 {
		in.TtlSeconds = 3600
	}
	code, err := randomToken("code")
	if err != nil {
		return nil, err
	}
	access, err := randomToken("access")
	if err != nil {
		return nil, err
	}
	refresh, err := randomToken("refresh")
	if err != nil {
		return nil, err
	}
	expiredAt := time.Now().Add(time.Duration(in.TtlSeconds) * time.Second)
	payload, err := json.Marshal(map[string]any{
		"accountId":   account.Id,
		"number":      account.Number,
		"appId":       app.Id,
		"redirectUri": in.RedirectURI,
		"scope":       in.Scope,
	})
	if err != nil {
		return nil, err
	}
	actorID := s.actorID(ctx)
	_, err = dao.Oauth2Token.Ctx(ctx).Data(do.Oauth2Token{
		ExpiredAt: expiredAt.UnixMilli(),
		Code:      code,
		Access:    access,
		Refresh:   refresh,
		Data:      string(payload),
	}).Insert()
	if err != nil {
		return nil, err
	}
	_, err = dao.OauthLog.Ctx(ctx).Data(do.OauthLog{
		UserId:      account.Id,
		AppId:       app.Id,
		RedirectUri: in.RedirectURI,
		Scope:       in.Scope,
		CreateBy:    actorID,
		UpdateBy:    actorID,
	}).Insert()
	if err != nil {
		return nil, err
	}
	millis := expiredAt.UnixMilli()
	return &OAuthIssueOutput{Code: code, Access: access, Refresh: refresh, ExpiredAt: &millis}, nil
}

func (s *serviceImpl) validateCASTicket(ctx context.Context, validateURL string, in CASLoginInput) (string, error) {
	response, err := g.Client().Post(ctx, validateURL, map[string]string{
		"ticket": in.Ticket,
		"userId": strconv.FormatInt(in.UserID, 10),
	})
	if err != nil {
		return "", err
	}
	defer response.Close()
	body := response.ReadAllString()
	payload := &casServiceResponse{}
	if err := xml.Unmarshal([]byte(body), payload); err != nil {
		return "", err
	}
	if payload.AuthenticationFailure != nil {
		return "", bizerr.NewCode(CodeCASValidationFailed)
	}
	if payload.AuthenticationSuccess == nil {
		return "", bizerr.NewCode(CodeCASValidationFailed)
	}
	number := strings.TrimSpace(payload.AuthenticationSuccess.Attributes.WorkCode)
	if number == "" {
		number = strings.TrimSpace(payload.AuthenticationSuccess.Attributes.Number)
	}
	if number == "" {
		number = strings.TrimSpace(payload.AuthenticationSuccess.User)
	}
	if number == "" {
		return "", bizerr.NewCode(CodeCASValidationFailed)
	}
	return number, nil
}

func (s *serviceImpl) getAccountByID(ctx context.Context, accountID int64) (*entity.Account, error) {
	var account *entity.Account
	err := dao.Account.Ctx(ctx).
		Where(dao.Account.Columns().Id, accountID).
		Scan(&account)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, bizerr.NewCode(CodeResourceNotFound)
	}
	return account, nil
}

func (s *serviceImpl) getAccountByNumber(ctx context.Context, number string) (*entity.Account, error) {
	var account *entity.Account
	err := dao.Account.Ctx(ctx).
		Where(dao.Account.Columns().Number, number).
		Scan(&account)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, bizerr.NewCode(CodeResourceNotFound)
	}
	return account, nil
}

func (s *serviceImpl) runtimeApplication(ctx context.Context, appID int64) (*entity.Applications, error) {
	if appID == 0 {
		return nil, nil
	}
	var app *entity.Applications
	err := dao.Applications.Ctx(ctx).
		Where(dao.Applications.Columns().Id, appID).
		Scan(&app)
	if err != nil {
		return nil, err
	}
	if app == nil {
		return nil, bizerr.NewCode(CodeResourceNotFound)
	}
	if app.Status != ApplicationStatusEnabled {
		return nil, bizerr.NewCode(CodeApplicationDisabled)
	}
	return app, nil
}

func (s *serviceImpl) ensureRuntimeAccess(ctx context.Context, account *entity.Account, app *entity.Applications) error {
	if account.Status == AccountStatusLocked {
		return bizerr.NewCode(CodeAccountLocked)
	}
	if account.Status != AccountStatusNormal {
		return bizerr.NewCode(CodeAccountInactive)
	}
	if app == nil {
		return nil
	}
	now := time.Now()
	accountBlacklistCount, err := dao.AccountAppBlacklist.Ctx(ctx).
		Where(dao.AccountAppBlacklist.Columns().AccountId, account.Id).
		Where(dao.AccountAppBlacklist.Columns().AppId, app.Id).
		Where("("+dao.AccountAppBlacklist.Columns().EffectAt+" IS NULL OR "+dao.AccountAppBlacklist.Columns().EffectAt+" <= ?)", now).
		Where("("+dao.AccountAppBlacklist.Columns().ExpireAt+" IS NULL OR "+dao.AccountAppBlacklist.Columns().ExpireAt+" >= ?)", now).
		Count()
	if err != nil {
		return err
	}
	if accountBlacklistCount > 0 {
		return bizerr.NewCode(CodeAccessDenied)
	}
	groupColumns := dao.AccountGroup.Columns()
	groupBlackColumns := dao.GroupAppBlacklist.Columns()
	groupIDs, err := dao.AccountGroup.Ctx(ctx).
		Fields(groupColumns.GroupsId).
		Where(groupColumns.AccountId, account.Id).
		Array()
	if err != nil {
		return err
	}
	if len(groupIDs) == 0 {
		return nil
	}
	groupBlacklistCount, err := dao.GroupAppBlacklist.Ctx(ctx).
		WhereIn(groupBlackColumns.GroupId, groupIDs).
		Where(groupBlackColumns.AppId, app.Id).
		Where("("+groupBlackColumns.EffectAt+" IS NULL OR "+groupBlackColumns.EffectAt+" <= ?)", now).
		Where("("+groupBlackColumns.ExpireAt+" IS NULL OR "+groupBlackColumns.ExpireAt+" >= ?)", now).
		Count()
	if err != nil {
		return err
	}
	if groupBlacklistCount > 0 {
		return bizerr.NewCode(CodeAccessDenied)
	}
	return nil
}

func (s *serviceImpl) recordCASLogin(ctx context.Context, accountID int64, appID int64, loginType string, message string) error {
	now := time.Now()
	actorID := s.actorID(ctx)
	_, err := dao.CasLoginLog.Ctx(ctx).Data(do.CasLoginLog{
		AccountId:       accountID,
		ChoiceAccountId: accountID,
		AppId:           appID,
		LoginTime:       &now,
		Msg:             message,
		LoginType:       loginType,
		CreateBy:        actorID,
		UpdateBy:        actorID,
	}).Insert()
	return err
}
