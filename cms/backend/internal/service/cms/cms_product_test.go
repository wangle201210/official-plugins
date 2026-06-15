// This file verifies CMS product management, visibility, and gallery encoding
// against an isolated PostgreSQL schema.

package cms

import (
	"context"
	"testing"
	"time"

	"lina-core/pkg/bizerr"
)

// createTestProduct creates one product through the service for test setup.
func createTestProduct(t *testing.T, ctx context.Context, svc Service, in ProductSaveInput) int64 {
	t.Helper()

	id, err := svc.CreateProduct(ctx, in)
	if err != nil {
		t.Fatalf("create CMS product %s: %v", in.Slug, err)
	}
	return id
}

// TestProductCRUDAndSlugConflict verifies the management product lifecycle and
// duplicate slug rejection.
func TestProductCRUDAndSlugConflict(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{code: "products", status: StatusEnabled, typeID: CategoryTypeProduct})
	svc := newTestCMSService()
	id := createTestProduct(t, ctx, svc, ProductSaveInput{
		CategoryId: categoryID,
		Name:       "Thermal pad",
		Slug:       "thermal-pad",
		Price:      "$99",
		Spec:       "8 W/(m·K)",
		Content:    "<p>Detail</p>",
		Gallery:    []string{"/uploads/p1.png", " ", "/uploads/p2.png"},
		Status:     ArticleStatusPublished,
	})

	detail, err := svc.GetProduct(ctx, id)
	if err != nil {
		t.Fatalf("get CMS product: %v", err)
	}
	if detail.CategoryName != "products" || detail.Price != "$99" {
		t.Fatalf("expected wrapped product detail, got category=%q price=%q", detail.CategoryName, detail.Price)
	}
	if len(detail.Gallery) != 2 || detail.Gallery[0] != "/uploads/p1.png" {
		t.Fatalf("expected cleaned gallery of two images, got %v", detail.Gallery)
	}
	if detail.PublishedAt == nil {
		t.Fatalf("expected first publish to stamp publication time")
	}

	if _, err = svc.CreateProduct(ctx, ProductSaveInput{CategoryId: categoryID, Name: "Dup", Slug: "thermal-pad", Content: "<p>x</p>", Status: ArticleStatusDraft}); !bizerr.Is(err, CodeProductSlugExists) {
		t.Fatalf("expected duplicate product slug error, got %v", err)
	}

	if err = svc.UpdateProduct(ctx, ProductSaveInput{Id: id, CategoryId: categoryID, Name: "Thermal pad v2", Slug: "thermal-pad", Content: "<p>v2</p>", Gallery: []string{"/uploads/p3.png"}, Status: ArticleStatusPublished}); err != nil {
		t.Fatalf("update CMS product: %v", err)
	}
	updated, err := svc.GetProduct(ctx, id)
	if err != nil {
		t.Fatalf("get updated CMS product: %v", err)
	}
	if updated.Name != "Thermal pad v2" || len(updated.Gallery) != 1 {
		t.Fatalf("expected updated product with one gallery image, got name=%q gallery=%v", updated.Name, updated.Gallery)
	}

	if err = svc.DeleteProduct(ctx, id); err != nil {
		t.Fatalf("delete CMS product: %v", err)
	}
	if _, err = svc.GetProduct(ctx, id); !bizerr.Is(err, CodeProductNotFound) {
		t.Fatalf("expected product not found after delete, got %v", err)
	}
}

// TestPublicProductsHideDraftsScheduledAndDisabledCategories verifies the
// public product visibility boundary.
func TestPublicProductsHideDraftsScheduledAndDisabledCategories(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	enabledID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{code: "p-visible", status: StatusEnabled, typeID: CategoryTypeProduct})
	disabledID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{code: "p-hidden", status: StatusDisabled, typeID: CategoryTypeProduct})
	svc := newTestCMSService()
	futureMillis := time.Now().Add(time.Hour).UnixMilli()
	createTestProduct(t, ctx, svc, ProductSaveInput{CategoryId: enabledID, Name: "Visible", Slug: "p-visible", Content: "<p>v</p>", Status: ArticleStatusPublished})
	createTestProduct(t, ctx, svc, ProductSaveInput{CategoryId: enabledID, Name: "Draft", Slug: "p-draft", Content: "<p>d</p>", Status: ArticleStatusDraft})
	createTestProduct(t, ctx, svc, ProductSaveInput{CategoryId: enabledID, Name: "Scheduled", Slug: "p-scheduled", Content: "<p>s</p>", Status: ArticleStatusPublished, PublishedAt: &futureMillis})
	createTestProduct(t, ctx, svc, ProductSaveInput{CategoryId: disabledID, Name: "Hidden", Slug: "p-hidden-cat", Content: "<p>h</p>", Status: ArticleStatusPublished})

	out, err := svc.ListPublicProducts(ctx, PublicProductListInput{PageNum: 1, PageSize: 20})
	if err != nil {
		t.Fatalf("list public CMS products: %v", err)
	}
	if out.Total != 1 || len(out.List) != 1 || out.List[0].Slug != "p-visible" {
		t.Fatalf("expected single visible public product, got total=%d list=%v", out.Total, out.List)
	}
	if _, err = svc.GetPublicProductBySlug(ctx, "p-scheduled"); !bizerr.Is(err, CodePublicContentNotFound) {
		t.Fatalf("expected scheduled product hidden from public detail, got %v", err)
	}
}

// TestPublicProductDetailIncrementsViews verifies public detail view counting.
func TestPublicProductDetailIncrementsViews(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{code: "p-views", status: StatusEnabled, typeID: CategoryTypeProduct})
	svc := newTestCMSService()
	createTestProduct(t, ctx, svc, ProductSaveInput{CategoryId: categoryID, Name: "Counter", Slug: "p-counter", Content: "<p>c</p>", Status: ArticleStatusPublished})

	first, err := svc.GetPublicProductBySlug(ctx, "p-counter")
	if err != nil {
		t.Fatalf("first public product read: %v", err)
	}
	second, err := svc.GetPublicProductBySlug(ctx, "p-counter")
	if err != nil {
		t.Fatalf("second public product read: %v", err)
	}
	if second.Views != first.Views+1 {
		t.Fatalf("expected view count to grow from %d, got %d", first.Views, second.Views)
	}
}
