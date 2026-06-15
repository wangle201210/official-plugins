// This file verifies CMS album management, whole-album image replacement,
// cascade deletion, and public visibility against an isolated PostgreSQL schema.

package cms

import (
	"context"
	"fmt"
	"testing"

	"lina-core/pkg/bizerr"
	"lina-plugin-cms/backend/internal/dao"
)

// createTestAlbum creates one album through the service for test setup.
func createTestAlbum(t *testing.T, ctx context.Context, svc Service, in AlbumSaveInput) int64 {
	t.Helper()

	id, err := svc.CreateAlbum(ctx, in)
	if err != nil {
		t.Fatalf("create CMS album %s: %v", in.Name, err)
	}
	return id
}

// TestAlbumWholeImageReplacement verifies album image lists are replaced as a
// whole on save and ordered by sort.
func TestAlbumWholeImageReplacement(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{code: "albums", status: StatusEnabled, typeID: CategoryTypeAlbum})
	svc := newTestCMSService()
	id := createTestAlbum(t, ctx, svc, AlbumSaveInput{
		CategoryId: categoryID,
		Name:       "Campus",
		Status:     StatusEnabled,
		Images: []AlbumImageInput{
			{Url: "/uploads/a1.png", Title: "First", Sort: 2},
			{Url: "/uploads/a2.png", Title: "Second", Sort: 1},
			{Url: "/uploads/a3.png", Title: "Third", Sort: 3},
		},
	})

	detail, err := svc.GetAlbum(ctx, id)
	if err != nil {
		t.Fatalf("get CMS album: %v", err)
	}
	if detail.ImageCount != 3 || len(detail.Images) != 3 {
		t.Fatalf("expected three album images, got count=%d len=%d", detail.ImageCount, len(detail.Images))
	}
	if detail.Images[0].Url != "/uploads/a2.png" {
		t.Fatalf("expected sort-ordered images, got first=%q", detail.Images[0].Url)
	}

	if err = svc.UpdateAlbum(ctx, AlbumSaveInput{
		Id:         id,
		CategoryId: categoryID,
		Name:       "Campus v2",
		Status:     StatusEnabled,
		Images: []AlbumImageInput{
			{Url: "/uploads/b1.png", Title: "New", Sort: 1},
			{Url: "/uploads/b2.png", Title: "New2", Sort: 2},
		},
	}); err != nil {
		t.Fatalf("update CMS album: %v", err)
	}
	updated, err := svc.GetAlbum(ctx, id)
	if err != nil {
		t.Fatalf("get updated CMS album: %v", err)
	}
	if updated.Name != "Campus v2" || updated.ImageCount != 2 || updated.Images[0].Url != "/uploads/b1.png" {
		t.Fatalf("expected whole replacement to two new images, got name=%q count=%d", updated.Name, updated.ImageCount)
	}
}

// TestAlbumImageLimitRejected verifies oversized image lists are rejected and
// leave the album untouched.
func TestAlbumImageLimitRejected(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{code: "limit", status: StatusEnabled, typeID: CategoryTypeAlbum})
	svc := newTestCMSService()
	id := createTestAlbum(t, ctx, svc, AlbumSaveInput{
		CategoryId: categoryID,
		Name:       "Limited",
		Status:     StatusEnabled,
		Images:     []AlbumImageInput{{Url: "/uploads/keep.png", Sort: 1}},
	})

	oversized := make([]AlbumImageInput, AlbumImageMax+1)
	for index := range oversized {
		oversized[index] = AlbumImageInput{Url: fmt.Sprintf("/uploads/over-%d.png", index), Sort: index}
	}
	err := svc.UpdateAlbum(ctx, AlbumSaveInput{Id: id, CategoryId: categoryID, Name: "Limited", Status: StatusEnabled, Images: oversized})
	if !bizerr.Is(err, CodeAlbumImageLimitExceeded) {
		t.Fatalf("expected album image limit error, got %v", err)
	}
	detail, err := svc.GetAlbum(ctx, id)
	if err != nil {
		t.Fatalf("get CMS album after rejected update: %v", err)
	}
	if detail.ImageCount != 1 || detail.Images[0].Url != "/uploads/keep.png" {
		t.Fatalf("expected original image kept after rejected update, got %v", detail.Images)
	}
}

// TestAlbumDeleteCascadesImages verifies album deletion removes its image rows.
func TestAlbumDeleteCascadesImages(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{code: "cascade", status: StatusEnabled, typeID: CategoryTypeAlbum})
	svc := newTestCMSService()
	id := createTestAlbum(t, ctx, svc, AlbumSaveInput{
		CategoryId: categoryID,
		Name:       "Doomed",
		Status:     StatusEnabled,
		Images:     []AlbumImageInput{{Url: "/uploads/d1.png", Sort: 1}, {Url: "/uploads/d2.png", Sort: 2}},
	})

	if err := svc.DeleteAlbum(ctx, id); err != nil {
		t.Fatalf("delete CMS album: %v", err)
	}
	if _, err := svc.GetAlbum(ctx, id); !bizerr.Is(err, CodeAlbumNotFound) {
		t.Fatalf("expected album not found after delete, got %v", err)
	}
	remaining, err := dao.CmsAlbumImage.Ctx(ctx).Where(dao.CmsAlbumImage.Columns().AlbumId, id).Count()
	if err != nil {
		t.Fatalf("count remaining album images: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected cascade-deleted image rows, got %d", remaining)
	}
}

// TestPublicAlbumsHideDisabledContent verifies the public album visibility
// boundary and batched image counts.
func TestPublicAlbumsHideDisabledContent(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	enabledID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{code: "a-visible", status: StatusEnabled, typeID: CategoryTypeAlbum})
	disabledID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{code: "a-hidden", status: StatusDisabled, typeID: CategoryTypeAlbum})
	svc := newTestCMSService()
	visibleID := createTestAlbum(t, ctx, svc, AlbumSaveInput{CategoryId: enabledID, Name: "Visible", Status: StatusEnabled, Images: []AlbumImageInput{{Url: "/uploads/v1.png", Sort: 1}, {Url: "/uploads/v2.png", Sort: 2}}})
	createTestAlbum(t, ctx, svc, AlbumSaveInput{CategoryId: enabledID, Name: "Disabled", Status: StatusDisabled})
	hiddenCategoryAlbumID := createTestAlbum(t, ctx, svc, AlbumSaveInput{CategoryId: disabledID, Name: "Hidden category", Status: StatusEnabled})

	out, err := svc.ListPublicAlbums(ctx, PublicAlbumListInput{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list public CMS albums: %v", err)
	}
	if out.Total != 1 || len(out.List) != 1 || out.List[0].Id != visibleID {
		t.Fatalf("expected single visible public album, got total=%d", out.Total)
	}
	if out.List[0].ImageCount != 2 {
		t.Fatalf("expected batched image count of 2, got %d", out.List[0].ImageCount)
	}
	if _, err = svc.GetPublicAlbum(ctx, hiddenCategoryAlbumID); !bizerr.Is(err, CodePublicContentNotFound) {
		t.Fatalf("expected hidden-category album rejected publicly, got %v", err)
	}
}
