// This file verifies legacy SMS captcha compatibility helpers without
// depending on database fixtures or external SMS gateways.

package uidentity

import (
	"context"
	"testing"
	"time"

	"github.com/mojocn/base64Captcha"
)

// TestLegacyCommonPassMatchesOldFormula verifies the old common-pass date code.
func TestLegacyCommonPassMatchesOldFormula(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.June, 2, 8, 30, 0, 0, time.Local)
	if !legacyCommonPassMatches("040206", now) {
		t.Fatal("expected old weekday/date common pass to match")
	}
	if legacyCommonPassMatches("030206", now) {
		t.Fatal("expected wrong old common pass to be rejected")
	}
}

// TestVerifyLegacySMSCaptchaAcceptsConfiguredCommonPass verifies the old
// commonPass=true default used by the migrated admin config.
func TestVerifyLegacySMSCaptchaAcceptsConfiguredCommonPass(t *testing.T) {
	t.Parallel()

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, "")}
	now := time.Date(2026, time.June, 2, 8, 30, 0, 0, time.Local)
	if err := service.verifyLegacySMSCaptcha(context.Background(), "040206", "", now); err != nil {
		t.Fatalf("expected common pass to skip captcha store: %v", err)
	}
}

// TestVerifyLegacySMSCaptchaRejectsWhenCommonPassDisabled verifies a plugin
// config override can disable the old bypass while still requiring uuid/code.
func TestVerifyLegacySMSCaptchaRejectsWhenCommonPassDisabled(t *testing.T) {
	t.Parallel()

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, `
legacy:
  app:
    commonPass: false
`)}
	now := time.Date(2026, time.June, 2, 8, 30, 0, 0, time.Local)
	if err := service.verifyLegacySMSCaptcha(context.Background(), "040206", "", now); err == nil {
		t.Fatal("expected disabled common pass without captcha uuid to be rejected")
	}
}

// TestVerifyLegacySMSCaptchaConsumesBase64CaptchaStore preserves the old
// captcha.Verify(uuid, code, true) one-shot behavior.
func TestVerifyLegacySMSCaptchaConsumesBase64CaptchaStore(t *testing.T) {
	t.Parallel()

	service := &serviceImpl{configSvc: newLegacyConfigTestService(t, `
legacy:
  app:
    commonPass: false
`)}
	uuid := "legacy-sms-captcha"
	if err := base64Captcha.DefaultMemStore.Set(uuid, "2468"); err != nil {
		t.Fatalf("set captcha store: %v", err)
	}
	if err := service.verifyLegacySMSCaptcha(context.Background(), "2468", uuid, time.Now()); err != nil {
		t.Fatalf("expected stored captcha to pass: %v", err)
	}
	if err := service.verifyLegacySMSCaptcha(context.Background(), "2468", uuid, time.Now()); err == nil {
		t.Fatal("expected stored captcha to be consumed after first verification")
	}
}
