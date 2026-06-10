// This file implements password policy validation, administrator reset, and
// self-service password reset challenge flow using plugin-owned persistence.

package uidentity

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"
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
	passwordFailureCodePrefix   = "cas:pwd:errnum:"
	passwordFailureTTL          = 30 * time.Minute
	passwordFailureLimit        = 10
	smsTypePasswordReset        = "pwd_change"
	smsTypeCasLogin             = "login"
	smsTypeCasActive            = "active"
	smsTypeCasBind              = "bind"
	smsStatusSuccess            = 1
	smsStatusFailed             = 2
)

type passwordChallengeData struct {
	AccountID int64  `json:"accountId"`
	Number    string `json:"number"`
	Phone     string `json:"phone"`
	Stage     string `json:"stage"`
}

type passwordFailureData struct {
	Number string `json:"number"`
	Count  int    `json:"count"`
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
	if err := s.syncLegacyLDAPPassword(ctx, account.Number, newPassword); err != nil {
		return err
	}
	return s.updateAccountWithAudit(ctx, account.Id, do.Account{
		PassLevel: level,
		Status:    AccountStatusNormal,
		UpdateBy:  s.actorID(ctx),
	})
}

// UnlockPasswordFailures clears short-lived password failure counters.
func (s *serviceImpl) UnlockPasswordFailures(ctx context.Context, numbers []string) ([]string, error) {
	cleaned := uniqueNonEmptyStrings(numbers, 0)
	if len(cleaned) == 0 {
		return nil, bizerr.NewCode(CodePasswordUnlockNumbersRequired)
	}
	if len(cleaned) > maxDeleteIDs {
		return nil, bizerr.NewCode(CodePasswordUnlockNumbersTooMany, bizerr.P("limit", maxDeleteIDs))
	}
	accountColumns := dao.Account.Columns()
	var rows []struct {
		Number string `json:"number"`
	}
	if err := dao.Account.Ctx(ctx).
		Fields(accountColumns.Number).
		WhereIn(accountColumns.Number, cleaned).
		Scan(&rows); err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []string{}, nil
	}
	visible := make([]string, 0, len(rows))
	for _, row := range rows {
		visible = append(visible, row.Number)
	}
	_, err := dao.Oauth2Token.Ctx(ctx).
		WhereIn(dao.Oauth2Token.Columns().Code, passwordFailureCodes(visible)).
		Delete()
	if err != nil {
		return nil, err
	}
	return visible, nil
}

