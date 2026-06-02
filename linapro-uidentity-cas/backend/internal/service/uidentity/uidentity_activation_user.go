// This file implements activation and user self-service runtime operations for
// the UIdentity CAS source plugin.

package uidentity

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

// accountActiveLogType identifies legacy account_active_log event categories.
type accountActiveLogType int

const (
	// accountActiveLogTypeActivation records activation and Wechat rebind callbacks.
	accountActiveLogTypeActivation accountActiveLogType = 0
	// accountActiveLogTypeUnionBind records explicit UnionID binding requests.
	accountActiveLogTypeUnionBind accountActiveLogType = 1
)

// StartActivation creates an activation challenge after base info matches.
func (s *serviceImpl) StartActivation(ctx context.Context, in ActivationStartInput) (*ActivationOutput, error) {
	account, err := s.getAccountByNumber(ctx, in.Number)
	if err != nil {
		return nil, err
	}
	if account.Name != in.Name {
		return nil, bizerr.NewCode(CodeInvalidCredentials)
	}
	detail, err := s.accountDetailByAccountID(ctx, account.Id)
	if err != nil {
		return nil, err
	}
	if detail.Idcard != in.Idcard {
		return nil, bizerr.NewCode(CodeInvalidCredentials)
	}
	if account.Status == AccountStatusLocked {
		return nil, bizerr.NewCode(CodeAccountLocked)
	}
	challengeID, err := randomToken("act")
	if err != nil {
		return nil, err
	}
	expiredAt := time.Now().Add(activationTTL)
	payload := activationChallengeData{
		AccountID: account.Id,
		Number:    account.Number,
		Stage:     "base",
	}
	if err := s.createRuntimeToken(ctx, do.OauthToken{
		Code:      ticketCodePrefixActivation + challengeID,
		ExpiredAt: &expiredAt,
	}, payload); err != nil {
		return nil, err
	}
	return &ActivationOutput{
		ChallengeID: challengeID,
		NeedFace:    strings.TrimSpace(detail.Face) == "",
		Status:      account.Status,
	}, nil
}

// RecordActivationFace stores face proof for one activation challenge.
func (s *serviceImpl) RecordActivationFace(ctx context.Context, in ActivationFaceInput) (*ActivationStepOutput, error) {
	token, payload, err := s.activationChallenge(ctx, in.ChallengeID)
	if err != nil {
		return nil, err
	}
	if err := s.updateAccountDetailWithAudit(ctx, payload.AccountID, do.AccountDetail{Face: strings.TrimSpace(in.FaceURL), UpdatedBy: s.actorID(ctx)}); err != nil {
		return nil, err
	}
	payload.Stage = "face"
	if err := s.updateRuntimePayload(ctx, token.Id, payload); err != nil {
		return nil, err
	}
	return &ActivationStepOutput{ChallengeID: in.ChallengeID, Success: true}, nil
}

// SetActivationPassword sets password for one activation challenge.
func (s *serviceImpl) SetActivationPassword(ctx context.Context, in ActivationPasswordInput) (*ActivationStepOutput, error) {
	token, payload, err := s.activationChallenge(ctx, in.ChallengeID)
	if err != nil {
		return nil, err
	}
	level, err := s.validatePassword(ctx, in.Password)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if err := s.updateAccountWithAudit(ctx, payload.AccountID, do.Account{
		PasswordHash:      hashPassword(in.Password),
		PasswordUpdatedAt: &now,
		PassLevel:         level,
		UpdatedBy:         s.actorID(ctx),
	}); err != nil {
		return nil, err
	}
	payload.Stage = "password"
	if err := s.updateRuntimePayload(ctx, token.Id, payload); err != nil {
		return nil, err
	}
	return &ActivationStepOutput{ChallengeID: in.ChallengeID, Success: true}, nil
}

