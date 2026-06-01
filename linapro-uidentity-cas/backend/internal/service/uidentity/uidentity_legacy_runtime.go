// This file implements legacy-compatible CAS, token, activation, and user
// runtime flows using plugin-owned account and token storage.

package uidentity

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/mssola/useragent"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const (
	ticketKindTGT        = "tgt"
	ticketKindST         = "st"
	ticketKindAccess     = "access"
	ticketKindActivation = "activation"
	ticketKindUnionBind  = "union_bind"

	ticketCodePrefixTGT        = "tgt:"
	ticketCodePrefixST         = "st:"
	ticketAccessPrefixRuntime  = "runtime:"
	ticketCodePrefixActivation = "activation:"
	ticketCodePrefixUnionBind  = "union_bind:"

	tgtTTL        = 10 * 24 * time.Hour
	stTTL         = 5 * time.Minute
	accessTTL     = 2 * time.Hour
	activationTTL = 30 * time.Minute
	unionBindTTL  = 10 * time.Minute

	configKeyUnionIDBindCallbackURL = "runtime.unionIdBindCallbackUrl"
)

type runtimeTicketPayload struct {
	Kind           string `json:"kind"`
	Ticket         string `json:"ticket"`
	AccountID      int64  `json:"accountId"`
	OwnerAccountID int64  `json:"ownerAccountId"`
	Number         string `json:"number"`
	AppID          int64  `json:"appId"`
	ClientID       string `json:"clientId"`
	Service        string `json:"service"`
	LogID          int64  `json:"logId"`
}

type activationChallengeData struct {
	AccountID int64  `json:"accountId"`
	Number    string `json:"number"`
	Stage     string `json:"stage"`
}

type unionIDChallengeData struct {
	UnionID string `json:"unionId"`
}

// LoginByPassword validates password login and issues CAS runtime tickets.
func (s *serviceImpl) LoginByPassword(ctx context.Context, in PasswordLoginInput) (*RuntimeLoginOutput, error) {
	app, err := s.runtimeApplicationByClientID(ctx, in.ClientID)
	if err != nil {
		return nil, err
	}
	account, err := s.getAccountByNumber(ctx, in.Number)
	if err != nil {
		return nil, err
	}
	if err := s.verifyAccountPassword(ctx, account, in.Password); err != nil {
		return nil, err
	}
	return s.issueRuntimeLogin(ctx, account, app, LoginTypePassword)
}

// LoginByPhone validates phone login and issues CAS runtime tickets.
func (s *serviceImpl) LoginByPhone(ctx context.Context, in PhoneLoginInput) (*RuntimeLoginOutput, error) {
	app, err := s.runtimeApplicationByClientID(ctx, in.ClientID)
	if err != nil {
		return nil, err
	}
	if err := s.verifySMSCode(ctx, in.Phone, in.Code, smsTypeCasLogin); err != nil {
		return nil, err
	}
	account, err := s.getAccountByPhone(ctx, in.Phone)
	if err != nil {
		return nil, err
	}
	return s.issueRuntimeLogin(ctx, account, app, LoginTypeSMS)
}

// LoginByUnionID resolves one bound Wechat union ID and issues tickets.
func (s *serviceImpl) LoginByUnionID(ctx context.Context, in UnionIDLoginInput) (*RuntimeLoginOutput, error) {
	app, err := s.runtimeApplicationByClientID(ctx, in.ClientID)
	if err != nil {
		return nil, err
	}
	account, err := s.getAccountByUnionID(ctx, in.UnionID)
	if err != nil {
		return nil, err
	}
	return s.issueRuntimeLogin(ctx, account, app, LoginTypeUnionID)
}

// IssueServiceTicketFromTGT issues one ST from an existing TGT.
func (s *serviceImpl) IssueServiceTicketFromTGT(ctx context.Context, in ServiceTicketInput) (*ServiceTicketOutput, error) {
	app, err := s.runtimeApplicationByClientID(ctx, in.ClientID)
	if err != nil {
		return nil, err
	}
	payload, err := s.runtimeTicketByCode(ctx, ticketCodePrefixTGT+in.TGT, ticketKindTGT)
	if err != nil {
		return nil, err
	}
	selectedID := payload.AccountID
	if in.AccountID > 0 {
		selectedID = in.AccountID
	}
	if selectedID != payload.AccountID {
		allowed, err := s.hasDelegatedAccess(ctx, payload.AccountID, selectedID, app.Id)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, bizerr.NewCode(CodeAccessDenied)
		}
	}
	selected, err := s.getAccountByID(ctx, selectedID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureRuntimeAccess(ctx, selected, app); err != nil {
		return nil, err
	}
	st, err := s.issueServiceTicket(ctx, app, payload.AccountID, selectedID, 0)
	if err != nil {
		return nil, err
	}
	return &ServiceTicketOutput{ST: st, CallbackURL: callbackWithTicket(app.CallbackUrl, st)}, nil
}

