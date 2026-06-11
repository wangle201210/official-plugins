// This file verifies the SMS verification contract: common-pass bypass, the
// 5-minute expiry window, and one-time consumption of codes.

package uidentity

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "lina-core/pkg/dbdriver"
	"lina-plugin-linapro-uidentity-cas/backend/internal/dao"
	"lina-plugin-linapro-uidentity-cas/backend/internal/model/do"
)

func insertSMSTestCode(t *testing.T, ctx context.Context, phone string, smsType string, code string, createdAt time.Time) int64 {
	t.Helper()
	id, err := dao.Sms.Ctx(ctx).Data(do.Sms{
		Phone:   phone,
		Type:    smsType,
		Content: code,
		Status:  smsStatusSuccess,
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert sms code: %v", err)
	}
	// gf's soft-time maintainer overrides created_at on Insert, so backdate
	// the row explicitly for expiry-window assertions.
	if _, err := dao.Sms.Ctx(ctx).
		Where(dao.Sms.Columns().Id, id).
		Data(map[string]any{dao.Sms.Columns().CreatedAt: createdAt}).
		Update(); err != nil {
		t.Fatalf("backdate sms code: %v", err)
	}
	return id
}

func cleanupSMSTestRows(t *testing.T, ctx context.Context, phone string) {
	t.Helper()
	if _, err := dao.Sms.Ctx(ctx).Unscoped().Where(dao.Sms.Columns().Phone, phone).Delete(); err != nil {
		t.Fatalf("cleanup sms rows: %v", err)
	}
}

func TestVerifySMSCodeEnforcesExpiryAndOneTimeConsumption(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.Sms.Table())

	phone := fmt.Sprintf("139%08d", time.Now().UnixNano()%100000000)
	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, `
legacy:
  app:
    commonPass: false
`)}

	cleanupSMSTestRows(t, ctx, phone)
	t.Cleanup(func() { cleanupSMSTestRows(t, ctx, phone) })

	// A fresh code verifies successfully...
	insertSMSTestCode(t, ctx, phone, smsTypeCasLogin, "123456", time.Now())
	if err := service.verifySMSCode(ctx, phone, "123456", smsTypeCasLogin); err != nil {
		t.Fatalf("expected fresh code to verify: %v", err)
	}
	// ...and is consumed one-time, so a replay is rejected as expired.
	if err := service.verifySMSCode(ctx, phone, "123456", smsTypeCasLogin); err == nil {
		t.Fatal("expected consumed code to be rejected on replay")
	}

	// An expired code (older than the 5-minute window) is rejected.
	cleanupSMSTestRows(t, ctx, phone)
	insertSMSTestCode(t, ctx, phone, smsTypeCasLogin, "222222", time.Now().Add(-smsCodeTTL-time.Minute))
	if err := service.verifySMSCode(ctx, phone, "222222", smsTypeCasLogin); err == nil {
		t.Fatal("expected expired code to be rejected")
	}

	// A wrong attempt burns the outstanding code; a later correct guess fails.
	cleanupSMSTestRows(t, ctx, phone)
	insertSMSTestCode(t, ctx, phone, smsTypeCasLogin, "333333", time.Now())
	if err := service.verifySMSCode(ctx, phone, "999999", smsTypeCasLogin); err == nil {
		t.Fatal("expected wrong code to fail")
	}
	if err := service.verifySMSCode(ctx, phone, "333333", smsTypeCasLogin); err == nil {
		t.Fatal("expected burned code to be rejected after a wrong attempt")
	}
}

func TestVerifySMSCodeAcceptsConfiguredCommonPass(t *testing.T) {
	ctx := context.Background()
	configureUIdentityTestDB(t, ctx, dao.Sms.Table())

	phone := fmt.Sprintf("138%08d", time.Now().UnixNano()%100000000)
	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, "")}
	commonPass := fmt.Sprintf("%02d%s", int(time.Now().Weekday())+2, time.Now().Format("0201"))
	if err := service.verifySMSCode(ctx, phone, commonPass, smsTypeCasLogin); err != nil {
		t.Fatalf("expected common pass to bypass SMS verification: %v", err)
	}
}