// SetActivationPhone binds phone and activates the account.
func (s *serviceImpl) SetActivationPhone(ctx context.Context, in ActivationPhoneInput) (*ActivationStepOutput, error) {
	token, payload, err := s.activationChallenge(ctx, in.ChallengeID)
	if err != nil {
		return nil, err
	}
	if err := s.verifySMSCode(ctx, in.Phone, in.Code, smsTypeCasActive); err != nil {
		return nil, err
	}
	if err := s.ensurePhoneAvailable(ctx, in.Phone, payload.AccountID); err != nil {
		return nil, err
	}
	if err := s.updateAccountWithAudit(ctx, payload.AccountID, do.Account{Phone: strings.TrimSpace(in.Phone), Status: AccountStatusNormal, UpdatedBy: s.actorID(ctx)}); err != nil {
		return nil, err
	}
	payload.Stage = "phone"
	if err := s.updateRuntimePayload(ctx, token.Id, payload); err != nil {
		return nil, err
	}
	return &ActivationStepOutput{ChallengeID: in.ChallengeID, Success: true}, nil
}

// SetActivationWechat binds union ID and activates the account.
func (s *serviceImpl) SetActivationWechat(ctx context.Context, in ActivationWechatInput) (*ActivationStepOutput, error) {
	token, payload, err := s.activationChallenge(ctx, in.ChallengeID)
	if err != nil {
		return nil, err
	}
	if err := s.bindUnionIDToAccount(ctx, payload.AccountID, in.UnionID); err != nil {
		return nil, err
	}
	if err := s.updateAccountWithAudit(ctx, payload.AccountID, do.Account{Status: AccountStatusNormal, UpdatedBy: s.actorID(ctx)}); err != nil {
		return nil, err
	}
	account, err := s.getAccountByID(ctx, payload.AccountID)
	if err != nil {
		return nil, err
	}
	if err := s.recordAccountActiveLog(ctx, account, strings.TrimSpace(in.UnionID), accountActiveLogTypeActivation); err != nil {
		return nil, err
	}
	payload.Stage = "wechat"
	payload.WechatStatus = activationWechatStatusSuccess
	payload.UnionID = strings.TrimSpace(in.UnionID)
	if err := s.updateRuntimePayload(ctx, token.Id, payload); err != nil {
		return nil, err
	}
	return &ActivationStepOutput{ChallengeID: in.ChallengeID, Success: true}, nil
}

// CreateActivationWechatState creates a Wechat authorization state for activation.
func (s *serviceImpl) CreateActivationWechatState(ctx context.Context, in ActivationWechatStateInput) (*ActivationWechatStateOutput, error) {
	token, payload, err := s.activationChallenge(ctx, in.ChallengeID)
	if err != nil {
		return nil, err
	}
	payload.Stage = "wechat_pending"
	payload.Callback = strings.TrimSpace(in.Callback)
	payload.WechatStatus = activationWechatStatusPending
	if err := s.updateRuntimePayload(ctx, token.Id, payload); err != nil {
		return nil, err
	}
	baseURL, err := s.configSvc.String(ctx, configKeyActivationWechatAuthorizeURL, "")
	if err != nil {
		return nil, err
	}
	return &ActivationWechatStateOutput{
		State:  in.ChallengeID,
		Status: payload.WechatStatus,
		URL:    activationWechatAuthorizeURL(baseURL, in.ChallengeID, payload.Callback),
	}, nil
}

