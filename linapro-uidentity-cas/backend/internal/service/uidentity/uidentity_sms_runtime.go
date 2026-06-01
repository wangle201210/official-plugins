// This file implements plugin-local SMS verification code sending and rate
// limiting for legacy CAS runtime flows.

package uidentity

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"lina-core/pkg/bizerr"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
)

const (
	smsLocalRateWindow = time.Hour
	smsLocalMaxCount   = 5
	smsCodeMin         = 100000
	smsCodeMax         = 999999
)

// SendSMSCode records one bounded plugin-local SMS verification code.
func (s *serviceImpl) SendSMSCode(ctx context.Context, in SMSSendInput) (*SMSSendOutput, error) {
	smsType := strings.TrimSpace(in.Type)
	if !validSMSType(smsType) {
		return nil, bizerr.NewCode(CodeSMSTypeInvalid)
	}
	phone := strings.TrimSpace(in.Phone)
	if phone == "" {
		return nil, bizerr.NewCode(CodeImportInvalid)
	}
	smsCols := dao.Sms.Columns()
	sentCount, err := s.tenantFilter.Apply(ctx, dao.Sms.Ctx(ctx), "").
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
	tenantID, actorID := s.baseOwnedDO(ctx, true)
	code := fmt.Sprintf("%06d", smsCodeMin+rand.Intn(smsCodeMax-smsCodeMin+1))
	id, err := dao.Sms.Ctx(ctx).Data(do.Sms{
		TenantId:  tenantID,
		Phone:     phone,
		Type:      smsType,
		Content:   code,
		Status:    smsStatusSuccess,
		RespMsg:   "recorded by plugin-local SMS sender",
		CreatedBy: actorID,
		UpdatedBy: actorID,
	}).InsertAndGetId()
	if err != nil {
		return nil, err
	}
	return &SMSSendOutput{ID: id}, nil
}

func validSMSType(smsType string) bool {
	switch smsType {
	case smsTypeCasLogin, smsTypeCasActive, smsTypeCasBind, smsTypePasswordReset:
		return true
	default:
		return false
	}
}