// ValidateServiceTicket consumes and validates one ST.
func (s *serviceImpl) ValidateServiceTicket(ctx context.Context, in ServiceValidateInput) (*ServiceValidateOutput, error) {
	token, payload, err := s.runtimeTicketRecordByCode(ctx, ticketCodePrefixST+in.Ticket, ticketKindST)
	if err != nil {
		return nil, err
	}
	ownerID := payload.OwnerAccountID
	if ownerID == 0 {
		ownerID = payload.AccountID
	}
	selectedID := payload.AccountID
	if in.UserID > 0 {
		selectedID = in.UserID
	}
	if selectedID != ownerID {
		allowed, err := s.hasDelegatedAccess(ctx, ownerID, selectedID, payload.AppID)
		if err != nil {
			return nil, err
		}
		if !allowed {
			return nil, bizerr.NewCode(CodeAccessDenied)
		}
	}
	app, err := s.runtimeApplication(ctx, payload.AppID)
	if err != nil {
		return nil, err
	}
	account, err := s.getAccountByID(ctx, selectedID)
	if err != nil {
		return nil, err
	}
	if err := s.ensureRuntimeAccess(ctx, account, app); err != nil {
		return nil, err
	}
	if payload.LogID > 0 {
		if _, err := s.tenantFilter.Apply(ctx, dao.CasLoginLog.Ctx(ctx), "").
			Where(dao.CasLoginLog.Columns().Id, payload.LogID).
			Data(do.CasLoginLog{ChoiceAccountId: selectedID, UpdatedBy: s.actorID(ctx)}).
			Update(); err != nil {
			return nil, err
		}
	}
	if _, err := s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "").
		Where(dao.OauthToken.Columns().Id, token.Id).
		Delete(); err != nil {
		return nil, err
	}
	user, err := s.runtimeAccountProjection(ctx, account)
	if err != nil {
		return nil, err
	}
	return &ServiceValidateOutput{
		Ticket:  in.Ticket,
		User:    user,
		App:     runtimeApplicationProjection(app),
		Success: true,
	}, nil
}

// DeleteTicket deletes a runtime ticket by value.
func (s *serviceImpl) DeleteTicket(ctx context.Context, ticket string) error {
	trimmed := strings.TrimSpace(ticket)
	if trimmed == "" {
		return nil
	}
	_, err := s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "").
		Where("("+dao.OauthToken.Columns().Code+" = ? OR "+dao.OauthToken.Columns().Code+" = ? OR "+dao.OauthToken.Columns().Access+" = ?)", ticketCodePrefixTGT+trimmed, ticketCodePrefixST+trimmed, ticketAccessPrefixRuntime+trimmed).
		Delete()
	return err
}

// IssueRuntimeToken validates password and application secret before issuing AT.
func (s *serviceImpl) IssueRuntimeToken(ctx context.Context, in RuntimeTokenInput) (*RuntimeTokenOutput, error) {
	app, err := s.runtimeApplicationByClientID(ctx, in.ClientID)
	if err != nil {
		return nil, err
	}
	if app.SecretKey != in.Secret {
		return nil, bizerr.NewCode(CodeApplicationSecretInvalid)
	}
	account, err := s.getAccountByNumber(ctx, in.Number)
	if err != nil {
		return nil, err
	}
	if err := s.verifyAccountPassword(ctx, account, in.Password); err != nil {
		return nil, err
	}
	if err := s.ensureRuntimeAccess(ctx, account, app); err != nil {
		return nil, err
	}
	access, err := randomToken("AT")
	if err != nil {
		return nil, err
	}
	expiredAt := time.Now().Add(accessTTL)
	payload := runtimeTicketPayload{
		Kind:      ticketKindAccess,
		Ticket:    access,
		AccountID: account.Id,
		Number:    account.Number,
		AppID:     app.Id,
		ClientID:  app.ClientId,
		Service:   app.ClientId,
	}
	if err := s.createRuntimeToken(ctx, do.OauthToken{
		Code:      "access:" + access,
		Access:    ticketAccessPrefixRuntime + access,
		ExpiredAt: &expiredAt,
	}, payload); err != nil {
		return nil, err
	}
	millis := expiredAt.UnixMilli()
	return &RuntimeTokenOutput{AccessToken: access, ExpiredAt: &millis}, nil
}