// CompleteActivationWechat records a Wechat callback result for activation.
func (s *serviceImpl) CompleteActivationWechat(ctx context.Context, in ActivationWechatCallbackInput) (*ActivationWechatStateOutput, error) {
	token, payload, err := s.activationChallengeUnscoped(ctx, in.State)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(in.Callback) != "" {
		payload.Callback = strings.TrimSpace(in.Callback)
	}
	payload.Code = strings.TrimSpace(in.Code)
	unionID := strings.TrimSpace(in.UnionID)
	if unionID == "" {
		payload.Stage = "wechat"
		payload.WechatStatus = activationWechatStatusUnsupported
		payload.ErrorCode = CodeUnsupportedExternalFlow.RuntimeCode()
		payload.Message = CodeUnsupportedExternalFlow.Fallback()
		payload.RedirectURL = s.activationWechatRedirectURL(ctx, payload, in.State)
		if err := s.updateRuntimePayloadForTenant(ctx, token.TenantId, token.Id, payload); err != nil {
			return nil, err
		}
		return activationWechatResult(in.State, payload), nil
	}
	if err := s.bindUnionIDToAccountForTenant(ctx, token.TenantId, payload.AccountID, unionID); err != nil {
		payload.Stage = "wechat"
		payload.WechatStatus = activationWechatStatusFailed
		if meta, ok := bizerr.As(err); ok {
			payload.ErrorCode = meta.RuntimeCode()
			payload.Message = meta.Fallback()
		} else {
			payload.Message = err.Error()
		}
		payload.RedirectURL = s.activationWechatRedirectURL(ctx, payload, in.State)
		if updateErr := s.updateRuntimePayloadForTenant(ctx, token.TenantId, token.Id, payload); updateErr != nil {
			return nil, updateErr
		}
		return nil, err
	}
	if err := s.updateAccountWithAuditForTenant(ctx, token.TenantId, payload.AccountID, do.Account{Status: AccountStatusNormal, UpdatedBy: s.actorID(ctx)}); err != nil {
		return nil, err
	}
	account, err := s.getAccountByIDForTenant(ctx, token.TenantId, payload.AccountID)
	if err != nil {
		return nil, err
	}
	if err := s.recordAccountActiveLogForTenant(ctx, token.TenantId, account, unionID, accountActiveLogTypeActivation); err != nil {
		return nil, err
	}
	payload.Stage = "wechat"
	payload.WechatStatus = activationWechatStatusSuccess
	payload.UnionID = unionID
	payload.RedirectURL = s.activationWechatRedirectURL(ctx, payload, in.State)
	if err := s.updateRuntimePayloadForTenant(ctx, token.TenantId, token.Id, payload); err != nil {
		return nil, err
	}
	return activationWechatResult(in.State, payload), nil
}

// ActivationState returns one activation challenge state.
func (s *serviceImpl) ActivationState(ctx context.Context, challengeID string) (*ActivationStateOutput, error) {
	_, payload, err := s.activationChallenge(ctx, challengeID)
	if err != nil {
		return nil, err
	}
	account, err := s.getAccountByID(ctx, payload.AccountID)
	if err != nil {
		return nil, err
	}
	return &ActivationStateOutput{
		ChallengeID:  challengeID,
		Success:      account.Status == AccountStatusNormal,
		Status:       account.Status,
		Stage:        payload.Stage,
		WechatStatus: payload.WechatStatus,
		RedirectURL:  payload.RedirectURL,
		ErrorCode:    payload.ErrorCode,
		Message:      payload.Message,
	}, nil
}

// LookupUnionID resolves one union ID or creates a bind challenge.
func (s *serviceImpl) LookupUnionID(ctx context.Context, unionID string) (*UnionIDLookupOutput, error) {
	detail, err := s.accountDetailByUnionID(ctx, unionID)
	if err == nil {
		account, err := s.getAccountByID(ctx, detail.AccountId)
		if err != nil {
			return nil, err
		}
		return &UnionIDLookupOutput{Number: account.Number}, nil
	}
	challengeID, err := randomToken("uid")
	if err != nil {
		return nil, err
	}
	expiredAt := time.Now().Add(unionBindTTL)
	payload := unionIDChallengeData{UnionID: strings.TrimSpace(unionID)}
	if err := s.createRuntimeToken(ctx, do.OauthToken{
		Code:      ticketCodePrefixUnionBind + challengeID,
		ExpiredAt: &expiredAt,
	}, payload); err != nil {
		return nil, err
	}
	callbackURL, err := s.configSvc.String(ctx, configKeyUnionIDBindCallbackURL, "")
	if err != nil {
		return nil, err
	}
	return &UnionIDLookupOutput{
		ChallengeID: challengeID,
		CallbackURL: callbackWithChallenge(callbackURL, challengeID),
	}, nil
}

