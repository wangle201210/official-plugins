// ranking_test.go covers the leaderboard service: the pure Top-N normalization
// applied by New and the DB-gated aggregation and self-rank correctness for the
// three boards. The pure normalization test always runs; the DB tests are gated
// on LINA_TEST_PGSQL_LINK and self-contained.

package ranking

import (
	"context"
	"testing"
)

// TestNewNormalizesTopN verifies that New falls back to the default Top-N for a
// non-positive configured value and preserves a positive one. This is pure logic
// and always runs.
func TestNewNormalizesTopN(t *testing.T) {
	cases := []struct {
		name     string
		configIn int
		want     int
	}{
		{name: "zero falls back to default", configIn: 0, want: defaultTopN},
		{name: "negative falls back to default", configIn: -5, want: defaultTopN},
		{name: "positive preserved", configIn: 25, want: 25},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, ok := New(nil, Config{TopN: tc.configIn}).(*serviceImpl)
			if !ok {
				t.Fatalf("expected *serviceImpl from New")
			}
			if svc.topN != tc.want {
				t.Fatalf("expected topN %d, got %d", tc.want, svc.topN)
			}
		})
	}
}

// TestFeedBoardAggregatesAndRanksSelf verifies the personal feeding board orders
// players by total effect descending, caps at Top-N, assembles nicknames and
// computes the requesting player's own rank.
func TestFeedBoardAggregatesAndRanksSelf(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLRankingDB(t, ctx)

	svc := New(nil, Config{TopN: 2})

	high := insertUserRow(t, ctx, "高分", friendIdentity, 0)
	mid := insertUserRow(t, ctx, "中分", studentIdentity, 0)
	low := insertUserRow(t, ctx, "低分", studentIdentity, 0)

	insertFeedingRow(t, ctx, high, 100)
	insertFeedingRow(t, ctx, high, 50) // high total 150
	insertFeedingRow(t, ctx, mid, 80)  // mid total 80
	insertFeedingRow(t, ctx, low, 10)  // low total 10

	board, err := svc.FeedBoard(ctx, low)
	if err != nil {
		t.Fatalf("FeedBoard failed: %v", err)
	}
	if len(board.List) != 2 {
		t.Fatalf("expected Top-2, got %d", len(board.List))
	}
	if board.List[0].UserId != high || board.List[0].Total != 150 || board.List[0].Rank != 1 {
		t.Fatalf("expected rank 1 high(150), got %+v", board.List[0])
	}
	if board.List[0].Nickname != "高分" {
		t.Fatalf("expected nickname 高分, got %q", board.List[0].Nickname)
	}
	if board.List[1].UserId != mid || board.List[1].Total != 80 || board.List[1].Rank != 2 {
		t.Fatalf("expected rank 2 mid(80), got %+v", board.List[1])
	}
	// low is off the Top-2 but its self rank must still be 3.
	if board.Self == nil || board.Self.Rank != 3 || board.Self.Total != 10 {
		t.Fatalf("expected self rank 3 total 10, got %+v", board.Self)
	}
}

// TestFeedBoardEmptyWhenNoFeeding verifies the board is empty and the self rank is
// zero when no feeding record exists.
func TestFeedBoardEmptyWhenNoFeeding(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLRankingDB(t, ctx)

	svc := New(nil, Config{TopN: 10})
	player := insertUserRow(t, ctx, "无喂草", studentIdentity, 0)

	board, err := svc.FeedBoard(ctx, player)
	if err != nil {
		t.Fatalf("FeedBoard failed: %v", err)
	}
	if len(board.List) != 0 {
		t.Fatalf("expected empty board, got %d rows", len(board.List))
	}
	if board.Self == nil || board.Self.Rank != 0 || board.Self.Total != 0 {
		t.Fatalf("expected self rank 0 total 0, got %+v", board.Self)
	}
}

// TestFeedBoardUsesCompetitionRanksForTies verifies the list and self projection
// use the same rank semantics: equal totals share a rank and the next rank skips.
func TestFeedBoardUsesCompetitionRanksForTies(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLRankingDB(t, ctx)

	svc := New(nil, Config{TopN: 10})
	firstTie := insertUserRow(t, ctx, "并列甲", friendIdentity, 0)
	secondTie := insertUserRow(t, ctx, "并列乙", friendIdentity, 0)
	lower := insertUserRow(t, ctx, "第三名", friendIdentity, 0)
	insertFeedingRow(t, ctx, firstTie, 100)
	insertFeedingRow(t, ctx, secondTie, 100)
	insertFeedingRow(t, ctx, lower, 50)

	board, err := svc.FeedBoard(ctx, secondTie)
	if err != nil {
		t.Fatalf("FeedBoard failed: %v", err)
	}
	if len(board.List) != 3 {
		t.Fatalf("expected three rows, got %d", len(board.List))
	}
	if board.List[0].Rank != 1 || board.List[1].Rank != 1 || board.List[2].Rank != 3 {
		t.Fatalf("expected competition ranks 1,1,3, got %d,%d,%d", board.List[0].Rank, board.List[1].Rank, board.List[2].Rank)
	}
	if board.Self == nil || board.Self.Rank != 1 || board.Self.Total != 100 {
		t.Fatalf("expected tied self rank 1 total 100, got %+v", board.Self)
	}
}

