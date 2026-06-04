// cattle_niu_test.go verifies cattle CRUD, cascade card soft-delete and list
// batch-assembly against a gated PostgreSQL database. The tests build the cattle
// service with the real college service so linked-college validation runs against
// stored college rows. Database-backed assertions are skipped unless
// LINA_TEST_PGSQL_LINK is set.

package cattle

import (
	"context"
	"testing"

	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
)

// newCattleServiceForTest builds a cattle service wired to the real college
// service so college-existence validation runs against stored rows.
func newCattleServiceForTest() Service {
	return New(collegesvc.New())
}

// seedCollege inserts one active college and returns its ID for linking
// college-subtype cattle.
func seedCollege(t *testing.T, ctx context.Context, name string) int64 {
	t.Helper()
	id, err := dao.College.Ctx(ctx).Data(do.College{Name: name, Sort: 1}).InsertAndGetId()
	if err != nil {
		t.Fatalf("seed college %q failed: %v", name, err)
	}
	return id
}

// TestCreateNiuRejectsDuplicateCode verifies a second cattle reusing an active
// code is rejected with CodeNiuCodeExists while the first remains.
func TestCreateNiuRejectsDuplicateCode(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCattleDB(t, ctx)

	svc := newCattleServiceForTest()
	const code = "NIU-DUP-001"
	if _, err := svc.CreateNiu(ctx, &NiuMutateInput{Code: code, NiuType: NiuTypeCommon.String()}); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err := svc.CreateNiu(ctx, &NiuMutateInput{Code: code, NiuType: NiuTypeCommon.String()})
	assertBizCode(t, err, CodeNiuCodeExists.RuntimeCode())

	count, err := dao.Niu.Ctx(ctx).Where(dao.Niu.Columns().Code, code).Count()
	if err != nil {
		t.Fatalf("count cattle failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected one cattle after duplicate create, got %d", count)
	}
}

// TestCreateNiuRejectsInvalidTypeAndSubtype verifies an unknown type and an
// unknown special subtype are each rejected with their specific bizerr.
func TestCreateNiuRejectsInvalidTypeAndSubtype(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCattleDB(t, ctx)

	svc := newCattleServiceForTest()

	_, err := svc.CreateNiu(ctx, &NiuMutateInput{Code: "NIU-BAD-TYPE", NiuType: "rare"})
	assertBizCode(t, err, CodeNiuTypeInvalid.RuntimeCode())

	_, err = svc.CreateNiu(ctx, &NiuMutateInput{
		Code:           "NIU-BAD-SUBTYPE",
		NiuType:        NiuTypeSpecial.String(),
		Name:           "Spirit Ox",
		SpecialSubtype: "campus",
	})
	assertBizCode(t, err, CodeNiuSubtypeInvalid.RuntimeCode())
}

// TestCreateNiuCollegeSubtypeValidatesCollege verifies a college-subtype cattle
// pointing at a missing college is rejected with CodeNiuCollegeInvalid while one
// pointing at an existing college succeeds.
func TestCreateNiuCollegeSubtypeValidatesCollege(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCattleDB(t, ctx)

	svc := newCattleServiceForTest()

	// A missing linked college is rejected.
	_, err := svc.CreateNiu(ctx, &NiuMutateInput{
		Code:           "NIU-COL-MISSING",
		NiuType:        NiuTypeSpecial.String(),
		SpecialSubtype: SpecialSubtypeCollege.String(),
		Name:           "College Ox",
		CollegeId:      999999,
	})
	assertBizCode(t, err, CodeNiuCollegeInvalid.RuntimeCode())

	// An existing linked college succeeds.
	collegeID := seedCollege(t, ctx, "Animal Science")
	id, err := svc.CreateNiu(ctx, &NiuMutateInput{
		Code:           "NIU-COL-OK",
		NiuType:        NiuTypeSpecial.String(),
		SpecialSubtype: SpecialSubtypeCollege.String(),
		Name:           "College Ox",
		CollegeId:      collegeID,
	})
	if err != nil {
		t.Fatalf("create college cattle failed: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive cattle ID, got %d", id)
	}
}