// BindUnionID consumes one bind challenge and attaches union ID to an account.
func (s *serviceImpl) BindUnionID(ctx context.Context, in UnionIDBindInput) (*UnionIDBindOutput, error) {
	token, payload, err := s.unionIDChallenge(ctx, in.ChallengeID)
	if err != nil {
		return nil, err
	}
	var account *entity.Account
	switch in.BindType {
	case 1:
		if err := s.verifySMSCode(ctx, in.Phone, in.Code, smsTypeCasActive); err != nil {
			return nil, err
		}
		account, err = s.getAccountByPhone(ctx, in.Phone)
	case 2:
		account, err = s.getAccountByNumber(ctx, in.Number)
		if err == nil {
			err = s.verifyAccountPassword(ctx, account, in.Password)
		}
	default:
		err = bizerr.NewCode(CodeUnionIDChallengeInvalid)
	}
	if err != nil {
		return nil, err
	}
	if err := s.rebindUnionIDToAccount(ctx, account, payload.UnionID); err != nil {
		return nil, err
	}
	if _, err := s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "").
		Where(dao.OauthToken.Columns().Id, token.Id).
		Delete(); err != nil {
		return nil, err
	}
	return &UnionIDBindOutput{Number: account.Number}, nil
}

// ChangeRuntimePassword updates one account password.
func (s *serviceImpl) ChangeRuntimePassword(ctx context.Context, number string, newPassword string) error {
	account, err := s.getAccountByNumber(ctx, number)
	if err != nil {
		return err
	}
	return s.ResetAccountPassword(ctx, account.Id, newPassword)
}

// ChangeRuntimePhone updates one account phone after SMS verification.
func (s *serviceImpl) ChangeRuntimePhone(ctx context.Context, in ChangePhoneInput) error {
	account, err := s.getAccountByNumber(ctx, in.Number)
	if err != nil {
		return err
	}
	if err := s.verifySMSCode(ctx, in.Phone, in.Code, smsTypeCasBind); err != nil {
		return err
	}
	if err := s.ensurePhoneAvailable(ctx, in.Phone, account.Id); err != nil {
		return err
	}
	return s.updateAccountWithAudit(ctx, account.Id, do.Account{Phone: strings.TrimSpace(in.Phone), UpdatedBy: s.actorID(ctx)})
}

// ChangeRuntimeEmail updates one account email.
func (s *serviceImpl) ChangeRuntimeEmail(ctx context.Context, number string, email string) error {
	account, err := s.getAccountByNumber(ctx, number)
	if err != nil {
		return err
	}
	return s.updateAccountDetail(ctx, account.Id, do.AccountDetail{Email: strings.TrimSpace(email), UpdatedBy: s.actorID(ctx)})
}

// ChangeRuntimeQQ updates one account QQ.
func (s *serviceImpl) ChangeRuntimeQQ(ctx context.Context, number string, qq string) error {
	account, err := s.getAccountByNumber(ctx, number)
	if err != nil {
		return err
	}
	return s.updateAccountDetail(ctx, account.Id, do.AccountDetail{Qq: strings.TrimSpace(qq), UpdatedBy: s.actorID(ctx)})
}

// UnbindRuntimeWechat clears one account Wechat union ID.
func (s *serviceImpl) UnbindRuntimeWechat(ctx context.Context, number string) error {
	account, err := s.getAccountByNumber(ctx, number)
	if err != nil {
		return err
	}
	return s.updateAccountDetail(ctx, account.Id, do.AccountDetail{Wechat: "", UpdatedBy: s.actorID(ctx)})
}

// GetRuntimeUserInfo returns runtime user projection.
func (s *serviceImpl) GetRuntimeUserInfo(ctx context.Context, number string) (*RuntimeAccount, error) {
	account, err := s.getAccountByNumber(ctx, number)
	if err != nil {
		return nil, err
	}
	return s.runtimeAccountProjection(ctx, account)
}

// ListRuntimeUserLoginLogs returns paged CAS logs for one account.
func (s *serviceImpl) ListRuntimeUserLoginLogs(ctx context.Context, in UserLogListInput) (*ResourceListOutput, error) {
	account, err := s.getAccountByNumber(ctx, in.Number)
	if err != nil {
		return nil, err
	}
	return s.ListResource(ctx, ResourceListInput{
		Resource:  "cas-login-logs",
		AccountId: account.Id,
		PageNum:   normalizedPageNum(in.PageNum),
		PageSize:  normalizedPageSize(in.PageSize),
		OrderBy:   "loginTime",
		Order:     "desc",
	})
}