// TestCollegeBoardAggregatesEnrolledStudents verifies the college board sums the
// feeding effect of enrolled students per college, ignores non-students and
// students without a college, and ranks colleges descending.
func TestCollegeBoardAggregatesEnrolledStudents(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLRankingDB(t, ctx)

	svc := New(nil, Config{TopN: 10})

	collegeA := insertCollegeRow(t, ctx, "甲学院")
	collegeB := insertCollegeRow(t, ctx, "乙学院")

	a1 := insertUserRow(t, ctx, "甲一", studentIdentity, collegeA)
	a2 := insertUserRow(t, ctx, "甲二", studentIdentity, collegeA)
	b1 := insertUserRow(t, ctx, "乙一", studentIdentity, collegeB)
	friendInA := insertUserRow(t, ctx, "好友", friendIdentity, collegeA) // not a student, excluded
	noCollege := insertUserRow(t, ctx, "无院", studentIdentity, 0)       // student without college, excluded

	insertFeedingRow(t, ctx, a1, 100)
	insertFeedingRow(t, ctx, a2, 50)  // college A student total 150
	insertFeedingRow(t, ctx, b1, 200) // college B total 200
	insertFeedingRow(t, ctx, friendInA, 999)
	insertFeedingRow(t, ctx, noCollege, 999)

	board, err := svc.CollegeBoard(ctx)
	if err != nil {
		t.Fatalf("CollegeBoard failed: %v", err)
	}
	if len(board.List) != 2 {
		t.Fatalf("expected 2 colleges, got %d", len(board.List))
	}
	if board.List[0].CollegeId != collegeB || board.List[0].Total != 200 || board.List[0].Rank != 1 {
		t.Fatalf("expected rank 1 college B(200), got %+v", board.List[0])
	}
	if board.List[0].CollegeName != "乙学院" {
		t.Fatalf("expected college name 乙学院, got %q", board.List[0].CollegeName)
	}
	if board.List[1].CollegeId != collegeA || board.List[1].Total != 150 || board.List[1].Rank != 2 {
		t.Fatalf("expected rank 2 college A(150), got %+v", board.List[1])
	}
}

// TestFriendBoardRestrictsToFriendsAndRanksSelf verifies the SICAU-friend board
// only ranks friend players and computes the requesting friend's own rank, while
// a non-friend requester gets rank zero.
func TestFriendBoardRestrictsToFriendsAndRanksSelf(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLRankingDB(t, ctx)

	svc := New(nil, Config{TopN: 10})

	friendHigh := insertUserRow(t, ctx, "好友高", friendIdentity, 0)
	friendLow := insertUserRow(t, ctx, "好友低", friendIdentity, 0)
	student := insertUserRow(t, ctx, "学生", studentIdentity, 0)

	insertFeedingRow(t, ctx, friendHigh, 300)
	insertFeedingRow(t, ctx, friendLow, 100)
	insertFeedingRow(t, ctx, student, 999) // excluded from friend board

	board, err := svc.FriendBoard(ctx, friendLow)
	if err != nil {
		t.Fatalf("FriendBoard failed: %v", err)
	}
	if len(board.List) != 2 {
		t.Fatalf("expected 2 friends, got %d", len(board.List))
	}
	if board.List[0].UserId != friendHigh || board.List[0].Total != 300 {
		t.Fatalf("expected friendHigh first, got %+v", board.List[0])
	}
	if board.Self == nil || board.Self.Rank != 2 || board.Self.Total != 100 {
		t.Fatalf("expected friendLow self rank 2 total 100, got %+v", board.Self)
	}

	// A student requester is not in the friend cohort, so self rank must be 0.
	studentBoard, err := svc.FriendBoard(ctx, student)
	if err != nil {
		t.Fatalf("FriendBoard for student failed: %v", err)
	}
	if studentBoard.Self == nil || studentBoard.Self.Rank != 0 {
		t.Fatalf("expected non-friend self rank 0, got %+v", studentBoard.Self)
	}
}