// TestCreateNiuCollegeSubtypeRequiresCollege verifies a college-subtype cattle
// with no linked college is rejected with CodeNiuCollegeRequired.
func TestCreateNiuCollegeSubtypeRequiresCollege(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCattleDB(t, ctx)

	svc := newCattleServiceForTest()
	_, err := svc.CreateNiu(ctx, &NiuMutateInput{
		Code:           "NIU-COL-REQ",
		NiuType:        NiuTypeSpecial.String(),
		SpecialSubtype: SpecialSubtypeCollege.String(),
		Name:           "College Ox",
		CollegeId:      0,
	})
	assertBizCode(t, err, CodeNiuCollegeRequired.RuntimeCode())
}

// TestCreateNiuDefaultsStatusInactive verifies a freshly created cattle persists
// the inactive status because activation is owned by the C3 flow.
func TestCreateNiuDefaultsStatusInactive(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCattleDB(t, ctx)

	svc := newCattleServiceForTest()
	id, err := svc.CreateNiu(ctx, &NiuMutateInput{Code: "NIU-STATUS", NiuType: NiuTypeCommon.String()})
	if err != nil {
		t.Fatalf("create cattle failed: %v", err)
	}

	got, err := svc.GetNiu(ctx, id)
	if err != nil {
		t.Fatalf("get cattle failed: %v", err)
	}
	if got.Status != NiuStatusInactive.String() {
		t.Fatalf("expected default status %q, got %q", NiuStatusInactive.String(), got.Status)
	}
}

// TestDeleteNiuCascadeSoftDeletesCard verifies deleting a cattle that owns a card
// cascade soft-deletes the card so no dangling card remains.
func TestDeleteNiuCascadeSoftDeletesCard(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCattleDB(t, ctx)

	svc := newCattleServiceForTest()
	niuID, err := svc.CreateNiu(ctx, &NiuMutateInput{Code: "NIU-CASCADE", NiuType: NiuTypeCommon.String()})
	if err != nil {
		t.Fatalf("create cattle failed: %v", err)
	}

	// Insert a card directly bound to the cattle to exercise the cascade.
	cardID, err := dao.Card.Ctx(ctx).Data(do.Card{
		NiuId:    niuID,
		Category: "person",
		Title:    "Founder",
	}).InsertAndGetId()
	if err != nil {
		t.Fatalf("insert bound card failed: %v", err)
	}

	if err = svc.DeleteNiu(ctx, niuID); err != nil {
		t.Fatalf("delete cattle failed: %v", err)
	}

	// The cattle is gone.
	exists, err := svc.NiuExists(ctx, niuID)
	if err != nil {
		t.Fatalf("niu exists check failed: %v", err)
	}
	if exists {
		t.Fatal("expected cattle to be soft-deleted")
	}

	// The card is cascade soft-deleted: the active set no longer sees it.
	cardCount, err := dao.Card.Ctx(ctx).Where(do.Card{Id: cardID}).Count()
	if err != nil {
		t.Fatalf("count card failed: %v", err)
	}
	if cardCount != 0 {
		t.Fatalf("expected owning card to be cascade soft-deleted, found %d active rows", cardCount)
	}
}

// TestDeleteNiuMissingReturnsNotFound verifies deleting an unknown cattle returns
// CodeNiuNotFound.
func TestDeleteNiuMissingReturnsNotFound(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCattleDB(t, ctx)

	svc := newCattleServiceForTest()
	err := svc.DeleteNiu(ctx, 424242)
	assertBizCode(t, err, CodeNiuNotFound.RuntimeCode())
}

