// This file implements plugin-local SMS verification code sending and rate
// limiting for legacy CAS runtime flows.

package uidentity

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/mojocn/base64Captcha"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
)

const (
	smsLocalRateWindow        = time.Hour
	smsLocalMaxCount          = 5
	smsCodeMin                = 100000
	smsCodeMax                = 999999
	configKeyLegacyCommonPass = "legacy.app.commonPass"
)

// SendSMSCode records one bounded plugin-local SMS verification code.
func (s *serviceImpl) SendSMSCode(ctx context.Context, in SMSSendInput) (*SMSSendOutput, error) {
	if err := s.verifyLegacySMSCaptcha(ctx, in.Code, in.UUID, time.Now()); err != nil {
		return nil, err
	}
	smsType := strings.TrimSpace(in.Type)
	if !validSMSType(smsType) {
		return nil, bizerr.NewCode(CodeSMSTypeInvalid)
	}
	phone := strings.TrimSpace(in.Phone)
	if phone == "" {
		return nil, bizerr.NewCode(CodeImportInvalid)
	}
	smsCols := dao.Sms.Columns()
	sentCount, err := dao.Sms.Ctx(ctx).
		Where(smsCols.Phone, phone).
		Where(smsCols.Type, smsType).
		Where(smsCols.Status, smsStatusSuccess).
		Where(smsCols.CreatedAt+" >= ?", time.Now().Add(-smsLocalRateWindow)).
		Count()
	if err != nil {
		return nil, err
	}
	if sentCount >= smsLocalMaxCount {
		return nil, bizerr.NewCode(CodeSMSRateLimited)
	}
	actorID := s.actorID(ctx)
	code := fmt.Sprintf("%06d", smsCodeMin+rand.Intn(smsCodeMax-smsCodeMin+1))
	record := do.Sms{
		Phone:    phone,
		Type:     smsType,
		Content:  code,
		Status:   smsStatusSuccess,
		RespMsg:  "recorded by plugin-local SMS sender",
		CreateBy: actorID,
		UpdateBy: actorID,
	}
	gateway, err := s.legacySMSGatewayConfig(ctx)
	if err != nil {
		return nil, err
	}
	var gatewayErr error
	if gateway.enabled() {
		body, sendErr := sendLegacySMSGateway(ctx, gateway, phone, gateway.content(code))
		record.RespMsg = body
		if sendErr != nil {
			// Keep the old behavior: persist the failed delivery record, then
			// surface the gateway failure to the caller.
			record.Status = smsStatusFailed
			gatewayErr = sendErr
		}
	}
	id, err := dao.Sms.Ctx(ctx).Data(record).InsertAndGetId()
	if err != nil {
		return nil, err
	}
	if gatewayErr != nil {
		return nil, gatewayErr
	}
	return &SMSSendOutput{ID: id}, nil
}

func (s *serviceImpl) verifyLegacySMSCaptcha(ctx context.Context, code string, uuid string, now time.Time) error {
	code = strings.TrimSpace(code)
	uuid = strings.TrimSpace(uuid)
	commonPass, err := s.legacyCommonPassEnabled(ctx)
	if err != nil {
		return err
	}
	if commonPass && legacyCommonPassMatches(code, now) {
		return nil
	}
	if uuid == "" || code == "" {
		return bizerr.NewCode(CodeSMSCaptchaInvalid)
	}
	if !base64Captcha.DefaultMemStore.Verify(uuid, code, true) {
		return bizerr.NewCode(CodeSMSCaptchaInvalid)
	}
	return nil
}

func (s *serviceImpl) legacyCommonPassEnabled(ctx context.Context) (bool, error) {
	if s == nil || s.configSvc == nil {
		return true, nil
	}
	return s.configSvc.Bool(ctx, configKeyLegacyCommonPass, true)
}

func legacyCommonPassMatches(code string, now time.Time) bool {
	if now.IsZero() {
		now = time.Now()
	}
	expected := fmt.Sprintf("%02d%s", int(now.Weekday())+2, now.Format("0201"))
	return strings.TrimSpace(code) == expected
}

func validSMSType(smsType string) bool {
	switch smsType {
	case smsTypeCasLogin, smsTypeCasActive, smsTypeCasBind, smsTypePasswordReset:
		return true
	default:
		return false
	}
}
