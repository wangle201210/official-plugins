// honor_crud_test.go covers the operator honor-definition CRUD: code uniqueness,
// honor-type/unlock-type enum validation and the threshold/category shape rules.
// These tests are DB-gated on LINA_TEST_PGSQL_LINK and self-contained.

package honor

import (
	"context"
	"testing"
)

// validParticipation returns a minimal valid participation-honor mutate input with
// the given code.
func validParticipation(code string) *MutateInput {
	return &MutateInput{
		HonorType:  HonorTypeBadge.String(),
		Code:       code,
		Name:       "参与徽章",
		UnlockType: UnlockTypeParticipation.String(),
	}
}

// TestCreateAndGetHonor verifies a valid create persists and is retrievable.
func TestCreateAndGetHonor(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLHonorDB(t, ctx)

	svc := New(NewBasicCertRenderer(), Config{})
	id, err := svc.Create(ctx, validParticipation("join"))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	item, err := svc.Get(ctx, id)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if item.Code != "join" || item.UnlockType != UnlockTypeParticipation.String() {
		t.Fatalf("unexpected honor item: %+v", item)
	}
}

// TestCreateHonorCodeUnique verifies a duplicate active code is rejected.
func TestCreateHonorCodeUnique(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLHonorDB(t, ctx)

	svc := New(NewBasicCertRenderer(), Config{})
	if _, err := svc.Create(ctx, validParticipation("dup")); err != nil {
		t.Fatalf("first Create failed: %v", err)
	}
	_, err := svc.Create(ctx, validParticipation("dup"))
	assertBizCode(t, err, CodeHonorCodeExists.RuntimeCode())
}

// TestCreateHonorEnumValidation verifies invalid honor-type and unlock-type are
// rejected with the corresponding bizerr.
func TestCreateHonorEnumValidation(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLHonorDB(t, ctx)

	svc := New(NewBasicCertRenderer(), Config{})

	badType := validParticipation("bt")
	badType.HonorType = "medal"
	assertBizCodeFromCreate(t, ctx, svc, badType, CodeHonorTypeInvalid.RuntimeCode())

	badUnlock := validParticipation("bu")
	badUnlock.UnlockType = "feed"
	assertBizCodeFromCreate(t, ctx, svc, badUnlock, CodeHonorUnlockTypeInvalid.RuntimeCode())
}

// TestCreateHonorShapeValidation verifies count rules require a positive threshold
// and category_complete requires a valid card category.
func TestCreateHonorShapeValidation(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLHonorDB(t, ctx)

	svc := New(NewBasicCertRenderer(), Config{})

	noThreshold := &MutateInput{
		HonorType:  HonorTypeBadge.String(),
		Code:       "feed0",
		Name:       "喂草徽章",
		UnlockType: UnlockTypeFeedCount.String(),
		Threshold:  0,
	}
	assertBizCodeFromCreate(t, ctx, svc, noThreshold, CodeHonorThresholdInvalid.RuntimeCode())

	badCategory := &MutateInput{
		HonorType:  HonorTypeCertificate.String(),
		Code:       "cat0",
		Name:       "集齐证书",
		UnlockType: UnlockTypeCategoryComplete.String(),
		Category:   "unknown",
	}
	assertBizCodeFromCreate(t, ctx, svc, badCategory, CodeHonorCategoryInvalid.RuntimeCode())

	emptyCategory := &MutateInput{
		HonorType:  HonorTypeCertificate.String(),
		Code:       "cat1",
		Name:       "集齐证书",
		UnlockType: UnlockTypeCategoryComplete.String(),
		Category:   "",
	}
	assertBizCodeFromCreate(t, ctx, svc, emptyCategory, CodeHonorCategoryInvalid.RuntimeCode())
}

// TestUpdateHonorKeepsOwnCode verifies updating a honor with its unchanged code
// does not collide with itself.
func TestUpdateHonorKeepsOwnCode(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLHonorDB(t, ctx)

	svc := New(NewBasicCertRenderer(), Config{})
	id, err := svc.Create(ctx, validParticipation("self"))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	in := validParticipation("self")
	in.Name = "更新名"
	if err = svc.Update(ctx, id, in); err != nil {
		t.Fatalf("Update with unchanged code failed: %v", err)
	}
}

// TestDeleteHonorRemovesFromList verifies a soft-deleted honor leaves the list.
func TestDeleteHonorRemovesFromList(t *testing.T) {
	ctx := context.Background()
	setupPostgreSQLHonorDB(t, ctx)

	svc := New(NewBasicCertRenderer(), Config{})
	id, err := svc.Create(ctx, validParticipation("gone"))
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if err = svc.Delete(ctx, id); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	out, err := svc.List(ctx, &ListInput{})
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if out.Total != 0 {
		t.Fatalf("expected empty list after delete, got total %d", out.Total)
	}
	// The freed code can be reused after deletion.
	if _, err = svc.Create(ctx, validParticipation("gone")); err != nil {
		t.Fatalf("expected to reuse freed code, got %v", err)
	}
}

// assertBizCodeFromCreate runs Create and asserts the resulting bizerr code.
func assertBizCodeFromCreate(t *testing.T, ctx context.Context, svc Service, in *MutateInput, wantCode string) {
	t.Helper()
	_, err := svc.Create(ctx, in)
	assertBizCode(t, err, wantCode)
}
