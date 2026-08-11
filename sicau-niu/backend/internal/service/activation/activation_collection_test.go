// activation_collection_test.go covers the DB-gated personal card collection: the
// player sees the complete bounded catalog with isolated ownership state, the
// optional category filter, invalid category rejection and locked-card redaction.

package activation

import (
	"context"
	"testing"
	"time"

	"lina-plugin-sicau-niu/backend/internal/model/do"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
)

// stageNiuWithCard inserts a cattle plus its main card and returns the cattle ID.
func stageNiuWithCard(t *testing.T, ctx context.Context, code, category, title string, lat float64) int64 {
	t.Helper()
	past := time.Now().Add(-time.Hour)
	niuID := insertNiuRow(t, ctx, do.Niu{
		Code: code, NiuType: cattlesvc.NiuTypeCommon.String(),
		Lat: lat, Lng: 103.0,
		OnlineAt: &past,
		Status:   cattlesvc.NiuStatusInactive.String(),
	})
	insertCardRow(t, ctx, do.Card{NiuId: niuID, Category: category, Title: title, Content: "c"})
	return niuID
}

// TestCollectionSelfIsolation verifies the catalog exposes only the requesting
// player's ownership and redacts card-face content for cards they do not own.
func TestCollectionSelfIsolation(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()

	mineNiu := stageNiuWithCard(t, ctx, "NIU-COL-1", "spirit", "我的卡", 30.0)
	otherNiu := stageNiuWithCard(t, ctx, "NIU-COL-2", "event", "他的卡", 30.0002)
	me := insertUserRow(t, ctx, do.User{Openid: "openid-me"})
	other := insertUserRow(t, ctx, do.User{Openid: "openid-other"})

	myActivation, err := svc.Activate(ctx, me, &ActivateInput{Lat: 30.0, Lng: 103.0})
	if err != nil {
		t.Fatalf("my activation failed: %v", err)
	}
	if myActivation.NiuId != mineNiu {
		t.Fatalf("expected my activation to match niu %d, got %d", mineNiu, myActivation.NiuId)
	}
	otherActivation, err := svc.Activate(ctx, other, &ActivateInput{Lat: 30.0002, Lng: 103.0})
	if err != nil {
		t.Fatalf("other activation failed: %v", err)
	}
	if otherActivation.NiuId != otherNiu {
		t.Fatalf("expected other activation to match niu %d, got %d", otherNiu, otherActivation.NiuId)
	}

	items, err := svc.Collection(ctx, me, "")
	if err != nil {
		t.Fatalf("collection failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected the complete two-card catalog, got %+v", items)
	}
	if items[0].NiuId != mineNiu || !items[0].Owned || items[0].Title != "我的卡" {
		t.Fatalf("expected my card to be owned and visible, got %+v", items[0])
	}
	if items[1].NiuId != otherNiu || items[1].Owned {
		t.Fatalf("expected the other player's card to stay locked, got %+v", items[1])
	}
	if items[1].Title != "" || items[1].Content != "" || items[1].ImagePath != "" {
		t.Fatalf("expected locked card-face content to be redacted, got %+v", items[1])
	}
}

// TestCollectionCategoryFilter verifies the category filter narrows the result to
// catalog cards of that category while preserving ownership state.
func TestCollectionCategoryFilter(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()

	spiritNiu := stageNiuWithCard(t, ctx, "NIU-COL-S", "spirit", "精神卡", 30.0)
	eventNiu := stageNiuWithCard(t, ctx, "NIU-COL-E", "event", "事件卡", 30.0002)
	me := insertUserRow(t, ctx, do.User{Openid: "openid-filter"})

	// Activate both cattle on distinct days by pre-seeding the event activation so
	// the daily limit does not block staging two collected cards.
	spiritActivation, err := svc.Activate(ctx, me, &ActivateInput{Lat: 30.0, Lng: 103.0})
	if err != nil {
		t.Fatalf("spirit activation failed: %v", err)
	}
	if spiritActivation.NiuId != spiritNiu {
		t.Fatalf("expected spirit activation to match niu %d, got %d", spiritNiu, spiritActivation.NiuId)
	}
	if _, err := daoInsertActivation(ctx, me, eventNiu, "2000-01-02", 0, 1); err != nil {
		t.Fatalf("seed event activation failed: %v", err)
	}

	items, err := svc.Collection(ctx, me, "event")
	if err != nil {
		t.Fatalf("filtered collection failed: %v", err)
	}
	if len(items) != 1 || items[0].Category != "event" || !items[0].Owned {
		t.Fatalf("expected only event cards, got %+v", items)
	}
}

// TestCollectionInvalidCategoryRejected verifies an unknown category filter is
// rejected with CodeCategoryInvalid.
func TestCollectionInvalidCategoryRejected(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	me := insertUserRow(t, ctx, do.User{Openid: "openid-badcat"})

	_, err := svc.Collection(ctx, me, "not-a-category")
	assertBizCode(t, err, CodeCategoryInvalid.RuntimeCode())
}

// TestCollectionLockedWhenNoActivation verifies a player with no activations gets
// the catalog with every card locked and card-face content redacted.
func TestCollectionLockedWhenNoActivation(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	niuID := stageNiuWithCard(t, ctx, "NIU-COL-LOCKED", "research", "保密卡面", 30.0)
	me := insertUserRow(t, ctx, do.User{Openid: "openid-empty"})

	items, err := svc.Collection(ctx, me, "")
	if err != nil {
		t.Fatalf("locked collection failed: %v", err)
	}
	if len(items) != 1 || items[0].NiuId != niuID || items[0].Owned {
		t.Fatalf("expected one locked catalog card, got %+v", items)
	}
	if items[0].Category != "research" || items[0].Title != "" || items[0].Content != "" || items[0].ImagePath != "" {
		t.Fatalf("expected category-only locked projection, got %+v", items[0])
	}
}