// ListRuntimeApplications returns enabled applications visible to one account.
func (s *serviceImpl) ListRuntimeApplications(ctx context.Context, in UserApplicationListInput) (*RuntimeApplicationListOutput, error) {
	account, err := s.getAccountByNumber(ctx, in.Number)
	if err != nil {
		return nil, err
	}
	blockedIDs, err := s.blockedApplicationIDs(ctx, account.Id)
	if err != nil {
		return nil, err
	}
	model := s.tenantFilter.Apply(ctx, dao.Application.Ctx(ctx), "").
		Where(dao.Application.Columns().Status, ApplicationStatusEnabled)
	if len(blockedIDs) > 0 {
		model = model.WhereNotIn(dao.Application.Columns().Id, blockedIDs)
	}
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	var apps []*entity.Application
	if err := model.OrderAsc(dao.Application.Columns().Id).
		Page(normalizedPageNum(in.PageNum), normalizedPageSize(in.PageSize)).
		Scan(&apps); err != nil {
		return nil, err
	}
	list := make([]*RuntimeApplication, 0, len(apps))
	for _, app := range apps {
		list = append(list, runtimeApplicationProjection(app))
	}
	return &RuntimeApplicationListOutput{List: list, Total: total}, nil
}

// ListRuntimeAppRoles returns delegated app roles for one granting account.
func (s *serviceImpl) ListRuntimeAppRoles(ctx context.Context, in UserAppRoleListInput) (*ResourceListOutput, error) {
	account, err := s.getAccountByNumber(ctx, in.Number)
	if err != nil {
		return nil, err
	}
	def, err := s.resourceDefinition("account-app-roles")
	if err != nil {
		return nil, err
	}
	cols := dao.AccountAppRole.Columns()
	model := s.tenantFilter.Apply(ctx, dao.AccountAppRole.Ctx(ctx), "").
		Where(cols.GiveAccountId, account.Id)
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	rows, err := model.
		Fields(projectionFields(def)...).
		OrderDesc(cols.Id).
		Page(normalizedPageNum(in.PageNum), normalizedPageSize(in.PageSize)).
		All()
	if err != nil {
		return nil, err
	}
	return &ResourceListOutput{List: projectResult(rows, def), Total: total}, nil
}

// CreateRuntimeAppRole creates one delegated role.
func (s *serviceImpl) CreateRuntimeAppRole(ctx context.Context, in UserAppRoleCreateInput) (int64, error) {
	give, err := s.getAccountByNumber(ctx, in.Number)
	if err != nil {
		return 0, err
	}
	empowered, err := s.getAccountByNumber(ctx, in.EmpoweredNumber)
	if err != nil {
		return 0, err
	}
	if _, err := s.runtimeApplication(ctx, in.AppID); err != nil {
		return 0, err
	}
	tenantID, actorID := s.baseOwnedDO(ctx, true)
	data := do.AccountAppRole{
		TenantId:           tenantID,
		GiveAccountId:      give.Id,
		EmpoweredAccountId: empowered.Id,
		AppId:              in.AppID,
		CreatedBy:          actorID,
		UpdatedBy:          actorID,
	}
	if in.ExpireAt != nil {
		expireAt := time.UnixMilli(*in.ExpireAt)
		data.ExpireAt = &expireAt
	}
	return dao.AccountAppRole.Ctx(ctx).Data(data).InsertAndGetId()
}

