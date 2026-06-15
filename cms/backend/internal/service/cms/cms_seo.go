// This file assembles bounded public SEO projections for sitemap and RSS
// rendering. Queries stay constant per call: one site read, one category
// projection, and one capped article projection.

package cms

import (
	"context"

	"github.com/gogf/gf/v2/os/gtime"

	"lina-plugin-cms/backend/internal/dao"
	entitymodel "lina-plugin-cms/backend/internal/model/entity"
)

// SeoArticleLimitMax caps the article rows a single SEO projection may load.
const SeoArticleLimitMax = 5000

// SeoArticleItem projects the article fields needed by sitemap and RSS output.
type SeoArticleItem struct {
	Id          int64
	Slug        string
	Title       string
	Summary     string
	PublishedAt *gtime.Time
	UpdatedAt   *gtime.Time
}

// SeoCategoryItem projects the category fields needed by sitemap output.
type SeoCategoryItem struct {
	Code      string
	Path      string
	Type      int
	UpdatedAt *gtime.Time
}

// SeoContentOutput bundles the public site settings with bounded category,
// article, and product projections for SEO endpoints.
type SeoContentOutput struct {
	Site       *SiteItem
	Categories []*SeoCategoryItem
	Articles   []*SeoArticleItem
	Products   []*SeoArticleItem
}

// GetPublicSeoContent loads the public SEO projection set: the site settings,
// all enabled non-external categories, the newest publicly visible articles,
// and the newest publicly visible products, each capped at articleLimit
// (clamped to [1, SeoArticleLimitMax]).
func (s *serviceImpl) GetPublicSeoContent(ctx context.Context, articleLimit int) (*SeoContentOutput, error) {
	site, err := s.GetSite(ctx, true)
	if err != nil {
		return nil, err
	}
	categories, err := s.listSeoCategories(ctx)
	if err != nil {
		return nil, err
	}
	limit := clampSeoArticleLimit(articleLimit)
	articles, err := s.listSeoArticles(ctx, limit)
	if err != nil {
		return nil, err
	}
	products, err := s.listSeoProducts(ctx, limit)
	if err != nil {
		return nil, err
	}
	return &SeoContentOutput{Site: site, Categories: categories, Articles: articles, Products: products}, nil
}

// listSeoCategories loads enabled list, single-page, product, and album
// categories as a sorted projection for sitemap URLs.
func (s *serviceImpl) listSeoCategories(ctx context.Context) ([]*SeoCategoryItem, error) {
	columns := dao.CmsCategory.Columns()
	rows := make([]*entitymodel.CmsCategory, 0)
	err := dao.CmsCategory.Ctx(ctx).
		Fields(columns.Code, columns.Path, columns.Type, columns.UpdatedAt).
		Where(columns.Status, StatusEnabled).
		WhereIn(columns.Type, []int{CategoryTypeList, CategoryTypeSingle, CategoryTypeProduct, CategoryTypeAlbum}).
		OrderAsc(columns.Sort).OrderAsc(columns.Id).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	items := make([]*SeoCategoryItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &SeoCategoryItem{Code: row.Code, Path: row.Path, Type: row.Type, UpdatedAt: row.UpdatedAt})
	}
	return items, nil
}

// listSeoArticles loads the newest publicly visible articles as a bounded
// projection ordered by publication time.
func (s *serviceImpl) listSeoArticles(ctx context.Context, limit int) ([]*SeoArticleItem, error) {
	columns := dao.CmsArticle.Columns()
	rows := make([]*entitymodel.CmsArticle, 0, limit)
	err := s.applyPublicArticleVisibility(ctx, dao.CmsArticle.Ctx(ctx), false).
		Fields(columns.Id, columns.Slug, columns.Title, columns.Summary, columns.PublishedAt, columns.UpdatedAt).
		OrderDesc(columns.PublishedAt).OrderDesc(columns.Id).
		Limit(limit).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	items := make([]*SeoArticleItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &SeoArticleItem{Id: row.Id, Slug: row.Slug, Title: row.Title, Summary: row.Summary, PublishedAt: row.PublishedAt, UpdatedAt: row.UpdatedAt})
	}
	return items, nil
}

// clampSeoArticleLimit keeps SEO article loads inside the fixed projection cap.
func clampSeoArticleLimit(limit int) int {
	if limit < 1 {
		return 1
	}
	if limit > SeoArticleLimitMax {
		return SeoArticleLimitMax
	}
	return limit
}

// listSeoProducts loads the newest publicly visible products as a bounded
// projection ordered by publication time; the slug feeds product detail URLs.
func (s *serviceImpl) listSeoProducts(ctx context.Context, limit int) ([]*SeoArticleItem, error) {
	columns := dao.CmsProduct.Columns()
	rows := make([]*entitymodel.CmsProduct, 0, limit)
	err := s.applyPublicProductVisibility(ctx, dao.CmsProduct.Ctx(ctx)).
		Fields(columns.Id, columns.Slug, columns.Name, columns.Summary, columns.PublishedAt, columns.UpdatedAt).
		OrderDesc(columns.PublishedAt).OrderDesc(columns.Id).
		Limit(limit).
		Scan(&rows)
	if err != nil {
		return nil, err
	}
	items := make([]*SeoArticleItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, &SeoArticleItem{Id: row.Id, Slug: row.Slug, Title: row.Name, Summary: row.Summary, PublishedAt: row.PublishedAt, UpdatedAt: row.UpdatedAt})
	}
	return items, nil
}
