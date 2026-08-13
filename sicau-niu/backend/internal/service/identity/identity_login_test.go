// identity_login_test.go verifies WeChat code login first-login provisioning and
// idempotent re-login against a gated PostgreSQL database. The tests build the
// identity service with the mock WeChat gateway so login runs without real
// WeChat credentials.

package identity

import (
	"context"
	"sync"
	"testing"
	"time"

	"lina-plugin-sicau-niu/backend/internal/dao"
	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
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

// TestLoginConcurrentFirstUseReusesOnePlayer verifies simultaneous first logins
// for one WeChat openid all resolve to the same account instead of exposing the
// backing unique-key conflict to losing callers.
func TestLoginConcurrentFirstUseReusesOnePlayer(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLIdentityDB(t, ctx)

	const fixedOpenid = "openid-login-concurrent"
	svc := newIdentityServiceForTest(t, fixedOpenid)
	const workers = 16

	start := make(chan struct{})
	results := make(chan *LoginOutput, workers)
	errs := make(chan error, workers)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(worker int) {
			defer wg.Done()
			<-start
			out, err := svc.Login(ctx, &LoginInput{Code: "wx-concurrent"})
			results <- out
			errs <- err
		}(i)
	}
	close(start)
	wg.Wait()
	close(results)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent login failed: %v", err)
		}
	}
	var playerID int64
	newCount := 0
	for out := range results {
		if out == nil || out.PlayerID <= 0 || out.Token == "" {
			t.Fatalf("invalid concurrent login output: %+v", out)
		}
		if playerID == 0 {
			playerID = out.PlayerID
		}
		if out.PlayerID != playerID {
			t.Fatalf("concurrent logins resolved different players: %d and %d", playerID, out.PlayerID)
		}
		if out.IsNewUser {
			newCount++
		}
	}
	if newCount != 1 {
		t.Fatalf("expected exactly one caller to provision the player, got %d", newCount)
	}
	count, err := dao.User.Ctx(ctx).Where(dao.User.Columns().Openid, fixedOpenid).Count()
	if err != nil || count != 1 {
		t.Fatalf("expected one player row, count=%d err=%v", count, err)
	}
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