// GetUserInfoByRuntimeToken returns token-bound account data.
func (s *serviceImpl) GetUserInfoByRuntimeToken(ctx context.Context, accessToken string) (*RuntimeTokenInfoOutput, error) {
	payload, err := s.runtimeTicketByAccess(ctx, ticketAccessPrefixRuntime+accessToken, ticketKindAccess)
	if err != nil {
		return nil, err
	}
	account, err := s.getAccountByID(ctx, payload.AccountID)
	if err != nil {
		return nil, err
	}
	app, err := s.runtimeApplication(ctx, payload.AppID)
	if err != nil {
		return nil, err
	}
	accounts, err := s.accessibleAccounts(ctx, account, app)
	if err != nil {
		return nil, err
	}
	return &RuntimeTokenInfoOutput{
		User:  accounts[0],
		Users: accounts,
		App:   runtimeApplicationProjection(app),
	}, nil
}

func (s *serviceImpl) issueRuntimeLogin(ctx context.Context, account *entity.Account, app *entity.Application, loginType string) (*RuntimeLoginOutput, error) {
	if err := s.ensureRuntimeAccess(ctx, account, app); err != nil {
		return nil, err
	}
	accounts, err := s.accessibleAccounts(ctx, account, app)
	if err != nil {
		return nil, err
	}
	tgt, err := s.issueTGT(ctx, account, app)
	if err != nil {
		return nil, err
	}
	logID, err := s.recordCASLoginWithID(ctx, account.Id, account.Id, app.Id, loginType, "CAS runtime login")
	if err != nil {
		return nil, err
	}
	st, err := s.issueServiceTicket(ctx, app, account.Id, account.Id, logID)
	if err != nil {
		return nil, err
	}
	return &RuntimeLoginOutput{
		CallbackURL: callbackWithTicket(app.CallbackUrl, st),
		TGT:         tgt,
		ST:          st,
		User:        accounts[0],
		Users:       accounts,
		App:         runtimeApplicationProjection(app),
	}, nil
}

func (s *serviceImpl) issueTGT(ctx context.Context, account *entity.Account, app *entity.Application) (string, error) {
	tgt, err := randomToken("TGT")
	if err != nil {
		return "", err
	}
	expiredAt := time.Now().Add(tgtTTL)
	payload := runtimeTicketPayload{
		Kind:      ticketKindTGT,
		Ticket:    tgt,
		AccountID: account.Id,
		Number:    account.Number,
		AppID:     app.Id,
		ClientID:  app.ClientId,
		Service:   app.ClientId,
	}
	if err := s.createRuntimeToken(ctx, do.OauthToken{
		Code:      ticketCodePrefixTGT + tgt,
		ExpiredAt: &expiredAt,
	}, payload); err != nil {
		return "", err
	}
	return tgt, nil
}

func (s *serviceImpl) issueServiceTicket(ctx context.Context, app *entity.Application, ownerID int64, selectedID int64, logID int64) (string, error) {
	st, err := randomToken("ST")
	if err != nil {
		return "", err
	}
	expiredAt := time.Now().Add(stTTL)
	payload := runtimeTicketPayload{
		Kind:           ticketKindST,
		Ticket:         st,
		AccountID:      selectedID,
		OwnerAccountID: ownerID,
		AppID:          app.Id,
		ClientID:       app.ClientId,
		Service:        app.ClientId,
		LogID:          logID,
	}
	if err := s.createRuntimeToken(ctx, do.OauthToken{
		Code:      ticketCodePrefixST + st,
		ExpiredAt: &expiredAt,
	}, payload); err != nil {
		return "", err
	}
	return st, nil
}

func (s *serviceImpl) createRuntimeToken(ctx context.Context, data do.OauthToken, payload any) error {
	content, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	tenantID, actorID := s.baseOwnedDO(ctx, true)
	data.TenantId = tenantID
	data.Data = string(content)
	data.CreatedBy = actorID
	data.UpdatedBy = actorID
	_, err = dao.OauthToken.Ctx(ctx).Data(data).Insert()
	return err
}