// UpdateRuntimeAppRole updates one delegated role owned by a granting account.
func (s *serviceImpl) UpdateRuntimeAppRole(ctx context.Context, in UserAppRoleUpdateInput) error {
	give, err := s.getAccountByNumber(ctx, in.Number)
	if err != nil {
		return err
	}
	var role *entity.AccountAppRole
	if err := s.tenantFilter.Apply(ctx, dao.AccountAppRole.Ctx(ctx), "").
		Where(dao.AccountAppRole.Columns().Id, in.ID).
		Scan(&role); err != nil {
		return err
	}
	if role == nil || role.GiveAccountId != give.Id {
		return bizerr.NewCode(CodeAccessDenied)
	}
	data := do.AccountAppRole{UpdatedBy: s.actorID(ctx)}
	if in.ExpireAt != nil {
		expireAt := time.UnixMilli(*in.ExpireAt)
		data.ExpireAt = &expireAt
	}
	_, err = s.tenantFilter.Apply(ctx, dao.AccountAppRole.Ctx(ctx), "").
		Where(dao.AccountAppRole.Columns().Id, in.ID).
		OmitNilData().
		Data(data).
		Update()
	return err
}

func (s *serviceImpl) activationChallenge(ctx context.Context, challengeID string) (*entity.OauthToken, *activationChallengeData, error) {
	model := s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "")
	return s.activationChallengeByModel(ctx, model, challengeID)
}

func (s *serviceImpl) activationChallengeUnscoped(ctx context.Context, challengeID string) (*entity.OauthToken, *activationChallengeData, error) {
	return s.activationChallengeByModel(ctx, dao.OauthToken.Ctx(ctx), challengeID)
}

func (s *serviceImpl) activationChallengeByModel(ctx context.Context, model *gdb.Model, challengeID string) (*entity.OauthToken, *activationChallengeData, error) {
	trimmed := strings.TrimSpace(challengeID)
	if trimmed == "" {
		return nil, nil, bizerr.NewCode(CodeActivationInvalid)
	}
	var token *entity.OauthToken
	err := model.
		Where(dao.OauthToken.Columns().Code, ticketCodePrefixActivation+trimmed).
		Scan(&token)
	if err != nil {
		return nil, nil, err
	}
	if token == nil || token.ExpiredAt == nil || token.ExpiredAt.Before(time.Now()) {
		return nil, nil, bizerr.NewCode(CodeActivationInvalid)
	}
	payload := &activationChallengeData{}
	if err := json.Unmarshal([]byte(token.Data), payload); err != nil {
		return nil, nil, bizerr.NewCode(CodeActivationInvalid)
	}
	return token, payload, nil
}

func (s *serviceImpl) unionIDChallenge(ctx context.Context, challengeID string) (*entity.OauthToken, *unionIDChallengeData, error) {
	var token *entity.OauthToken
	err := s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "").
		Where(dao.OauthToken.Columns().Code, ticketCodePrefixUnionBind+challengeID).
		Scan(&token)
	if err != nil {
		return nil, nil, err
	}
	if token == nil || token.ExpiredAt == nil || token.ExpiredAt.Before(time.Now()) {
		return nil, nil, bizerr.NewCode(CodeUnionIDChallengeInvalid)
	}
	payload := &unionIDChallengeData{}
	if err := json.Unmarshal([]byte(token.Data), payload); err != nil {
		return nil, nil, bizerr.NewCode(CodeUnionIDChallengeInvalid)
	}
	return token, payload, nil
}

func (s *serviceImpl) updateRuntimePayload(ctx context.Context, tokenID int64, payload any) error {
	content, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "").
		Where(dao.OauthToken.Columns().Id, tokenID).
		Data(do.OauthToken{Data: string(content), UpdatedBy: s.actorID(ctx)}).
		Update()
	return err
}

func (s *serviceImpl) updateAccountDetail(ctx context.Context, accountID int64, data do.AccountDetail) error {
	count, err := s.tenantFilter.Apply(ctx, dao.AccountDetail.Ctx(ctx), "").
		Where(dao.AccountDetail.Columns().AccountId, accountID).
		Count()
	if err != nil {
		return err
	}
	if count == 0 {
		tenantID, actorID := s.baseOwnedDO(ctx, true)
		data.TenantId = tenantID
		data.AccountId = accountID
		data.CreatedBy = actorID
		data.UpdatedBy = actorID
		return s.createAccountDetailWithAudit(ctx, data, accountID)
	}
	return s.updateAccountDetailWithAudit(ctx, accountID, data)
}

