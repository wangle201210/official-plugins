// This file verifies the bounded public SEO projections used by sitemap and
// RSS rendering.

package cms

import (
	"context"
	"testing"
	"time"
)

// TestGetPublicSeoContentFiltersHiddenContent verifies SEO projections expose
// only enabled non-external categories and publicly visible articles.
func TestGetPublicSeoContentFiltersHiddenContent(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	enabledID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{code: "seo-news", status: StatusEnabled, typeID: CategoryTypeList})
	insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{code: "seo-hidden", status: StatusDisabled, typeID: CategoryTypeList})
	insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{code: "seo-outlink", status: StatusEnabled, typeID: CategoryTypeExternal})
	insertCMSArticle(t, ctx, enabledID, "seo-visible", ArticleStatusPublished)
	insertCMSArticle(t, ctx, enabledID, "seo-draft", ArticleStatusDraft)
	futureMillis := time.Now().Add(time.Hour).UnixMilli()
	svc := newTestCMSService()
	if _, err := svc.CreateArticle(ctx, ArticleSaveInput{
		CategoryId:  enabledID,
		Title:       "Scheduled",
		Slug:        "seo-scheduled",
		Content:     "<p>Scheduled</p>",
		Status:      ArticleStatusPublished,
		PublishedAt: &futureMillis,
	}); err != nil {
		t.Fatalf("create scheduled CMS article: %v", err)
	}

	out, err := svc.GetPublicSeoContent(ctx, 10)
	if err != nil {
		t.Fatalf("load CMS SEO content: %v", err)
	}
	if out.Site == nil {
		t.Fatalf("expected site settings in SEO content")
	}
	if len(out.Categories) != 1 || out.Categories[0].Code != "seo-news" {
		t.Fatalf("expected only the enabled list category, got %+v", out.Categories)
	}
	if len(out.Articles) != 1 || out.Articles[0].Slug != "seo-visible" {
		t.Fatalf("expected only the visible published article, got %+v", out.Articles)
	}
	if out.Articles[0].PublishedAt == nil {
		t.Fatalf("expected projected publication time for SEO article")
	}
}

// TestGetPublicSeoContentIncludesProductsAndNewCategoryTypes verifies SEO
// projections cover product/album categories and published products while
// hiding scheduled products.
func TestGetPublicSeoContentIncludesProductsAndNewCategoryTypes(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	productCategoryID := insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{code: "seo-products", status: StatusEnabled, typeID: CategoryTypeProduct})
	insertCMSCategoryWithOptions(t, ctx, cmsCategoryOptions{code: "seo-albums", status: StatusEnabled, typeID: CategoryTypeAlbum})
	svc := newTestCMSService()
	futureMillis := time.Now().Add(time.Hour).UnixMilli()
	if _, err := svc.CreateProduct(ctx, ProductSaveInput{CategoryId: productCategoryID, Name: "SEO visible", Slug: "seo-product-visible", Content: "<p>v</p>", Status: ArticleStatusPublished}); err != nil {
		t.Fatalf("create visible SEO product: %v", err)
	}
	if _, err := svc.CreateProduct(ctx, ProductSaveInput{CategoryId: productCategoryID, Name: "SEO scheduled", Slug: "seo-product-scheduled", Content: "<p>s</p>", Status: ArticleStatusPublished, PublishedAt: &futureMillis}); err != nil {
		t.Fatalf("create scheduled SEO product: %v", err)
	}

	out, err := svc.GetPublicSeoContent(ctx, 10)
	if err != nil {
		t.Fatalf("load CMS SEO content with products: %v", err)
	}
	codes := make(map[string]bool, len(out.Categories))
	for _, category := range out.Categories {
		codes[category.Code] = true
	}
	if !codes["seo-products"] || !codes["seo-albums"] {
		t.Fatalf("expected product and album categories in SEO projection, got %+v", out.Categories)
	}
	if len(out.Products) != 1 || out.Products[0].Slug != "seo-product-visible" {
		t.Fatalf("expected only the visible published product, got %+v", out.Products)
	}
}

// TestGetPublicSeoContentClampsArticleLimit verifies the article limit is
// clamped to a sane bounded window.
func TestGetPublicSeoContentClampsArticleLimit(t *testing.T) {
	ctx := context.Background()
	setupSQLiteCMSDB(t, ctx)

	categoryID := insertCMSCategory(t, ctx, "seo-limit", StatusEnabled)
	insertCMSArticle(t, ctx, categoryID, "seo-limit-first", ArticleStatusPublished)
	insertCMSArticle(t, ctx, categoryID, "seo-limit-second", ArticleStatusPublished)

	out, err := newTestCMSService().GetPublicSeoContent(ctx, 0)
	if err != nil {
		t.Fatalf("load CMS SEO content with zero limit: %v", err)
	}
	if len(out.Articles) != 1 {
		t.Fatalf("expected zero limit clamped to one article, got %d", len(out.Articles))
	}
}