func (s *serviceImpl) runtimeTicketByCode(ctx context.Context, code string, kind string) (*runtimeTicketPayload, error) {
	_, payload, err := s.runtimeTicketRecordByCode(ctx, code, kind)
	return payload, err
}

func (s *serviceImpl) runtimeTicketRecordByCode(ctx context.Context, code string, kind string) (*entity.OauthToken, *runtimeTicketPayload, error) {
	var token *entity.OauthToken
	err := s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "").
		Where(dao.OauthToken.Columns().Code, code).
		Scan(&token)
	if err != nil {
		return nil, nil, err
	}
	return parseRuntimeToken(token, kind)
}

func (s *serviceImpl) runtimeTicketByAccess(ctx context.Context, access string, kind string) (*runtimeTicketPayload, error) {
	var token *entity.OauthToken
	err := s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "").
		Where(dao.OauthToken.Columns().Access, access).
		Scan(&token)
	if err != nil {
		return nil, err
	}
	_, payload, err := parseRuntimeToken(token, kind)
	return payload, err
}

func parseRuntimeToken(token *entity.OauthToken, kind string) (*entity.OauthToken, *runtimeTicketPayload, error) {
	if token == nil || token.ExpiredAt == nil || token.ExpiredAt.Before(time.Now()) {
		return nil, nil, bizerr.NewCode(CodeTicketInvalid)
	}
	payload := &runtimeTicketPayload{}
	if err := json.Unmarshal([]byte(token.Data), payload); err != nil {
		return nil, nil, bizerr.NewCode(CodeTicketInvalid)
	}
	if payload.Kind != kind {
		return nil, nil, bizerr.NewCode(CodeTicketInvalid)
	}
	return token, payload, nil
}

func (s *serviceImpl) runtimeApplicationByClientID(ctx context.Context, clientID string) (*entity.Application, error) {
	var app *entity.Application
	err := s.tenantFilter.Apply(ctx, dao.Application.Ctx(ctx), "").
		Where(dao.Application.Columns().ClientId, strings.TrimSpace(clientID)).
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

func (s *serviceImpl) getAccountByPhone(ctx context.Context, phone string) (*entity.Account, error) {
	var account *entity.Account
	err := s.tenantFilter.Apply(ctx, dao.Account.Ctx(ctx), "").
		Where(dao.Account.Columns().Phone, strings.TrimSpace(phone)).
		Scan(&account)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, bizerr.NewCode(CodeResourceNotFound)
	}
	return account, nil
}

func (s *serviceImpl) getAccountByUnionID(ctx context.Context, unionID string) (*entity.Account, error) {
	detail, err := s.accountDetailByUnionID(ctx, unionID)
	if err != nil {
		return nil, err
	}
	return s.getAccountByID(ctx, detail.AccountId)
}

func (s *serviceImpl) accountDetailByUnionID(ctx context.Context, unionID string) (*entity.AccountDetail, error) {
	var detail *entity.AccountDetail
	err := s.tenantFilter.Apply(ctx, dao.AccountDetail.Ctx(ctx), "").
		Where(dao.AccountDetail.Columns().Wechat, strings.TrimSpace(unionID)).
		Scan(&detail)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return nil, bizerr.NewCode(CodeResourceNotFound)
	}
	return detail, nil
}

func (s *serviceImpl) accountDetailByAccountID(ctx context.Context, accountID int64) (*entity.AccountDetail, error) {
	var detail *entity.AccountDetail
	err := s.tenantFilter.Apply(ctx, dao.AccountDetail.Ctx(ctx), "").
		Where(dao.AccountDetail.Columns().AccountId, accountID).
		Scan(&detail)
	if err != nil {
		return nil, err
	}
	if detail == nil {
		return &entity.AccountDetail{AccountId: accountID}, nil
	}
	return detail, nil
}

func passwordMatches(account *entity.Account, password string) bool {
	return account != nil && account.PasswordHash != "" && account.PasswordHash == hashPassword(password)
}

func callbackWithTicket(callbackURL string, ticket string) string {
	return callbackWithQuery(callbackURL, "ticket", ticket)
}

