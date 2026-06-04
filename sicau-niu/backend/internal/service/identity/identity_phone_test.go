// identity_phone_test.go verifies phone-authorization binding and the
// one-phone-one-account uniqueness constraint against a gated PostgreSQL
// database.

package identity

import (
	"context"
	"testing"

	"lina-plugin-sicau-niu/backend/internal/dao"
)

// provisionPlayer logs in with a fixed openid and returns the provisioned player
// ID. Each caller uses a distinct openid so players are independent.
func provisionPlayer(t *testing.T, ctx context.Context, openid string) (Service, int64) {
	t.Helper()
	svc := newIdentityServiceForTest(t, openid)
	out, err := svc.Login(ctx, &LoginInput{Code: "code-" + openid})
	if err != nil {
		t.Fatalf("provision player %q failed: %v", openid, err)
	}
	return svc, out.PlayerID
}

// TestBindPhoneEnforcesOnePhoneOneAccount verifies player A binds a phone
// successfully and player B binding the same phone is rejected with
// CodePhoneTaken, while the phone remains attached to player A.
func TestBindPhoneEnforcesOnePhoneOneAccount(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLIdentityDB(t, ctx)

	const sharedPhone = "13800000001"
	svcA, playerA := provisionPlayer(t, ctx, "openid-phone-a")
	_, playerB := provisionPlayer(t, ctx, "openid-phone-b")
	if playerA == playerB {
		t.Fatal("expected two distinct players")
	}

	if err := svcA.BindPhone(ctx, playerA, &BindPhoneInput{
		PhoneOverride:     sharedPhone,
		DeviceFingerprint: "device-a",
	}); err != nil {
		t.Fatalf("bind phone to player A failed: %v", err)
	}

	// Player B claims the same phone: rejected by the uniqueness constraint.
	err := svcA.BindPhone(ctx, playerB, &BindPhoneInput{PhoneOverride: sharedPhone})
	assertBizCode(t, err, CodePhoneTaken.RuntimeCode())

	// The phone must still belong to player A only.
	ownerCount, err := dao.User.Ctx(ctx).
		Where(dao.User.Columns().Phone, sharedPhone).
		Count()
	if err != nil {
		t.Fatalf("count phone owners failed: %v", err)
	}
	if ownerCount != 1 {
		t.Fatalf("expected exactly one owner of the phone, got %d", ownerCount)
	}

	var fingerprint string
	value, err := dao.User.Ctx(ctx).
		Where(dao.User.Columns().Id, playerA).
		Fields(dao.User.Columns().DeviceFingerprint).
		Value()
	if err != nil {
		t.Fatalf("read player A fingerprint failed: %v", err)
	}
	fingerprint = value.String()
	if fingerprint != "device-a" {
		t.Fatalf("expected recorded device fingerprint device-a, got %q", fingerprint)
	}
}

// TestBindPhoneAllowsSamePlayerRebind verifies binding the same phone again for
// the same player is idempotent and not treated as taken by another account.
func TestBindPhoneAllowsSamePlayerRebind(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLIdentityDB(t, ctx)

	const phone = "13800000002"
	svc, player := provisionPlayer(t, ctx, "openid-phone-rebind")

	if err := svc.BindPhone(ctx, player, &BindPhoneInput{PhoneOverride: phone}); err != nil {
		t.Fatalf("first bind failed: %v", err)
	}
	if err := svc.BindPhone(ctx, player, &BindPhoneInput{PhoneOverride: phone}); err != nil {
		t.Fatalf("re-bind same phone for same player failed: %v", err)
	}
}
