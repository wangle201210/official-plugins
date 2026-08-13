// activation_poster_test.go covers the DB-gated activation poster composition: the
// poster fields for an activated cattle (nickname, identity, code, order, quote,
// badge) and the rejection when the player has not activated the requested cattle.

package activation

import (
	"context"
	"testing"
	"time"

	"lina-plugin-sicau-niu/backend/internal/model/do"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
	identitysvc "lina-plugin-sicau-niu/backend/internal/service/identity"
)

// newPosterServiceForTest builds an activation service with a stub identity
// service returning the given profile, plus the basic poster renderer.
func newPosterServiceForTest(profile *identitysvc.ProfileOutput) *serviceImpl {
	return &serviceImpl{
		identitySvc:    &fakeIdentityService{profile: profile},
		posterRenderer: NewBasicPosterRenderer(),
		lbsThreshold:   50,
		campusBadge:    "TEST-BADGE",
	}
}

// TestPosterReturnsCompositionData verifies the poster returns the player profile,
// cattle code, arrival order, a random enabled quote and the campus badge for an
// activated cattle.
func TestPosterReturnsCompositionData(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newPosterServiceForTest(&identitysvc.ProfileOutput{
		Nickname:     "川农牛同学",
		IdentityType: "student",
	})

	past := time.Now().Add(-time.Hour)
	niuID := insertNiuRow(t, ctx, do.Niu{
		Code: "NIU-POSTER-1", NiuType: cattlesvc.NiuTypeCommon.String(),
		Lat: 30.0, Lng: 103.0,
		OnlineAt: &past,
		Status:   cattlesvc.NiuStatusInactive.String(),
	})
	insertQuoteRow(t, ctx, do.Quote{Content: "任重道远", Enabled: 1})
	me := insertUserRow(t, ctx, do.User{Openid: "openid-poster"})

	if _, err := svc.Activate(ctx, me, &ActivateInput{RequestID: "activation-poster", Lat: 30.0, Lng: 103.0}); err != nil {
		t.Fatalf("activation failed: %v", err)
	}

	out, err := svc.Poster(ctx, me, niuID)
	if err != nil {
		t.Fatalf("poster failed: %v", err)
	}
	if out.Nickname != "川农牛同学" || out.IdentityType != "student" {
		t.Fatalf("expected profile fields, got %+v", out)
	}
	if out.NiuCode != "NIU-POSTER-1" || out.OrderNo != 1 {
		t.Fatalf("expected code/order, got code=%q order=%d", out.NiuCode, out.OrderNo)
	}
	if out.Quote != "任重道远" {
		t.Fatalf("expected the only enabled quote, got %q", out.Quote)
	}
	if out.CampusBadge != "TEST-BADGE" {
		t.Fatalf("expected campus badge, got %q", out.CampusBadge)
	}
}

// TestPosterRejectedWhenNotActivated verifies requesting a poster for a cattle the
// player has not activated is rejected with CodeActivationNotFound.
func TestPosterRejectedWhenNotActivated(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newPosterServiceForTest(nil)

	past := time.Now().Add(-time.Hour)
	niuID := insertNiuRow(t, ctx, do.Niu{
		Code: "NIU-POSTER-2", NiuType: cattlesvc.NiuTypeCommon.String(),
		Lat: 30.0, Lng: 103.0,
		OnlineAt: &past,
		Status:   cattlesvc.NiuStatusInactive.String(),
	})
	me := insertUserRow(t, ctx, do.User{Openid: "openid-noposter"})

	_, err := svc.Poster(ctx, me, niuID)
	assertBizCode(t, err, CodeActivationNotFound.RuntimeCode())
}