func runtimeApplicationProjection(app *entity.Application) *RuntimeApplication {
	if app == nil {
		return nil
	}
	return &RuntimeApplication{
		ID:          app.Id,
		Name:        app.Name,
		Alias:       app.Alias,
		ClientID:    app.ClientId,
		AccessModel: app.AccessModel,
		CallbackURL: app.CallbackUrl,
	}
}

func (s *serviceImpl) runtimeAccountProjection(ctx context.Context, account *entity.Account) (*RuntimeAccount, error) {
	if account == nil {
		return nil, bizerr.NewCode(CodeResourceNotFound)
	}
	var (
		containerName string
		unitName      string
	)
	if account.ContainerId > 0 {
		names, err := s.nameMap(ctx, dao.Container.Ctx(ctx), dao.Container.Columns().Id, dao.Container.Columns().Alias, []int64{account.ContainerId})
		if err != nil {
			return nil, err
		}
		containerName = names[account.ContainerId]
	}
	if account.UnitId > 0 {
		names, err := s.nameMap(ctx, dao.Unit.Ctx(ctx), dao.Unit.Columns().Id, dao.Unit.Columns().Alias, []int64{account.UnitId})
		if err != nil {
			return nil, err
		}
		unitName = names[account.UnitId]
	}
	groupNames, err := s.accountGroupNames(ctx, []int64{account.Id})
	if err != nil {
		return nil, err
	}
	detail, err := s.accountDetailByAccountID(ctx, account.Id)
	if err != nil {
		return nil, err
	}
	return &RuntimeAccount{
		ID:            account.Id,
		Number:        account.Number,
		Name:          account.Name,
		Phone:         account.Phone,
		Status:        account.Status,
		PassLevel:     account.PassLevel,
		ContainerID:   account.ContainerId,
		ContainerName: containerName,
		UnitID:        account.UnitId,
		UnitName:      unitName,
		ExpireAt:      apitime.Milli(account.ExpireAt),
		Groups:        groupNames[account.Id],
		Detail: &RuntimeAccountDetail{
			Birthday: detail.Birthday,
			Email:    detail.Email,
			Gender:   detail.Gender,
			QQ:       detail.Qq,
			Wechat:   detail.Wechat,
			Idcard:   detail.Idcard,
			Avatar:   detail.Avatar,
			Face:     detail.Face,
		},
	}, nil
}

func (s *serviceImpl) runtimeAccountProjectionBatch(ctx context.Context, accounts []*entity.Account) ([]*RuntimeAccount, error) {
	result := make([]*RuntimeAccount, 0, len(accounts))
	if len(accounts) == 0 {
		return result, nil
	}
	accountIDs := make([]int64, 0, len(accounts))
	containerIDs := make([]int64, 0, len(accounts))
	unitIDs := make([]int64, 0, len(accounts))
	for _, account := range accounts {
		accountIDs = append(accountIDs, account.Id)
		if account.ContainerId > 0 {
			containerIDs = append(containerIDs, account.ContainerId)
		}
		if account.UnitId > 0 {
			unitIDs = append(unitIDs, account.UnitId)
		}
	}
	containerNames, err := s.nameMap(ctx, dao.Container.Ctx(ctx), dao.Container.Columns().Id, dao.Container.Columns().Alias, uniqueInt64s(containerIDs))
	if err != nil {
		return nil, err
	}
	unitNames, err := s.nameMap(ctx, dao.Unit.Ctx(ctx), dao.Unit.Columns().Id, dao.Unit.Columns().Alias, uniqueInt64s(unitIDs))
	if err != nil {
		return nil, err
	}
	groupNames, err := s.accountGroupNames(ctx, accountIDs)
	if err != nil {
		return nil, err
	}
	details, err := s.accountDetailsByAccountIDs(ctx, accountIDs)
	if err != nil {
		return nil, err
	}
	for _, account := range accounts {
		detail := details[account.Id]
		if detail == nil {
			detail = &entity.AccountDetail{AccountId: account.Id}
		}
		result = append(result, &RuntimeAccount{
			ID:            account.Id,
			Number:        account.Number,
			Name:          account.Name,
			Phone:         account.Phone,
			Status:        account.Status,
			PassLevel:     account.PassLevel,
			ContainerID:   account.ContainerId,
			ContainerName: containerNames[account.ContainerId],
			UnitID:        account.UnitId,
			UnitName:      unitNames[account.UnitId],
			ExpireAt:      apitime.Milli(account.ExpireAt),
			Groups:        groupNames[account.Id],
			Detail: &RuntimeAccountDetail{
				Birthday: detail.Birthday,
				Email:    detail.Email,
				Gender:   detail.Gender,
				QQ:       detail.Qq,
				Wechat:   detail.Wechat,
				Idcard:   detail.Idcard,
				Avatar:   detail.Avatar,
				Face:     detail.Face,
			},
		})
	}
	return result, nil
}

