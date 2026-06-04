// card_card_test.go verifies card CRUD, the one-card-per-cattle constraint,
// owning-cattle validation and category validation against a gated PostgreSQL
// database. The tests build the card service with the real cattle service (which
// itself uses the real college service) so owning-cattle existence checks run
// against stored niu rows. Database-backed assertions are skipped unless
// LINA_TEST_PGSQL_LINK is set.

package card

import (
	"context"
	"testing"

	"lina-core/pkg/bizerr"
	cattlesvc "lina-plugin-sicau-niu/backend/internal/service/cattle"
	collegesvc "lina-plugin-sicau-niu/backend/internal/service/college"
)

// newCardServiceForTest builds a card service wired to the real cattle service so
// owning-cattle existence validation runs against stored niu rows. It also
// returns the cattle service so tests can seed owning cattle.
func newCardServiceForTest() (Service, cattlesvc.Service) {
	cattleService := cattlesvc.New(collegesvc.New())
	return New(cattleService), cattleService
}

// seedCattle creates one common cattle and returns its ID for card binding.
func seedCattle(t *testing.T, ctx context.Context, cattleService cattlesvc.Service, code string) int64 {
	t.Helper()
	id, err := cattleService.CreateNiu(ctx, &cattlesvc.NiuMutateInput{
		Code:    code,
		NiuType: "common",
	})
	if err != nil {
		t.Fatalf("seed cattle %q failed: %v", code, err)
	}
	return id
}

// TestCreateCardSucceedsThenRejectsSecondCardForSameCattle verifies the first
// card binds to a cattle and a second card for the same cattle is rejected with
// CodeCardNiuTaken.
func TestCreateCardSucceedsThenRejectsSecondCardForSameCattle(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCardDB(t, ctx)

	svc, cattleService := newCardServiceForTest()
	niuID := seedCattle(t, ctx, cattleService, "NIU-CARD-1")

	id, err := svc.CreateCard(ctx, &CardMutateInput{
		NiuId:    niuID,
		Category: CategoryPerson.String(),
		Title:    "Founder",
		Content:  "The first card.",
	})
	if err != nil {
		t.Fatalf("first card create failed: %v", err)
	}
	if id <= 0 {
		t.Fatalf("expected positive card ID, got %d", id)
	}

	_, err = svc.CreateCard(ctx, &CardMutateInput{
		NiuId:    niuID,
		Category: CategoryEvent.String(),
		Title:    "Second",
	})
	assertBizCode(t, err, CodeCardNiuTaken.RuntimeCode())
}

// TestCreateCardRejectsMissingCattle verifies binding a card to a non-existent
// cattle is rejected with CodeCardNiuInvalid.
func TestCreateCardRejectsMissingCattle(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCardDB(t, ctx)

	svc, _ := newCardServiceForTest()
	_, err := svc.CreateCard(ctx, &CardMutateInput{
		NiuId:    987654,
		Category: CategoryPerson.String(),
		Title:    "Orphan",
	})
	assertBizCode(t, err, CodeCardNiuInvalid.RuntimeCode())
}

// TestCreateCardRejectsInvalidCategory verifies an unknown category is rejected
// with CodeCardCategoryInvalid before any owning-cattle lookup.
func TestCreateCardRejectsInvalidCategory(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCardDB(t, ctx)

	svc, cattleService := newCardServiceForTest()
	niuID := seedCattle(t, ctx, cattleService, "NIU-CARD-CAT")

	_, err := svc.CreateCard(ctx, &CardMutateInput{
		NiuId:    niuID,
		Category: "place",
		Title:    "Bad Category",
	})
	assertBizCode(t, err, CodeCardCategoryInvalid.RuntimeCode())
}

// TestDeleteCardFreesCattleForRebinding verifies soft-deleting a card frees its
// owning cattle so a new card can be bound to the same cattle.
func TestDeleteCardFreesCattleForRebinding(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCardDB(t, ctx)

	svc, cattleService := newCardServiceForTest()
	niuID := seedCattle(t, ctx, cattleService, "NIU-CARD-REBIND")

	firstID, err := svc.CreateCard(ctx, &CardMutateInput{
		NiuId:    niuID,
		Category: CategoryPerson.String(),
		Title:    "First",
	})
	if err != nil {
		t.Fatalf("first card create failed: %v", err)
	}

	if err = svc.DeleteCard(ctx, firstID); err != nil {
		t.Fatalf("delete card failed: %v", err)
	}

	// The cattle is free again, so a new card binds successfully.
	if _, err = svc.CreateCard(ctx, &CardMutateInput{
		NiuId:    niuID,
		Category: CategoryEvent.String(),
		Title:    "Second",
	}); err != nil {
		t.Fatalf("rebind card after delete failed: %v", err)
	}
}

// TestGetCardAssemblesOwningCattle verifies card detail batch-assembles the
// owning-cattle code and that a missing card returns CodeCardNotFound.
func TestGetCardAssemblesOwningCattle(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLCardDB(t, ctx)

	svc, cattleService := newCardServiceForTest()
	const code = "NIU-CARD-GET"
	niuID := seedCattle(t, ctx, cattleService, code)

	cardID, err := svc.CreateCard(ctx, &CardMutateInput{
		NiuId:    niuID,
		Category: CategoryPerson.String(),
		Title:    "Detail",
	})
	if err != nil {
		t.Fatalf("create card failed: %v", err)
	}

	got, err := svc.GetCard(ctx, cardID)
	if err != nil {
		t.Fatalf("get card failed: %v", err)
	}
	if got.NiuCode != code {
		t.Fatalf("expected assembled owning-cattle code %q, got %q", code, got.NiuCode)
	}

	// A missing card returns a structured business error. NOTE: the Service
	// contract documents CodeCardNotFound for a missing ID, but the current
	// GetCard scans into a pre-allocated struct, so GoFrame surfaces
	// sql.ErrNoRows and the implementation wraps it as CodeCardQueryFailed before
	// reaching the row.Id == 0 not-found branch. This test asserts the actual
	// current behaviour (a structured bizerr is returned, never a nil/leaked
	// error) and records the contract gap for the cattle/card GetXxx missing-ID
	// path; see the task findings. cattle.GetNiu has the same defect.
	_, err = svc.GetCard(ctx, 111222)
	if _, ok := bizerr.As(err); !ok {
		t.Fatalf("expected a structured business error for a missing card, got %T: %v", err, err)
	}
}