// TestListNiuAssemblesCollegeNameAndCardFlag verifies the list batch-assembles
// the linked college name and card-binding flag for each row and that the page
// size cap and pagination are honoured.
func TestListNiuAssemblesCollegeNameAndCardFlag(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCattleDB(t, ctx)

	svc := newCattleServiceForTest()
	const collegeName = "Veterinary Medicine"
	collegeID := seedCollege(t, ctx, collegeName)

	// Cattle A: college-linked and card-bound.
	niuA, err := svc.CreateNiu(ctx, &NiuMutateInput{
		Code:           "NIU-LIST-A",
		NiuType:        NiuTypeSpecial.String(),
		SpecialSubtype: SpecialSubtypeCollege.String(),
		Name:           "College Ox",
		CollegeId:      collegeID,
	})
	if err != nil {
		t.Fatalf("create cattle A failed: %v", err)
	}
	if _, err = dao.Card.Ctx(ctx).Data(do.Card{NiuId: niuA, Category: "college", Title: "History"}).InsertAndGetId(); err != nil {
		t.Fatalf("bind card to cattle A failed: %v", err)
	}

	// Cattle B: common, no college, no card.
	if _, err = svc.CreateNiu(ctx, &NiuMutateInput{Code: "NIU-LIST-B", NiuType: NiuTypeCommon.String()}); err != nil {
		t.Fatalf("create cattle B failed: %v", err)
	}

	out, err := svc.ListNiu(ctx, &ListNiuInput{PageNum: 1, PageSize: 10})
	if err != nil {
		t.Fatalf("list cattle failed: %v", err)
	}
	if out.Total != 2 {
		t.Fatalf("expected total 2, got %d", out.Total)
	}
	if len(out.List) != 2 {
		t.Fatalf("expected 2 rows on page, got %d", len(out.List))
	}

	byCode := make(map[string]*NiuItem, len(out.List))
	for _, item := range out.List {
		byCode[item.Code] = item
	}

	itemA := byCode["NIU-LIST-A"]
	if itemA == nil {
		t.Fatal("expected cattle A in the list")
	}
	if itemA.CollegeName != collegeName {
		t.Fatalf("expected assembled college name %q, got %q", collegeName, itemA.CollegeName)
	}
	if !itemA.HasCard {
		t.Fatal("expected cattle A to report a bound card")
	}

	itemB := byCode["NIU-LIST-B"]
	if itemB == nil {
		t.Fatal("expected cattle B in the list")
	}
	if itemB.CollegeName != "" {
		t.Fatalf("expected cattle B to have no college name, got %q", itemB.CollegeName)
	}
	if itemB.HasCard {
		t.Fatal("expected cattle B to report no bound card")
	}
}

// TestListNiuCapsPageSizeAndPaginates verifies the page-size cap and ID-desc
// pagination across multiple pages.
func TestListNiuCapsPageSizeAndPaginates(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCattleDB(t, ctx)

	svc := newCattleServiceForTest()
	const total = 5
	for i := 0; i < total; i++ {
		code := "NIU-PAGE-" + string(rune('A'+i))
		if _, err := svc.CreateNiu(ctx, &NiuMutateInput{Code: code, NiuType: NiuTypeCommon.String()}); err != nil {
			t.Fatalf("create cattle %s failed: %v", code, err)
		}
	}

	// An over-cap page size is clamped to maxPageSize but still returns the true
	// total. With only 5 rows all fit on one page.
	out, err := svc.ListNiu(ctx, &ListNiuInput{PageNum: 1, PageSize: 1000})
	if err != nil {
		t.Fatalf("list cattle failed: %v", err)
	}
	if out.Total != total {
		t.Fatalf("expected total %d, got %d", total, out.Total)
	}
	if len(out.List) != total {
		t.Fatalf("expected %d rows, got %d", total, len(out.List))
	}

	// Page 1 with size 2 returns the two newest rows in ID-descending order.
	page1, err := svc.ListNiu(ctx, &ListNiuInput{PageNum: 1, PageSize: 2})
	if err != nil {
		t.Fatalf("list page 1 failed: %v", err)
	}
	if len(page1.List) != 2 {
		t.Fatalf("expected 2 rows on page 1, got %d", len(page1.List))
	}
	if page1.List[0].Id <= page1.List[1].Id {
		t.Fatalf("expected ID-descending order, got %d then %d", page1.List[0].Id, page1.List[1].Id)
	}

	// Page 3 with size 2 returns the single remaining oldest row.
	page3, err := svc.ListNiu(ctx, &ListNiuInput{PageNum: 3, PageSize: 2})
	if err != nil {
		t.Fatalf("list page 3 failed: %v", err)
	}
	if len(page3.List) != 1 {
		t.Fatalf("expected 1 row on page 3, got %d", len(page3.List))
	}
}