func (s *serviceImpl) accountDetailsByAccountIDs(ctx context.Context, accountIDs []int64) (map[int64]*entity.AccountDetail, error) {
	result := make(map[int64]*entity.AccountDetail)
	if len(accountIDs) == 0 {
		return result, nil
	}
	var details []*entity.AccountDetail
	err := s.tenantFilter.Apply(ctx, dao.AccountDetail.Ctx(ctx), "").
		WhereIn(dao.AccountDetail.Columns().AccountId, accountIDs).
		Scan(&details)
	if err != nil {
		return nil, err
	}
	for _, detail := range details {
		result[detail.AccountId] = detail
	}
	return result, nil
}

func (s *serviceImpl) accessibleAccounts(ctx context.Context, account *entity.Account, app *entity.Application) ([]*RuntimeAccount, error) {
	now := time.Now()
	roleCols := dao.AccountAppRole.Columns()
	rows, err := s.tenantFilter.Apply(ctx, dao.AccountAppRole.Ctx(ctx), "").
		Fields(roleCols.GiveAccountId).
		Where(roleCols.EmpoweredAccountId, account.Id).
		Where(roleCols.AppId, app.Id).
		Where("("+roleCols.ExpireAt+" IS NULL OR "+roleCols.ExpireAt+" >= ?)", now).
		All()
	if err != nil {
		return nil, err
	}
	ids := []int64{account.Id}
	for _, row := range rows {
		id := row[roleCols.GiveAccountId].Int64()
		if id > 0 {
			ids = append(ids, id)
		}
	}
	var accounts []*entity.Account
	if err := s.tenantFilter.Apply(ctx, dao.Account.Ctx(ctx), "").
		WhereIn(dao.Account.Columns().Id, uniqueInt64s(ids)).
		Scan(&accounts); err != nil {
		return nil, err
	}
	candidateIDs := make([]int64, 0, len(accounts))
	for _, candidate := range accounts {
		candidateIDs = append(candidateIDs, candidate.Id)
	}
	accountBlocked, err := s.blockedAccountIDSet(ctx, candidateIDs, app.Id)
	if err != nil {
		return nil, err
	}
	groupBlocked, err := s.groupBlockedAccountIDSet(ctx, candidateIDs, app.Id)
	if err != nil {
		return nil, err
	}
	allowed := make([]*entity.Account, 0, len(accounts))
	for _, candidate := range accounts {
		if candidate.Status != AccountStatusNormal {
			continue
		}
		if _, ok := accountBlocked[candidate.Id]; ok {
			continue
		}
		if _, ok := groupBlocked[candidate.Id]; ok {
			continue
		}
		allowed = append(allowed, candidate)
	}
	if len(allowed) == 0 {
		return nil, bizerr.NewCode(CodeAccessDenied)
	}
	ordered := orderPrimaryFirst(allowed, account.Id)
	return s.runtimeAccountProjectionBatch(ctx, ordered)
}

func (s *serviceImpl) blockedAccountIDSet(ctx context.Context, accountIDs []int64, appID int64) (map[int64]struct{}, error) {
	blocked := make(map[int64]struct{})
	if len(accountIDs) == 0 {
		return blocked, nil
	}
	now := time.Now()
	cols := dao.AccountAppBlacklist.Columns()
	rows, err := s.tenantFilter.Apply(ctx, dao.AccountAppBlacklist.Ctx(ctx), "").
		Fields(cols.AccountId).
		WhereIn(cols.AccountId, accountIDs).
		Where(cols.AppId, appID).
		Where("("+cols.EffectAt+" IS NULL OR "+cols.EffectAt+" <= ?)", now).
		Where("("+cols.ExpireAt+" IS NULL OR "+cols.ExpireAt+" >= ?)", now).
		All()
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		blocked[row[cols.AccountId].Int64()] = struct{}{}
	}
	return blocked, nil
}

