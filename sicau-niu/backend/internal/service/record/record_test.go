// record_test.go holds the database-gated integration tests for the activity-record
// queries: feeding filtering/pagination with player+cattle name assembly, steal dual
// nickname assembly and grass-ledger paging. The tests are skipped unless
// LINA_TEST_PGSQL_LINK is set.

package record

import (
	"context"
	"testing"
)

// TestListFeedingsFilterPaginateAssemble verifies feeding records are filtered by
// cattle, paged on the database side, and carry assembled player and cattle names.
func TestListFeedingsFilterPaginateAssemble(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLRecordDB(t, ctx)
	svc := New()

	userA := insertUserRow(t, ctx, "玩家甲")
	userB := insertUserRow(t, ctx, "玩家乙")
	niu1 := insertNiuRow(t, ctx, "信工牛")
	niu2 := insertNiuRow(t, ctx, "水利牛")
	// 3 feedings on niu1 (2 by A, 1 by B), 1 feeding on niu2.
	insertFeedingRow(t, ctx, userA, niu1, 20, 30)
	insertFeedingRow(t, ctx, userA, niu1, 10, 15)
	insertFeedingRow(t, ctx, userB, niu1, 12, 12)
	insertFeedingRow(t, ctx, userA, niu2, 40, 60)

	// Filter by niu1: 3 records total, page size 2 -> 2 on page 1.
	out, err := svc.ListFeedings(ctx, &ListFeedingsInput{NiuId: niu1, PageNum: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("ListFeedings error: %v", err)
	}
	if out.Total != 3 {
		t.Fatalf("expected total 3 for niu1, got %d", out.Total)
	}
	if len(out.List) != 2 {
		t.Fatalf("expected page size 2, got %d", len(out.List))
	}
	for _, row := range out.List {
		if row.NiuId != niu1 || row.NiuName != "信工牛" || row.NiuCode == "" {
			t.Fatalf("cattle not assembled: %+v", row)
		}
		if row.Nickname == "" {
			t.Fatalf("player nickname not assembled for user %d", row.UserId)
		}
	}

	// Filter by user B on niu1: only 1 record.
	byB, err := svc.ListFeedings(ctx, &ListFeedingsInput{NiuId: niu1, UserId: userB, PageNum: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListFeedings(B) error: %v", err)
	}
	if byB.Total != 1 || len(byB.List) != 1 || byB.List[0].Nickname != "玩家乙" {
		t.Fatalf("expected 1 feeding by 玩家乙, got total=%d list=%d", byB.Total, len(byB.List))
	}
}

// TestListStealsAssemblesBothNicknames verifies steal records carry both actor and
// target nicknames.
func TestListStealsAssemblesBothNicknames(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLRecordDB(t, ctx)
	svc := New()

	actor := insertUserRow(t, ctx, "偷草的")
	target := insertUserRow(t, ctx, "被偷的")
	insertStealRow(t, ctx, actor, target, 1)
	insertStealRow(t, ctx, actor, target, 2)

	out, err := svc.ListSteals(ctx, &ListStealsInput{PageNum: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("ListSteals error: %v", err)
	}
	if out.Total != 2 || len(out.List) != 2 {
		t.Fatalf("expected 2 steals, got total=%d list=%d", out.Total, len(out.List))
	}
	row := out.List[0]
	if row.ActorNickname != "偷草的" || row.TargetNickname != "被偷的" {
		t.Fatalf("nicknames not assembled: actor=%q target=%q", row.ActorNickname, row.TargetNickname)
	}

	// Filter by actor returns the same two; filter by an unrelated target returns none.
	byActor, _ := svc.ListSteals(ctx, &ListStealsInput{ActorUserId: actor, PageNum: 1, PageSize: 10})
	if byActor.Total != 2 {
		t.Fatalf("expected 2 by actor, got %d", byActor.Total)
	}
	none, _ := svc.ListSteals(ctx, &ListStealsInput{TargetUserId: actor, PageNum: 1, PageSize: 10})
	if none.Total != 0 || len(none.List) != 0 {
		t.Fatalf("expected empty steal page, got total=%d", none.Total)
	}
}

// TestListGrassTxnsPaged verifies grass-ledger entries are paged and carry the player
// nickname.
func TestListGrassTxnsPaged(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLRecordDB(t, ctx)
	svc := New()

	user := insertUserRow(t, ctx, "账本玩家")
	insertGrassTxnRow(t, ctx, user, 35, "checkin")
	insertGrassTxnRow(t, ctx, user, -12, "gift")
	insertGrassTxnRow(t, ctx, user, 8, "steal")

	out, err := svc.ListGrassTxns(ctx, &ListGrassTxnsInput{UserId: user, PageNum: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("ListGrassTxns error: %v", err)
	}
	if out.Total != 3 {
		t.Fatalf("expected total 3, got %d", out.Total)
	}
	if len(out.List) != 2 {
		t.Fatalf("expected page size 2, got %d", len(out.List))
	}
	if out.List[0].Nickname != "账本玩家" {
		t.Fatalf("nickname not assembled, got %q", out.List[0].Nickname)
	}
}
