// identity_login_test.go verifies WeChat code login first-login provisioning and
// idempotent re-login against a gated PostgreSQL database. The tests build the
// identity service with the mock WeChat gateway so login runs without real
// WeChat credentials.

package identity

import (
	"context"
	"testing"
	"time"

	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
	"lina-plugin-sicau-niu/backend/internal/dao"
	tokensvc "lina-plugin-sicau-niu/backend/internal/service/token"
	wechatsvc "lina-plugin-sicau-niu/backend/internal/service/wechat"
)

// newIdentityServiceForTest builds an identity service wired to a mock WeChat
// gateway, a real token issuer and the real college service. When fixedOpenid is
// non-empty the gateway returns it for every login code.
func newIdentityServiceForTest(t *testing.T, fixedOpenid string) Service {
	t.Helper()
	gateway := wechatsvc.New(wechatsvc.Config{Mock: true, MockOpenid: fixedOpenid})
	tokenService, err := tokensvc.New(tokensvc.Config{Secret: "unit-test-secret", TTL: time.Hour})
	if err != nil {
		t.Fatalf("construct token service failed: %v", err)
	}
	return New(gateway, tokenService, collegesvc.New())
}

// TestLoginProvisionsThenReusesPlayer verifies the first login with a given
// openid provisions a new player (IsNewUser=true) and a second login with the
// same openid reuses the same row without creating a duplicate.
func TestLoginProvisionsThenReusesPlayer(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLIdentityDB(t, ctx)

	const fixedOpenid = "openid-login-reuse"
	svc := newIdentityServiceForTest(t, fixedOpenid)

	first, err := svc.Login(ctx, &LoginInput{Code: "wx-code-1"})
	if err != nil {
		t.Fatalf("first login failed: %v", err)
	}
	if !first.IsNewUser {
		t.Fatal("expected first login to provision a new player")
	}
	if first.PlayerID <= 0 {
		t.Fatalf("expected positive player ID, got %d", first.PlayerID)
	}
	if first.Token == "" {
		t.Fatal("expected a non-empty session token on login")
	}

	second, err := svc.Login(ctx, &LoginInput{Code: "wx-code-2"})
	if err != nil {
		t.Fatalf("second login failed: %v", err)
	}
	if second.IsNewUser {
		t.Fatal("expected second login to reuse the existing player")
	}
	if second.PlayerID != first.PlayerID {
		t.Fatalf("expected same player ID on re-login, got %d then %d", first.PlayerID, second.PlayerID)
	}

	count, err := dao.User.Ctx(ctx).Where(dao.User.Columns().Openid, fixedOpenid).Count()
	if err != nil {
		t.Fatalf("count players failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected exactly one player row for openid, got %d", count)
	}
}

// TestLoginRejectsEmptyCode verifies a blank login code returns
// CodeLoginCodeRequired before touching the store. This assertion is pure-logic
// and still runs when the DB harness skips, so it sets up its own service only.
func TestLoginRejectsEmptyCode(t *testing.T) {
	ctx := context.Background()
	svc := newIdentityServiceForTest(t, "openid-empty-code")

	_, err := svc.Login(ctx, &LoginInput{Code: "   "})
	assertBizCode(t, err, CodeLoginCodeRequired.RuntimeCode())
}