func (s *serviceImpl) bindUnionIDToAccount(ctx context.Context, accountID int64, unionID string) error {
	trimmed := strings.TrimSpace(unionID)
	if trimmed == "" {
		return bizerr.NewCode(CodeUnionIDChallengeInvalid)
	}
	count, err := s.tenantFilter.Apply(ctx, dao.AccountDetail.Ctx(ctx), "").
		Where(dao.AccountDetail.Columns().Wechat, trimmed).
		WhereNot(dao.AccountDetail.Columns().AccountId, accountID).
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewCode(CodeContactConflict)
	}
	return s.updateAccountDetail(ctx, accountID, do.AccountDetail{Wechat: trimmed, UpdatedBy: s.actorID(ctx)})
}

// rebindUnionIDToAccount moves one UnionID from any previous account to account.
func (s *serviceImpl) rebindUnionIDToAccount(ctx context.Context, account *entity.Account, unionID string) error {
	trimmed := strings.TrimSpace(unionID)
	if trimmed == "" || account == nil {
		return bizerr.NewCode(CodeUnionIDChallengeInvalid)
	}
	return dao.AccountDetail.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		detailCols := dao.AccountDetail.Columns()
		oldRecords, err := s.accountDetailRecordsByWechat(ctx, trimmed, account.Id)
		if err != nil {
			return err
		}
		_, err = s.tenantFilter.Apply(ctx, dao.AccountDetail.Ctx(ctx), "").
			Where(detailCols.Wechat, trimmed).
			WhereNot(detailCols.AccountId, account.Id).
			Data(do.AccountDetail{Wechat: "", UpdatedBy: s.actorID(ctx)}).
			Update()
		if err != nil {
			return err
		}
		if err := s.insertAccountDetailUpdateAudits(ctx, oldRecords); err != nil {
			return err
		}
		if err := s.updateAccountDetail(ctx, account.Id, do.AccountDetail{Wechat: trimmed, UpdatedBy: s.actorID(ctx)}); err != nil {
			return err
		}
		return s.recordAccountActiveLog(ctx, account, trimmed, accountActiveLogTypeUnionBind)
	})
}