func (s *serviceImpl) groupBlockedAccountIDSet(ctx context.Context, accountIDs []int64, appID int64) (map[int64]struct{}, error) {
	blocked := make(map[int64]struct{})
	if len(accountIDs) == 0 {
		return blocked, nil
	}
	accountGroupCols := dao.AccountGroup.Columns()
	relations, err := s.tenantFilter.Apply(ctx, dao.AccountGroup.Ctx(ctx), "").
		Fields(accountGroupCols.AccountId, accountGroupCols.GroupId).
		WhereIn(accountGroupCols.AccountId, accountIDs).
		All()
	if err != nil {
		return nil, err
	}
	groupIDs := make([]int64, 0, len(relations))
	for _, relation := range relations {
		groupIDs = append(groupIDs, relation[accountGroupCols.GroupId].Int64())
	}
	if len(groupIDs) == 0 {
		return blocked, nil
	}
	now := time.Now()
	groupCols := dao.GroupAppBlacklist.Columns()
	groupRows, err := s.tenantFilter.Apply(ctx, dao.GroupAppBlacklist.Ctx(ctx), "").
		Fields(groupCols.GroupId).
		WhereIn(groupCols.GroupId, uniqueInt64s(groupIDs)).
		Where(groupCols.AppId, appID).
		Where("("+groupCols.EffectAt+" IS NULL OR "+groupCols.EffectAt+" <= ?)", now).
		Where("("+groupCols.ExpireAt+" IS NULL OR "+groupCols.ExpireAt+" >= ?)", now).
		All()
	if err != nil {
		return nil, err
	}
	blockedGroups := make(map[int64]struct{}, len(groupRows))
	for _, row := range groupRows {
		blockedGroups[row[groupCols.GroupId].Int64()] = struct{}{}
	}
	for _, relation := range relations {
		if _, ok := blockedGroups[relation[accountGroupCols.GroupId].Int64()]; ok {
			blocked[relation[accountGroupCols.AccountId].Int64()] = struct{}{}
		}
	}
	return blocked, nil
}

func orderPrimaryFirst(accounts []*entity.Account, primaryID int64) []*entity.Account {
	result := make([]*entity.Account, 0, len(accounts))
	for _, account := range accounts {
		if account.Id == primaryID {
			result = append(result, account)
			break
		}
	}
	for _, account := range accounts {
		if account.Id != primaryID {
			result = append(result, account)
		}
	}
	return result
}

func uniqueInt64s(values []int64) []int64 {
	seen := make(map[int64]struct{}, len(values))
	result := make([]int64, 0, len(values))
	for _, value := range values {
		if value <= 0 {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func (s *serviceImpl) hasDelegatedAccess(ctx context.Context, ownerID int64, selectedID int64, appID int64) (bool, error) {
	if ownerID == selectedID {
		return true, nil
	}
	now := time.Now()
	count, err := s.tenantFilter.Apply(ctx, dao.AccountAppRole.Ctx(ctx), "").
		Where(dao.AccountAppRole.Columns().EmpoweredAccountId, ownerID).
		Where(dao.AccountAppRole.Columns().GiveAccountId, selectedID).
		Where(dao.AccountAppRole.Columns().AppId, appID).
		Where("("+dao.AccountAppRole.Columns().ExpireAt+" IS NULL OR "+dao.AccountAppRole.Columns().ExpireAt+" >= ?)", now).
		Count()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *serviceImpl) recordCASLoginWithID(ctx context.Context, accountID int64, choiceAccountID int64, appID int64, loginType string, message string) (int64, error) {
	now := time.Now()
	tenantID, actorID := s.baseOwnedDO(ctx, true)
	data := do.CasLoginLog{
		TenantId:        tenantID,
		AccountId:       accountID,
		ChoiceAccountId: choiceAccountID,
		AppId:           appID,
		LoginTime:       &now,
		Msg:             message,
		LoginType:       loginType,
		CreatedBy:       actorID,
		UpdatedBy:       actorID,
	}
	if r := g.RequestFromCtx(ctx); r != nil {
		ua := useragent.New(r.GetHeader("User-Agent"))
		browserName, browserVersion := ua.Browser()
		data.Ipaddr = r.GetClientIp()
		data.Browser = strings.TrimSpace(browserName + " " + browserVersion)
		data.Os = ua.OS()
		data.Platform = ua.Platform()
		data.Remark = r.GetHeader("User-Agent")
	}
	return dao.CasLoginLog.Ctx(ctx).Data(data).InsertAndGetId()
}