// CreatePasswordChallenge creates one short-lived password reset challenge.
func (s *serviceImpl) CreatePasswordChallenge(ctx context.Context, number string) (*PasswordChallengeOutput, error) {
	accountColumns := dao.Account.Columns()
	var account *entity.Account
	err := dao.Account.Ctx(ctx).
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
	_, err = dao.Oauth2Token.Ctx(ctx).Data(do.Oauth2Token{
		Code:      passwordChallengeCodePrefix + challengeID,
		Data:      string(payload),
		ExpiredAt: expiredAt.UnixMilli(),
	}).Insert()
	if err != nil {
		return nil, err
	}
	return &PasswordChallengeOutput{ChallengeID: challengeID, Status: int(account.Status)}, nil
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
	if err := s.verifySMSCode(ctx, phone, code, smsTypePasswordReset); err != nil {
		return "", err
	}
	payload.Stage = "phone"
	content, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	_, err = dao.Oauth2Token.Ctx(ctx).
		Where(dao.Oauth2Token.Columns().Id, token.Id).
		Data(do.Oauth2Token{
			Code: passwordVerifiedDataPrefix + challengeID,
			Data: string(content),
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
	_, err = dao.Oauth2Token.Ctx(ctx).
		Where(dao.Oauth2Token.Columns().Id, token.Id).
		Delete()
	return err
}

func (s *serviceImpl) passwordChallenge(ctx context.Context, code string) (*entity.Oauth2Token, passwordChallengeData, error) {
	tokenColumns := dao.Oauth2Token.Columns()
	var token *entity.Oauth2Token
	err := dao.Oauth2Token.Ctx(ctx).
		Where(tokenColumns.Code, code).
		Scan(&token)
	if err != nil {
		return nil, passwordChallengeData{}, err
	}
	if runtimeTokenExpired(token.ExpiredAt, time.Now()) {
		return nil, passwordChallengeData{}, bizerr.NewCode(CodePasswordChallengeInvalid)
	}
	payload := passwordChallengeData{}
	if err := json.Unmarshal([]byte(token.Data), &payload); err != nil {
		return nil, passwordChallengeData{}, bizerr.NewCode(CodePasswordChallengeInvalid)
	}
	return token, payload, nil
}

func (s *serviceImpl) verifySMSCode(ctx context.Context, phone string, code string, smsType string) error {
	smsColumns := dao.Sms.Columns()
	count, err := dao.Sms.Ctx(ctx).
		Where(smsColumns.Phone, phone).
		Where(smsColumns.Type, smsType).
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

func (s *serviceImpl) verifyAccountPassword(ctx context.Context, account *entity.Account, password string) error {
	if account == nil {
		return bizerr.NewCode(CodeResourceNotFound)
	}
	failureCount, err := s.passwordFailureCount(ctx, account.Number)
	if err != nil {
		return err
	}
	if failureCount >= passwordFailureLimit {
		return bizerr.NewCode(CodePasswordFailuresLocked)
	}
	if !legacyCommonPassMatches(password, time.Now()) {
		ok, err := s.verifyLegacyLDAPPassword(ctx, account.Number, password)
		if err != nil {
			return err
		}
		if !ok {
			if err := s.recordPasswordFailure(ctx, account.Number, failureCount+1); err != nil {
				return err
			}
			return bizerr.NewCode(CodeInvalidCredentials)
		}
	}
	return s.clearPasswordFailure(ctx, account.Number)
}

func (s *serviceImpl) passwordFailureCount(ctx context.Context, number string) (int, error) {
	token, data, err := s.passwordFailureToken(ctx, number)
	if err != nil || token == nil {
		return 0, err
	}
	return data.Count, nil
}

func (s *serviceImpl) recordPasswordFailure(ctx context.Context, number string, count int) error {
	number = strings.TrimSpace(number)
	if number == "" {
		return nil
	}
	if count < 1 {
		count = 1
	}
	if count > passwordFailureLimit {
		count = passwordFailureLimit
	}
	content, err := json.Marshal(passwordFailureData{Number: number, Count: count})
	if err != nil {
		return err
	}
	expiredAt := time.Now().Add(passwordFailureTTL)
	token, _, err := s.passwordFailureToken(ctx, number)
	if err != nil {
		return err
	}
	if token != nil {
		_, err = dao.Oauth2Token.Ctx(ctx).
			Where(dao.Oauth2Token.Columns().Id, token.Id).
			Data(do.Oauth2Token{
				Data:      string(content),
				ExpiredAt: expiredAt.UnixMilli(),
			}).
			Update()
		return err
	}
	_, err = dao.Oauth2Token.Ctx(ctx).Data(do.Oauth2Token{
		Code:      passwordFailureCode(number),
		Data:      string(content),
		ExpiredAt: expiredAt.UnixMilli(),
	}).Insert()
	if err != nil {
		existing, _, getErr := s.passwordFailureToken(ctx, number)
		if getErr != nil {
			return getErr
		}
		if existing != nil {
			_, updateErr := dao.Oauth2Token.Ctx(ctx).
				Where(dao.Oauth2Token.Columns().Id, existing.Id).
				Data(do.Oauth2Token{
					Data:      string(content),
					ExpiredAt: expiredAt.UnixMilli(),
				}).
				Update()
			return updateErr
		}
	}
	return err
}

func (s *serviceImpl) clearPasswordFailure(ctx context.Context, number string) error {
	number = strings.TrimSpace(number)
	if number == "" {
		return nil
	}
	_, err := dao.Oauth2Token.Ctx(ctx).
		Where(dao.Oauth2Token.Columns().Code, passwordFailureCode(number)).
		Delete()
	return err
}

func (s *serviceImpl) passwordFailureToken(ctx context.Context, number string) (*entity.Oauth2Token, passwordFailureData, error) {
	number = strings.TrimSpace(number)
	if number == "" {
		return nil, passwordFailureData{}, nil
	}
	var token *entity.Oauth2Token
	err := dao.Oauth2Token.Ctx(ctx).
		Where(dao.Oauth2Token.Columns().Code, passwordFailureCode(number)).
		Scan(&token)
	if err != nil {
		return nil, passwordFailureData{}, err
	}
	if token == nil {
		return nil, passwordFailureData{}, nil
	}
	if runtimeTokenExpired(token.ExpiredAt, time.Now()) {
		if err := s.clearPasswordFailure(ctx, number); err != nil {
			return nil, passwordFailureData{}, err
		}
		return nil, passwordFailureData{}, nil
	}
	data := passwordFailureData{}
	if err := json.Unmarshal([]byte(token.Data), &data); err != nil {
		_ = s.clearPasswordFailure(ctx, number)
		return nil, passwordFailureData{}, nil
	}
	return token, data, nil
}

func (s *serviceImpl) validatePassword(ctx context.Context, password string) (int, error) {
	rule := &entity.PassRuler{}
	err := dao.PassRuler.Ctx(ctx).
		Where(dao.PassRuler.Columns().Status, 1).
		OrderDesc(dao.PassRuler.Columns().Id).
		Scan(rule)
	if err != nil {
		return 0, err
	}
	if rule.Length == 0 {
		rule.Length = 8
	}
	if len(password) < int(rule.Length) {
		return 0, bizerr.NewCode(CodePasswordWeak)
	}
	score := 1
	checks := []struct {
		required int
		match    bool
	}{
		{required: int(rule.Capital), match: regexp.MustCompile(`[A-Z]`).MatchString(password)},
		{required: int(rule.Number), match: regexp.MustCompile(`[0-9]`).MatchString(password)},
		{required: int(rule.Lower), match: regexp.MustCompile(`[a-z]`).MatchString(password)},
		{required: int(rule.Symbol), match: regexp.MustCompile(`[!@#~$%^&*()+|_.,;:<>{}[\]/\\\-?"]`).MatchString(password)},
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

func randomToken(prefix string) (string, error) {
	buffer := make([]byte, 24)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return prefix + "_" + hex.EncodeToString(buffer), nil
}

func uniqueNonEmptyStrings(values []string, limit int) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
		if limit > 0 && len(result) == limit {
			break
		}
	}
	return result
}

func passwordFailureCode(number string) string {
	return passwordFailureCodePrefix + strings.TrimSpace(number)
}

func passwordFailureCodes(numbers []string) []string {
	codes := make([]string, 0, len(numbers))
	for _, number := range numbers {
		if trimmed := strings.TrimSpace(number); trimmed != "" {
			codes = append(codes, passwordFailureCode(trimmed))
		}
	}
	return codes
}
