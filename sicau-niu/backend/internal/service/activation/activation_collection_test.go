// activation_collection_test.go covers the DB-gated personal card collection: the
// player only sees cards for cattle they themselves activated (self-isolation),
// the optional category filter, an invalid category rejection and the empty
// collection for a player who has not activated any cattle.

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

// TestCollectionSelfIsolation verifies the collection only contains cards for the
// cattle the requesting player activated, never another player's activations.
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
	otherActivation, err := svc.Activate(ctx, other, &ActivateInput{Lat: 30.0, Lng: 103.0})
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
	if len(items) != 1 || items[0].NiuId != mineNiu {
		t.Fatalf("expected only my activated card, got %+v", items)
	}
}

// TestCollectionCategoryFilter verifies the category filter narrows the result to
// cards of that category among the player's activations.
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
	if len(items) != 1 || items[0].Category != "event" {
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

// TestCollectionEmptyWhenNoActivation verifies a player with no activations gets an
// empty, non-nil collection.
func TestCollectionEmptyWhenNoActivation(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLActivationDB(t, ctx)
	svc := newActivationServiceForTest()
	me := insertUserRow(t, ctx, do.User{Openid: "openid-empty"})

	items, err := svc.Collection(ctx, me, "")
	if err != nil {
		t.Fatalf("empty collection failed: %v", err)
	}
	if items == nil || len(items) != 0 {
		t.Fatalf("expected empty non-nil collection, got %+v", items)
	}
}
