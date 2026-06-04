// grasssocial_inbox_test.go covers the DB-gated player inbox: a player sees only
// their own messages, marking an owned message read succeeds, and marking another
// player's message is rejected. The inbox pagination normalization is covered as a
// pure-logic check.

package grasssocial

import (
	"context"
	"testing"
)

// TestMessagesSelfIsolation verifies a player's inbox contains only their own
// messages, produced here by a gift to them.
func TestMessagesSelfIsolation(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()

	giver := insertUserRow(t, ctx, "openid-inbox-giver")
	me := insertUserRow(t, ctx, "openid-inbox-me")
	other := insertUserRow(t, ctx, "openid-inbox-other")
	seedGrass(t, ctx, giver, 100)

	if _, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: me, Amount: 12}); err != nil {
		t.Fatalf("gift to me failed: %v", err)
	}

	mine, err := svc.Messages(ctx, me, &MessagesInput{})
	if err != nil {
		t.Fatalf("messages failed: %v", err)
	}
	if mine.Total != 1 || len(mine.List) != 1 {
		t.Fatalf("expected 1 of my messages, got total %d list %d", mine.Total, len(mine.List))
	}

	theirs, err := svc.Messages(ctx, other, &MessagesInput{})
	if err != nil {
		t.Fatalf("messages failed: %v", err)
	}
	if theirs.Total != 0 {
		t.Fatalf("expected the other player to have no messages, got %d", theirs.Total)
	}
}

// TestMarkReadOwnership verifies a player can mark their own message read but
// cannot mark another player's message.
func TestMarkReadOwnership(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLSocialDB(t, ctx)
	svc := newSocialServiceForTest()

	giver := insertUserRow(t, ctx, "openid-mark-giver")
	me := insertUserRow(t, ctx, "openid-mark-me")
	stranger := insertUserRow(t, ctx, "openid-mark-stranger")
	seedGrass(t, ctx, giver, 100)

	if _, err := svc.Gift(ctx, giver, &GiftInput{ToUserId: me, Amount: 12}); err != nil {
		t.Fatalf("gift to me failed: %v", err)
	}
	mine, err := svc.Messages(ctx, me, &MessagesInput{})
	if err != nil {
		t.Fatalf("messages failed: %v", err)
	}
	messageID := mine.List[0].Id

	// A stranger cannot mark my message read.
	err = svc.MarkRead(ctx, stranger, messageID)
	assertBizCode(t, err, CodeMessageNotFound.RuntimeCode())

	// I can mark my own message read.
	if err = svc.MarkRead(ctx, me, messageID); err != nil {
		t.Fatalf("mark read failed: %v", err)
	}
	after, err := svc.Messages(ctx, me, &MessagesInput{})
	if err != nil {
		t.Fatalf("messages failed: %v", err)
	}
	if !after.List[0].IsRead {
		t.Fatalf("expected message marked read")
	}
}

// TestNormalizeInboxPagination verifies the inbox paging defaults and the max
// page-size cap, without a database.
func TestNormalizeInboxPagination(t *testing.T) {
	cases := []struct {
		name              string
		in                *MessagesInput
		wantNum, wantSize int
	}{
		{"nil", nil, defaultInboxPageNum, defaultInboxPageSize},
		{"zero", &MessagesInput{}, defaultInboxPageNum, defaultInboxPageSize},
		{"explicit", &MessagesInput{PageNum: 3, PageSize: 25}, 3, 25},
		{"over-cap", &MessagesInput{PageNum: 1, PageSize: 500}, 1, maxInboxPageSize},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			gotNum, gotSize := normalizeInboxPagination(tc.in)
			if gotNum != tc.wantNum || gotSize != tc.wantSize {
				t.Fatalf("normalizeInboxPagination = (%d,%d), want (%d,%d)", gotNum, gotSize, tc.wantNum, tc.wantSize)
			}
		})
	}
}