// getAccountByIDForTenant loads one account from an explicit tenant boundary.
func (s *serviceImpl) getAccountByIDForTenant(ctx context.Context, tenantID int, accountID int64) (*entity.Account, error) {
	var account *entity.Account
	err := dao.Account.Ctx(ctx).
		Where(dao.Account.Columns().TenantId, tenantID).
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

// recordAccountActiveLog writes one legacy activation log in the current tenant.
func (s *serviceImpl) recordAccountActiveLog(ctx context.Context, account *entity.Account, unionID string, logType accountActiveLogType) error {
	if account == nil {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	return s.recordAccountActiveLogForTenant(ctx, s.tenantID(ctx), account, unionID, logType)
}

// recordAccountActiveLogForTenant writes one legacy activation log for tenantID.
func (s *serviceImpl) recordAccountActiveLogForTenant(ctx context.Context, tenantID int, account *entity.Account, unionID string, logType accountActiveLogType) error {
	if account == nil {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	actorID := s.actorID(ctx)
	_, err := dao.AccountActiveLog.Ctx(ctx).Data(do.AccountActiveLog{
		TenantId:  tenantID,
		Number:    strings.TrimSpace(account.Number),
		Phone:     strings.TrimSpace(account.Phone),
		Wechat:    strings.TrimSpace(unionID),
		Type:      int(logType),
		CreatedBy: actorID,
		UpdatedBy: actorID,
	}).Insert()
	return err
}

func (s *serviceImpl) activationWechatRedirectURL(ctx context.Context, payload *activationChallengeData, state string) string {
	baseURL, err := s.configSvc.String(ctx, configKeyActivationWechatRedirectURL, "")
	if err != nil || payload == nil || strings.TrimSpace(baseURL) == "" {
		return ""
	}
	result := callbackWithQuery(baseURL, "state", strings.TrimSpace(state))
	result = callbackWithQuery(result, "status", payload.WechatStatus)
	callback := payload.Callback
	if strings.TrimSpace(callback) == "" {
		callback = "active"
	}
	result = callbackWithQuery(result, "cascallback", callback)
	if payload.Message != "" {
		result = callbackWithQuery(result, "msg", payload.Message)
	}
	return result
}

func activationWechatAuthorizeURL(baseURL string, state string, callback string) string {
	result := callbackWithQuery(baseURL, "state", state)
	if strings.TrimSpace(callback) != "" {
		result = callbackWithQuery(result, "cascallback", callback)
	}
	return result
}

func activationWechatResult(state string, payload *activationChallengeData) *ActivationWechatStateOutput {
	if payload == nil {
		return nil
	}
	return &ActivationWechatStateOutput{
		State:       strings.TrimSpace(state),
		Status:      payload.WechatStatus,
		Success:     payload.WechatStatus == activationWechatStatusSuccess,
		RedirectURL: payload.RedirectURL,
		ErrorCode:   payload.ErrorCode,
		Message:     payload.Message,
	}
}

func (s *serviceImpl) ensurePhoneAvailable(ctx context.Context, phone string, accountID int64) error {
	count, err := s.tenantFilter.Apply(ctx, dao.Account.Ctx(ctx), "").
		Where(dao.Account.Columns().Phone, strings.TrimSpace(phone)).
		WhereNot(dao.Account.Columns().Id, accountID).
		Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewCode(CodeContactConflict)
	}
	return nil
}

func callbackWithChallenge(callbackURL string, challengeID string) string {
	if strings.TrimSpace(callbackURL) == "" {
		return ""
	}
	return callbackWithQuery(callbackURL, "challengeId", challengeID)
}

func callbackWithQuery(callbackURL string, key string, value string) string {
	trimmed := strings.TrimSpace(callbackURL)
	if trimmed == "" {
		return ""
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}
	query := parsed.Query()
	query.Set(key, value)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func (s *serviceImpl) blockedApplicationIDs(ctx context.Context, accountID int64) ([]int64, error) {
	now := time.Now()
	blocked := make(map[int64]struct{})
	accountCols := dao.AccountAppBlacklist.Columns()
	accountRows, err := s.tenantFilter.Apply(ctx, dao.AccountAppBlacklist.Ctx(ctx), "").
		Fields(accountCols.AppId).
		Where(accountCols.AccountId, accountID).
		Where("("+accountCols.EffectAt+" IS NULL OR "+accountCols.EffectAt+" <= ?)", now).
		Where("("+accountCols.ExpireAt+" IS NULL OR "+accountCols.ExpireAt+" >= ?)", now).
		All()
	if err != nil {
		return nil, err
	}
	for _, row := range accountRows {
		blocked[row[accountCols.AppId].Int64()] = struct{}{}
	}
	groupIDs, err := s.tenantFilter.Apply(ctx, dao.AccountGroup.Ctx(ctx), "").
		Fields(dao.AccountGroup.Columns().GroupId).
		Where(dao.AccountGroup.Columns().AccountId, accountID).
		Array()
	if err != nil {
		return nil, err
	}
	if len(groupIDs) > 0 {
		groupCols := dao.GroupAppBlacklist.Columns()
		groupRows, err := s.tenantFilter.Apply(ctx, dao.GroupAppBlacklist.Ctx(ctx), "").
			Fields(groupCols.AppId).
			WhereIn(groupCols.GroupId, groupIDs).
			Where("("+groupCols.EffectAt+" IS NULL OR "+groupCols.EffectAt+" <= ?)", now).
			Where("("+groupCols.ExpireAt+" IS NULL OR "+groupCols.ExpireAt+" >= ?)", now).
			All()
		if err != nil {
			return nil, err
		}
		for _, row := range groupRows {
			blocked[row[groupCols.AppId].Int64()] = struct{}{}
		}
	}
	result := make([]int64, 0, len(blocked))
	for id := range blocked {
		if id > 0 {
			result = append(result, id)
		}
	}
	return result, nil
}

func normalizedPageNum(pageNum int) int {
	if pageNum <= 0 {
		return 1
	}
	return pageNum
}

func normalizedPageSize(pageSize int) int {
	switch {
	case pageSize <= 0:
		return 20
	case pageSize > 100:
		return 100
	default:
		return pageSize
	}
}
