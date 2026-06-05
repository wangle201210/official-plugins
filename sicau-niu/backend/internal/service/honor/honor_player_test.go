// honor_player_test.go covers the read-only player honor unlock computation
// end-to-end against the database, exercising every unlock rule (participation,
// feed_count, activation_count, category_complete, full_complete) and confirming
// the computation never writes a grant. The tests are DB-gated on
// LINA_TEST_PGSQL_LINK and self-contained.

package honor

import (
	"context"
	"testing"

	"lina-plugin-sicau-niu/backend/internal/dao"
	cardsvc "lina-plugin-sicau-niu/backend/internal/service/card"
)

// unlockedByCode indexes the player honor list by honor code for assertions.
func unlockedByCode(items []*PlayerHonorItem) map[string]bool {
	out := make(map[string]bool, len(items))
	for _, item := range items {
		out[item.Code] = item.Unlocked
	}
	return out
}

// TestPlayerHonorsUnlockRules seeds one honor of each unlock rule plus the
// player's feeding, activation and collection state, then verifies each rule's
// unlock status. The category-complete and full-complete catalogs are sized so the
// player completes one category and the full set.
func TestPlayerHonorsUnlockRules(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLHonorDB(t, ctx)

	svc := New(NewBasicCertRenderer(), nil, Config{})
	player := insertUserRow(t, ctx, "玩家")

	// Catalog: two cattle, both person-category cards. The full active set is 2.
	insertCardRow(t, ctx, 101, cardsvc.CategoryPerson.String())
	insertCardRow(t, ctx, 102, cardsvc.CategoryPerson.String())
	insertCardRow(t, ctx, 201, cardsvc.CategoryEvent.String()) // an event card the player never collects

	// Player feeds 3 times and activates the two person cattle (collecting both
	// person cards) but never activates the event cattle.
	insertFeedingRow(t, ctx, player)
	insertFeedingRow(t, ctx, player)
	insertFeedingRow(t, ctx, player)
	insertActivationRow(t, ctx, player, 101)
	insertActivationRow(t, ctx, player, 102)

	mustCreate(t, ctx, svc, &MutateInput{HonorType: HonorTypeBadge.String(), Code: "join", Name: "参与", UnlockType: UnlockTypeParticipation.String()})
	mustCreate(t, ctx, svc, &MutateInput{HonorType: HonorTypeBadge.String(), Code: "feed3", Name: "喂草3", UnlockType: UnlockTypeFeedCount.String(), Threshold: 3})
	mustCreate(t, ctx, svc, &MutateInput{HonorType: HonorTypeBadge.String(), Code: "feed4", Name: "喂草4", UnlockType: UnlockTypeFeedCount.String(), Threshold: 4})
	mustCreate(t, ctx, svc, &MutateInput{HonorType: HonorTypeBadge.String(), Code: "act2", Name: "激活2", UnlockType: UnlockTypeActivationCount.String(), Threshold: 2})
	mustCreate(t, ctx, svc, &MutateInput{HonorType: HonorTypeBadge.String(), Code: "act3", Name: "激活3", UnlockType: UnlockTypeActivationCount.String(), Threshold: 3})
	mustCreate(t, ctx, svc, &MutateInput{HonorType: HonorTypeCertificate.String(), Code: "person", Name: "集齐人物", UnlockType: UnlockTypeCategoryComplete.String(), Category: cardsvc.CategoryPerson.String()})
	mustCreate(t, ctx, svc, &MutateInput{HonorType: HonorTypeCertificate.String(), Code: "event", Name: "集齐事件", UnlockType: UnlockTypeCategoryComplete.String(), Category: cardsvc.CategoryEvent.String()})
	mustCreate(t, ctx, svc, &MutateInput{HonorType: HonorTypeCertificate.String(), Code: "full", Name: "集齐全套", UnlockType: UnlockTypeFullComplete.String()})

	items, err := svc.PlayerHonors(ctx, player)
	if err != nil {
		t.Fatalf("PlayerHonors failed: %v", err)
	}
	got := unlockedByCode(items)

	want := map[string]bool{
		"join":   true,  // participation always unlocked
		"feed3":  true,  // 3 >= 3
		"feed4":  false, // 3 < 4
		"act2":   true,  // 2 >= 2
		"act3":   false, // 2 < 3
		"person": true,  // collected 2/2 person cards
		"event":  false, // collected 0/1 event card
		"full":   false, // collected 2/3 active cards overall
	}
	for code, wantUnlocked := range want {
		if got[code] != wantUnlocked {
			t.Fatalf("honor %q unlocked = %v, want %v", code, got[code], wantUnlocked)
		}
	}

	// The read-only computation must never write a grant row.
	grantCount, err := dao.UserHonor.Ctx(ctx).Count()
	if err != nil {
		t.Fatalf("count user_honor failed: %v", err)
	}
	if grantCount != 0 {
		t.Fatalf("expected no persisted grant, got %d rows", grantCount)
	}
}

// TestPlayerHonorsFullComplete verifies full_complete unlocks once the player has
// collected every active card.
func TestPlayerHonorsFullComplete(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLHonorDB(t, ctx)

	svc := New(NewBasicCertRenderer(), nil, Config{})
	player := insertUserRow(t, ctx, "全集玩家")

	insertCardRow(t, ctx, 301, cardsvc.CategoryPerson.String())
	insertCardRow(t, ctx, 302, cardsvc.CategoryEvent.String())
	insertActivationRow(t, ctx, player, 301)
	insertActivationRow(t, ctx, player, 302)

	mustCreate(t, ctx, svc, &MutateInput{HonorType: HonorTypeCertificate.String(), Code: "full", Name: "集齐全套", UnlockType: UnlockTypeFullComplete.String()})

	items, err := svc.PlayerHonors(ctx, player)
	if err != nil {
		t.Fatalf("PlayerHonors failed: %v", err)
	}
	if !unlockedByCode(items)["full"] {
		t.Fatalf("expected full_complete unlocked when 2/2 collected")
	}
}

// mustCreate creates a honor definition and fails the test on error.
func mustCreate(t *testing.T, ctx context.Context, svc Service, in *MutateInput) {
	t.Helper()
	if _, err := svc.Create(ctx, in); err != nil {
		t.Fatalf("create honor %q failed: %v", in.Code, err)
	}
}
