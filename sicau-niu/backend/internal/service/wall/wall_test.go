// wall_test.go holds the database-gated integration tests for the public
// memorial-wall service: the first-activator wall derivation and ordering, the
// privacy guarantee that no phone is ever surfaced, the campus-history highlights
// sampling (enabled quotes only, card cattle names assembled) and the public
// activity stat counts. The tests are skipped unless LINA_TEST_PGSQL_LINK is set.

package wall

import (
	"context"
	"strings"
	"testing"
)

// TestFirstActivatorsReturnsOnlyFirstOrdered verifies the wall returns only
// is_first activations, ordered by activation time ascending with 1-based
// sequence numbers, with player identity and cattle name/code batch-assembled,
// and never surfaces the seeded phone.
func TestFirstActivatorsReturnsOnlyFirstOrdered(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLWallDB(t, ctx)
	svc := New()

	user1 := insertUserRow(t, ctx, "首发牛友A", "student")
	user2 := insertUserRow(t, ctx, "首发牛友B", "friend")
	user3 := insertUserRow(t, ctx, "普通牛友C", "student")
	niu1 := insertNiuRow(t, ctx, "信工牛", "active")
	niu2 := insertNiuRow(t, ctx, "水利牛", "active")
	niu3 := insertNiuRow(t, ctx, "农学牛", "active")

	// user1 is the first activator earliest, user2 next; user3 is a non-first
	// activation and must be excluded from the wall.
	insertActivationRow(t, ctx, user1, niu1, 1, 1)
	insertActivationRow(t, ctx, user2, niu2, 1, 2)
	insertActivationRow(t, ctx, user3, niu3, 0, 3)

	board, err := svc.FirstActivators(ctx)
	if err != nil {
		t.Fatalf("FirstActivators returned error: %v", err)
	}
	if len(board.List) != 2 {
		t.Fatalf("expected 2 first activators, got %d", len(board.List))
	}

	first := board.List[0]
	if first.Seq != 1 || first.UserId != user1 {
		t.Fatalf("expected first row seq=1 user=%d, got seq=%d user=%d", user1, first.Seq, first.UserId)
	}
	if first.Nickname != "首发牛友A" || first.IdentityType != "student" {
		t.Fatalf("first row identity not assembled: nickname=%q identity=%q", first.Nickname, first.IdentityType)
	}
	if first.NiuName != "信工牛" || first.NiuCode == "" {
		t.Fatalf("first row cattle not assembled: name=%q code=%q", first.NiuName, first.NiuCode)
	}
	if first.ActivatedAt == nil || *first.ActivatedAt <= 0 {
		t.Fatalf("first row activatedAt not populated: %v", first.ActivatedAt)
	}

	second := board.List[1]
	if second.Seq != 2 || second.UserId != user2 {
		t.Fatalf("expected second row seq=2 user=%d, got seq=%d user=%d", user2, second.Seq, second.UserId)
	}

	// Privacy guard: the seeded phone for user1 (138000000001-style) must never
	// appear in any projected public field.
	for _, item := range board.List {
		joined := strings.Join([]string{item.Nickname, item.IdentityType, item.NiuName, item.NiuCode}, "|")
		if strings.Contains(joined, "1380000") {
			t.Fatalf("public wall leaked a phone-like value: %q", joined)
		}
	}
}

// TestFirstActivatorsEmpty verifies the wall returns an empty list when there are
// no first activations.
func TestFirstActivatorsEmpty(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLWallDB(t, ctx)
	svc := New()

	board, err := svc.FirstActivators(ctx)
	if err != nil {
		t.Fatalf("FirstActivators returned error: %v", err)
	}
	if len(board.List) != 0 {
		t.Fatalf("expected empty wall, got %d rows", len(board.List))
	}
}

// TestHighlightsCardsAndEnabledQuotes verifies the highlights sample assembles
// card cattle names and returns only enabled quotes.
func TestHighlightsCardsAndEnabledQuotes(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLWallDB(t, ctx)
	svc := New()

	// One main card per cattle (the card table is unique on niu_id), so seed two
	// cattle to exercise the multi-card sample.
	niu1 := insertNiuRow(t, ctx, "校史牛一", "active")
	niu2 := insertNiuRow(t, ctx, "校史牛二", "active")
	insertCardRow(t, ctx, niu1, "history", "建校120周年")
	insertCardRow(t, ctx, niu2, "spirit", "川农大精神")
	insertQuoteRow(t, ctx, "爱国敬业、艰苦奋斗", 1)
	insertQuoteRow(t, ctx, "停用金句不应展示", 0)

	highlights, err := svc.Highlights(ctx)
	if err != nil {
		t.Fatalf("Highlights returned error: %v", err)
	}
	if len(highlights.Cards) != 2 {
		t.Fatalf("expected 2 cards, got %d", len(highlights.Cards))
	}
	for _, card := range highlights.Cards {
		if card.NiuName == "" {
			t.Fatalf("card cattle name not assembled for niu %d", card.NiuId)
		}
	}
	if len(highlights.Quotes) != 1 {
		t.Fatalf("expected only the enabled quote, got %d", len(highlights.Quotes))
	}
	if highlights.Quotes[0].Content != "爱国敬业、艰苦奋斗" {
		t.Fatalf("unexpected enabled quote content: %q", highlights.Quotes[0].Content)
	}
}

// TestStatsCounts verifies the public stats count activated cattle, total cattle,
// first activators and players on the database side.
func TestStatsCounts(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLWallDB(t, ctx)
	svc := New()

	user1 := insertUserRow(t, ctx, "玩家A", "student")
	user2 := insertUserRow(t, ctx, "玩家B", "friend")
	activeNiu1 := insertNiuRow(t, ctx, "活跃牛一", "active")
	insertNiuRow(t, ctx, "活跃牛二", "active")
	insertNiuRow(t, ctx, "未激活牛", "inactive")

	insertActivationRow(t, ctx, user1, activeNiu1, 1, 1)
	insertActivationRow(t, ctx, user2, activeNiu1, 0, 2)

	stats, err := svc.Stats(ctx)
	if err != nil {
		t.Fatalf("Stats returned error: %v", err)
	}
	if stats.ActivatedNiuCount != 2 {
		t.Fatalf("expected 2 activated cattle, got %d", stats.ActivatedNiuCount)
	}
	if stats.TotalNiuCount != 3 {
		t.Fatalf("expected 3 total cattle, got %d", stats.TotalNiuCount)
	}
	if stats.FirstActivatorCount != 1 {
		t.Fatalf("expected 1 first activator, got %d", stats.FirstActivatorCount)
	}
	if stats.PlayerCount != 2 {
		t.Fatalf("expected 2 players, got %d", stats.PlayerCount)
	}
}
