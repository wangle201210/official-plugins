// This file implements password policy validation, administrator reset, and
// self-service password reset challenge flow using plugin-owned persistence.

package uidentity

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"time"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/entity"
)

const (
	passwordChallengeTTL        = 15 * time.Minute
	passwordChallengeCodePrefix = "pwd_challenge:"
	passwordVerifiedDataPrefix  = "pwd_verified:"
	smsTypePasswordReset        = "cas_pwd_change"
	smsStatusSuccess            = 1
)

type passwordChallengeData struct {
	AccountID int64  `json:"accountId"`
	Number    string `json:"number"`
	Phone     string `json:"phone"`
	Stage     string `json:"stage"`
}

// ResetAccountPassword resets one account password by administrator action.
func (s *serviceImpl) ResetAccountPassword(ctx context.Context, accountID int64, newPassword string) error {
	account, err := s.getAccountByID(ctx, accountID)
	if err != nil {
		return err
	}
	level, err := s.validatePassword(ctx, newPassword)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = s.tenantFilter.Apply(ctx, dao.Account.Ctx(ctx), "").
		Where(dao.Account.Columns().Id, account.Id).
		Data(do.Account{
			PasswordHash:      hashPassword(newPassword),
			PasswordUpdatedAt: &now,
			PassLevel:         level,
			Status:            AccountStatusNormal,
			UpdatedBy:         s.actorID(ctx),
		}).
		Update()
	return err
}

// CreatePasswordChallenge creates one short-lived password reset challenge.
func (s *serviceImpl) CreatePasswordChallenge(ctx context.Context, number string) (*PasswordChallengeOutput, error) {
	accountColumns := dao.Account.Columns()
	var account *entity.Account
	err := s.tenantFilter.Apply(ctx, dao.Account.Ctx(ctx), "").
		Where(accountColumns.Number, number).
		Scan(&account)
	if err != nil {
		return nil, err
	}
	if account == nil {
		return nil, bizerr.NewCode(CodeResourceNotFound)
	}
	challengeID, err := randomToken("pwd")
	if err != nil {
		return nil, err
	}
	payload, err := json.Marshal(passwordChallengeData{
		AccountID: account.Id,
		Number:    account.Number,
		Phone:     account.Phone,
		Stage:     "number",
	})
	if err != nil {
		return nil, err
	}
	expiredAt := time.Now().Add(passwordChallengeTTL)
	tenantID, actorID := s.baseOwnedDO(ctx, true)
	_, err = dao.OauthToken.Ctx(ctx).Data(do.OauthToken{
		TenantId:  tenantID,
		Code:      passwordChallengeCodePrefix + challengeID,
		Data:      string(payload),
		ExpiredAt: &expiredAt,
		CreatedBy: actorID,
		UpdatedBy: actorID,
	}).Insert()
	if err != nil {
		return nil, err
	}
	return &PasswordChallengeOutput{ChallengeID: challengeID, Status: account.Status}, nil
}

// VerifyPasswordChallengePhone verifies phone and SMS code for a challenge.
func (s *serviceImpl) VerifyPasswordChallengePhone(ctx context.Context, challengeID string, phone string, code string) (string, error) {
	token, payload, err := s.passwordChallenge(ctx, passwordChallengeCodePrefix+challengeID)
	if err != nil {
		return "", err
	}
	if payload.Phone != phone {
		return "", bizerr.NewCode(CodePasswordChallengeInvalid)
	}
	if err := s.verifySMSCode(ctx, phone, code); err != nil {
		return "", err
	}
	payload.Stage = "phone"
	content, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	_, err = s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "").
		Where(dao.OauthToken.Columns().Id, token.Id).
		Data(do.OauthToken{
			Code:      passwordVerifiedDataPrefix + challengeID,
			Data:      string(content),
			UpdatedBy: s.actorID(ctx),
		}).
		Update()
	if err != nil {
		return "", err
	}
	return challengeID, nil
}

// ResetPasswordByChallenge consumes a verified challenge and resets password.
func (s *serviceImpl) ResetPasswordByChallenge(ctx context.Context, challengeID string, newPassword string) error {
	token, payload, err := s.passwordChallenge(ctx, passwordVerifiedDataPrefix+challengeID)
	if err != nil {
		return err
	}
	if payload.Stage != "phone" {
		return bizerr.NewCode(CodePasswordChallengeInvalid)
	}
	if err := s.ResetAccountPassword(ctx, payload.AccountID, newPassword); err != nil {
		return err
	}
	_, err = s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "").
		Where(dao.OauthToken.Columns().Id, token.Id).
		Delete()
	return err
}

func (s *serviceImpl) passwordChallenge(ctx context.Context, code string) (*entity.OauthToken, passwordChallengeData, error) {
	tokenColumns := dao.OauthToken.Columns()
	var token *entity.OauthToken
	err := s.tenantFilter.Apply(ctx, dao.OauthToken.Ctx(ctx), "").
		Where(tokenColumns.Code, code).
		Scan(&token)
	if err != nil {
		return nil, passwordChallengeData{}, err
	}
	if token == nil || token.ExpiredAt == nil || token.ExpiredAt.Before(time.Now()) {
		return nil, passwordChallengeData{}, bizerr.NewCode(CodePasswordChallengeInvalid)
	}
	payload := passwordChallengeData{}
	if err := json.Unmarshal([]byte(token.Data), &payload); err != nil {
		return nil, passwordChallengeData{}, bizerr.NewCode(CodePasswordChallengeInvalid)
	}
	return token, payload, nil
}

func (s *serviceImpl) verifySMSCode(ctx context.Context, phone string, code string) error {
	smsColumns := dao.Sms.Columns()
	count, err := s.tenantFilter.Apply(ctx, dao.Sms.Ctx(ctx), "").
		Where(smsColumns.Phone, phone).
		Where(smsColumns.Type, smsTypePasswordReset).
		Where(smsColumns.Content, code).
		Where(smsColumns.Status, smsStatusSuccess).
		Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return bizerr.NewCode(CodeSMSCodeInvalid)
	}
	return nil
}

func (s *serviceImpl) validatePassword(ctx context.Context, password string) (int, error) {
	rule := &entity.PassRule{}
	err := s.tenantFilter.Apply(ctx, dao.PassRule.Ctx(ctx), "").
		Where(dao.PassRule.Columns().Status, 1).
		OrderDesc(dao.PassRule.Columns().Id).
		Scan(rule)
	if err != nil {
		return 0, err
	}
	if rule.Length == 0 {
		rule.Length = 8
	}
	if len(password) < rule.Length {
		return 0, bizerr.NewCode(CodePasswordWeak)
	}
	score := 1
	checks := []struct {
		required int
		match    bool
	}{
		{required: rule.Capital, match: regexp.MustCompile(`[A-Z]`).MatchString(password)},
		{required: rule.Number, match: regexp.MustCompile(`[0-9]`).MatchString(password)},
		{required: rule.Lower, match: regexp.MustCompile(`[a-z]`).MatchString(password)},
		{required: rule.Symbol, match: regexp.MustCompile(`[!@#~$%^&*()+|_.,;:<>{}[\]/\\\-?"]`).MatchString(password)},
	}
	for _, check := range checks {
		if check.match {
			score++
			continue
		}
		if check.required == 1 {
			return 0, bizerr.NewCode(CodePasswordWeak)
		}
	}
	return score, nil
}

func hashPassword(password string) string {
	sum := sha256.Sum256([]byte(password))
	return hex.EncodeToString(sum[:])
}

func randomToken(prefix string) (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(buffer), nil
}
