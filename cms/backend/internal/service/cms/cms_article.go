// This file implements CMS article management, public article reads, and article query helpers.

package cms

import (
	"context"
	"fmt"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/os/gtime"
	"html"
	"lina-core/pkg/bizerr"
	"lina-plugin-cms/backend/internal/dao"
	"lina-plugin-cms/backend/internal/model/do"
	entitymodel "lina-plugin-cms/backend/internal/model/entity"
	"strings"
	"time"
)

// ListArticles returns paged management articles with filters and category names.
func (s *serviceImpl) ListArticles(ctx context.Context, in ArticleListInput) (*ArticleListOutput, error) {
	model, err := s.applyArticleManagementFilters(ctx, dao.CmsArticle.Ctx(ctx), in)
	if err != nil {
		return nil, err
	}
	return s.scanArticlePage(ctx, model, in.PageNum, in.PageSize)
}

// GetArticle returns one management article with its category name.
func (s *serviceImpl) GetArticle(ctx context.Context, id int64) (*ArticleItem, error) {
	var article *entitymodel.CmsArticle
	if err := dao.CmsArticle.Ctx(ctx).Where(dao.CmsArticle.Columns().Id, id).Scan(&article); err != nil {
		return nil, err
	}
	if article == nil {
		return nil, bizerr.NewCode(CodeArticleNotFound)
	}
	return s.wrapArticleItem(ctx, article)
}

// CreateArticle validates and creates one CMS article.
func (s *serviceImpl) CreateArticle(ctx context.Context, in ArticleSaveInput) (int64, error) {
	if err := s.ensureCategoryExists(ctx, in.CategoryId); err != nil {
		return 0, err
	}
	if err := s.ensureArticleSlugAvailable(ctx, in.Slug, 0); err != nil {
		return 0, err
	}
	userID := s.currentUserID(ctx)
	data := do.CmsArticle{CategoryId: in.CategoryId, Title: in.Title, Subtitle: in.Subtitle, Slug: in.Slug, Summary: in.Summary, Cover: in.Cover, Author: in.Author, Source: in.Source, Content: in.Content, Tags: in.Tags, Keywords: in.Keywords, Description: in.Description, Sort: in.Sort, Status: in.Status, IsTop: in.IsTop, IsRecommend: in.IsRecommend, PublishedAt: publishedAtForStatus(in.Status, in.PublishedAt, nil), CreatedBy: userID, UpdatedBy: userID}
	return dao.CmsArticle.Ctx(ctx).Data(data).InsertAndGetId()
}

// UpdateArticle validates and updates one CMS article.
func (s *serviceImpl) UpdateArticle(ctx context.Context, in ArticleSaveInput) error {
	columns := dao.CmsArticle.Columns()
	var oldArticle *entitymodel.CmsArticle
	if err := dao.CmsArticle.Ctx(ctx).Where(columns.Id, in.Id).Scan(&oldArticle); err != nil {
		return err
	}
	if oldArticle == nil {
		return bizerr.NewCode(CodeArticleNotFound)
	}
	if err := s.ensureCategoryExists(ctx, in.CategoryId); err != nil {
		return err
	}
	if err := s.ensureArticleSlugAvailable(ctx, in.Slug, in.Id); err != nil {
		return err
	}
	_, err := dao.CmsArticle.Ctx(ctx).Where(columns.Id, in.Id).Data(do.CmsArticle{CategoryId: in.CategoryId, Title: in.Title, Subtitle: in.Subtitle, Slug: in.Slug, Summary: in.Summary, Cover: in.Cover, Author: in.Author, Source: in.Source, Content: in.Content, Tags: in.Tags, Keywords: in.Keywords, Description: in.Description, Sort: in.Sort, Status: in.Status, IsTop: in.IsTop, IsRecommend: in.IsRecommend, PublishedAt: publishedAtForStatus(in.Status, in.PublishedAt, oldArticle.PublishedAt), UpdatedBy: s.currentUserID(ctx)}).Update()
	return err
}

// DeleteArticle removes one CMS article and its tag rows.
func (s *serviceImpl) DeleteArticle(ctx context.Context, id int64) error {
	columns := dao.CmsArticle.Columns()
	if _, err := s.GetArticle(ctx, id); err != nil {
		return err
	}
	_, err := dao.CmsArticle.Ctx(ctx).Where(columns.Id, id).Delete()
	return err
}

// ArticleBatchMax caps the number of article IDs one batch operation accepts.
const ArticleBatchMax = 100

// BatchUpdateArticleStatus publishes or unpublishes multiple CMS articles in one
// transaction with collection statements, rejecting the whole batch when any
// target article does not exist or the batch exceeds ArticleBatchMax.
func (s *serviceImpl) BatchUpdateArticleStatus(ctx context.Context, in ArticleBatchStatusInput) error {
	ids := normalizeArticleBatchIDs(in.Ids)
	if len(ids) > ArticleBatchMax {
		return bizerr.NewCode(CodeArticleBatchLimitExceeded)
	}
	columns := dao.CmsArticle.Columns()
	userID := s.currentUserID(ctx)
	return dao.CmsArticle.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if err := s.ensureArticlesExist(ctx, ids); err != nil {
			return err
		}
		if _, err := dao.CmsArticle.Ctx(ctx).WhereIn(columns.Id, ids).Data(do.CmsArticle{Status: in.Status, UpdatedBy: userID}).Update(); err != nil {
			return err
		}
		if in.Status != ArticleStatusPublished {
			return nil
		}
		_, err := dao.CmsArticle.Ctx(ctx).WhereIn(columns.Id, ids).WhereNull(columns.PublishedAt).Data(do.CmsArticle{PublishedAt: gtime.Now(), UpdatedBy: userID}).Update()
		return err
	})
}

// BatchDeleteArticles soft deletes multiple CMS articles with one collection
// statement, rejecting the whole batch when any target article does not exist
// or the batch exceeds ArticleBatchMax.
func (s *serviceImpl) BatchDeleteArticles(ctx context.Context, ids []int64) error {
	ids = normalizeArticleBatchIDs(ids)
	if len(ids) > ArticleBatchMax {
		return bizerr.NewCode(CodeArticleBatchLimitExceeded)
	}
	columns := dao.CmsArticle.Columns()
	return dao.CmsArticle.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if err := s.ensureArticlesExist(ctx, ids); err != nil {
			return err
		}
		_, err := dao.CmsArticle.Ctx(ctx).WhereIn(columns.Id, ids).Delete()
		return err
	})
}

// normalizeArticleBatchIDs drops non-positive IDs and duplicates so existence
// counting stays exact.
func normalizeArticleBatchIDs(ids []int64) []int64 {
	result := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// ensureArticlesExist rejects a batch when any target article is missing,
// using one counting query for the whole ID set.
func (s *serviceImpl) ensureArticlesExist(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return bizerr.NewCode(CodeArticleNotFound)
	}
	count, err := dao.CmsArticle.Ctx(ctx).WhereIn(dao.CmsArticle.Columns().Id, ids).Count()
	if err != nil {
		return err
	}
	if count != len(ids) {
		return bizerr.NewCode(CodeArticleNotFound)
	}
	return nil
}

// ListPublicArticles returns visible public articles with template ordering rules applied.
func (s *serviceImpl) ListPublicArticles(ctx context.Context, in PublicArticleListInput) (*ArticleListOutput, error) {
	columns := dao.CmsArticle.Columns()
	model := s.applyPublicArticleVisibility(ctx, dao.CmsArticle.Ctx(ctx), in.IncludeHiddenCategories)
	if in.CategoryId > 0 {
		model = model.Where(columns.CategoryId, in.CategoryId)
	}
	if keywordText := strings.TrimSpace(in.Keyword); keywordText != "" {
		keyword := "%" + keywordText + "%"
		model = model.Where(fmt.Sprintf("(%s LIKE ? OR %s LIKE ? OR %s LIKE ? OR %s LIKE ? OR %s LIKE ? OR %s LIKE ? OR %s LIKE ?)", columns.Title, columns.Subtitle, columns.Summary, columns.Content, columns.Tags, columns.Keywords, columns.Description), keyword, keyword, keyword, keyword, keyword, keyword, keyword)
	}
	model = s.applyPublicArticleOrder(model, in.Order)
	return s.scanArticlePage(ctx, model, in.PageNum, in.PageSize)
}

// GetPublicArticleBySlug returns one visible public article by slug and increments views.
func (s *serviceImpl) GetPublicArticleBySlug(ctx context.Context, slug string) (*ArticleItem, error) {
	columns := dao.CmsArticle.Columns()
	var article *entitymodel.CmsArticle
	err := s.applyPublicArticleVisibility(ctx, dao.CmsArticle.Ctx(ctx), true).Where(columns.Slug, strings.TrimSpace(slug)).Scan(&article)
	if err != nil {
		return nil, err
	}
	if article == nil {
		return nil, bizerr.NewCode(CodePublicContentNotFound)
	}
	if _, err = dao.CmsArticle.Ctx(ctx).Where(columns.Id, article.Id).Data(do.CmsArticle{Views: gdb.Raw(columns.Views + " + 1")}).Update(); err != nil {
		return nil, err
	}
	article.Views++
	return s.wrapArticleItem(ctx, article)
}

// applyArticleManagementFilters applies management article filters to a query.
func (s *serviceImpl) applyArticleManagementFilters(ctx context.Context, model *gdb.Model, in ArticleListInput) (*gdb.Model, error) {
	columns := dao.CmsArticle.Columns()
	if in.CategoryId > 0 {
		ids, err := s.categoryIDsForArticleFilter(ctx, in.CategoryId, in.IncludeChildren)
		if err != nil {
			return nil, err
		}
		model = model.WhereIn(columns.CategoryId, ids)
	} else if in.CategoryType > 0 {
		categoryColumns := dao.CmsCategory.Columns()
		subQuery := dao.CmsCategory.Ctx(ctx).Fields(categoryColumns.Id).Where(categoryColumns.Type, in.CategoryType)
		model = model.Where(columns.CategoryId+" IN (?)", subQuery)
	}
	if in.Status != nil {
		model = model.Where(columns.Status, *in.Status)
	}
	if in.Title != "" {
		model = model.WhereLike(columns.Title, "%"+in.Title+"%")
	}
	return model.OrderDesc(columns.IsTop).OrderDesc(columns.PublishedAt).OrderDesc(columns.Id), nil
}

// categoryIDsForArticleFilter expands category filters when child categories are included.
func (s *serviceImpl) categoryIDsForArticleFilter(ctx context.Context, categoryID int64, includeChildren bool) ([]int64, error) {
	if categoryID <= 0 {
		return nil, nil
	}
	if !includeChildren {
		return []int64{categoryID}, nil
	}
	categories := make([]*entitymodel.CmsCategory, 0)
	columns := dao.CmsCategory.Columns()
	if err := dao.CmsCategory.Ctx(ctx).Fields(columns.Id, columns.ParentId).Scan(&categories); err != nil {
		return nil, err
	}
	childrenByParent := make(map[int64][]int64)
	for _, category := range categories {
		childrenByParent[category.ParentId] = append(childrenByParent[category.ParentId], category.Id)
	}
	ids := []int64{categoryID}
	queue := []int64{categoryID}
	visited := map[int64]struct{}{categoryID: {}}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		for _, childID := range childrenByParent[current] {
			if _, ok := visited[childID]; ok {
				continue
			}
			visited[childID] = struct{}{}
			ids = append(ids, childID)
			queue = append(queue, childID)
		}
	}
	return ids, nil
}

// applyPublicArticleVisibility adds public article status and category visibility filters.
// Scheduled articles stay hidden: only rows whose published_at is set and not in
// the future pass (SQL comparison is false for NULL published_at).
func (s *serviceImpl) applyPublicArticleVisibility(ctx context.Context, model *gdb.Model, includeHiddenCategories bool) *gdb.Model {
	articleColumns := dao.CmsArticle.Columns()
	model = model.Where(articleColumns.Status, ArticleStatusPublished).WhereLTE(articleColumns.PublishedAt, gtime.Now())
	if includeHiddenCategories {
		return model
	}
	categoryColumns := dao.CmsCategory.Columns()
	enabledCategorySubQuery := dao.CmsCategory.Ctx(ctx).Fields(categoryColumns.Id).Where(categoryColumns.Status, StatusEnabled)
	return model.Where(articleColumns.CategoryId+" IN (?)", enabledCategorySubQuery)
}

// applyPublicArticleOrder applies stable public article ordering rules.
func (s *serviceImpl) applyPublicArticleOrder(model *gdb.Model, order PublicArticleOrder) *gdb.Model {
	columns := dao.CmsArticle.Columns()
	switch NormalizePublicArticleOrder(string(order)) {
	case PublicArticleOrderID:
		return model.OrderDesc(columns.Id).OrderDesc(columns.IsTop).OrderDesc(columns.IsRecommend).OrderAsc(columns.Sort).OrderDesc(columns.PublishedAt)
	case PublicArticleOrderDate:
		return model.OrderDesc(columns.IsTop).OrderDesc(columns.PublishedAt).OrderDesc(columns.IsRecommend).OrderAsc(columns.Sort).OrderDesc(columns.Id)
	case PublicArticleOrderManual:
		return model.OrderAsc(columns.Sort).OrderDesc(columns.IsTop).OrderDesc(columns.IsRecommend).OrderDesc(columns.PublishedAt).OrderDesc(columns.Id)
	case PublicArticleOrderViews:
		return model.OrderDesc(columns.Views).OrderDesc(columns.IsTop).OrderDesc(columns.IsRecommend).OrderAsc(columns.Sort).OrderDesc(columns.PublishedAt).OrderDesc(columns.Id)
	default:
		return model.OrderDesc(columns.IsTop).OrderDesc(columns.PublishedAt).OrderDesc(columns.Id)
	}
}

// scanArticlePage scans a paged article query and wraps category names.
func (s *serviceImpl) scanArticlePage(ctx context.Context, model *gdb.Model, pageNum int, pageSize int) (*ArticleListOutput, error) {
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	list := make([]*entitymodel.CmsArticle, 0)
	if err = model.Page(normalizePageNum(pageNum), normalizePageSize(pageSize)).Scan(&list); err != nil {
		return nil, err
	}
	items, err := s.wrapArticleItems(ctx, list)
	if err != nil {
		return nil, err
	}
	return &ArticleListOutput{List: items, Total: total}, nil
}

// wrapArticleItems attaches category names to article rows.
func (s *serviceImpl) wrapArticleItems(ctx context.Context, list []*entitymodel.CmsArticle) ([]*ArticleItem, error) {
	categoryNames, err := s.categoryNameMap(ctx, list)
	if err != nil {
		return nil, err
	}
	items := make([]*ArticleItem, 0, len(list))
	for _, article := range list {
		article.Content = normalizeImportedArticleContent(article.Content)
		items = append(items, &ArticleItem{CmsArticle: article, CategoryName: categoryNames[article.CategoryId]})
	}
	return items, nil
}

// wrapArticleItem attaches the category name to a single article row.
func (s *serviceImpl) wrapArticleItem(ctx context.Context, article *entitymodel.CmsArticle) (*ArticleItem, error) {
	items, err := s.wrapArticleItems(ctx, []*entitymodel.CmsArticle{article})
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return nil, bizerr.NewCode(CodeArticleNotFound)
	}
	return items[0], nil
}

// categoryNameMap loads category names for the supplied article category IDs.
func (s *serviceImpl) categoryNameMap(ctx context.Context, list []*entitymodel.CmsArticle) (map[int64]string, error) {
	ids := make([]int64, 0, len(list))
	seen := make(map[int64]bool)
	for _, article := range list {
		if article == nil || article.CategoryId <= 0 || seen[article.CategoryId] {
			continue
		}
		ids = append(ids, article.CategoryId)
		seen[article.CategoryId] = true
	}
	return s.categoryNameMapByIDs(ctx, ids)
}

// categoryNameMapByIDs loads category names for a distinct category ID set
// with one batched query.
func (s *serviceImpl) categoryNameMapByIDs(ctx context.Context, ids []int64) (map[int64]string, error) {
	result := make(map[int64]string)
	if len(ids) == 0 {
		return result, nil
	}
	columns := dao.CmsCategory.Columns()
	categories := make([]*entitymodel.CmsCategory, 0, len(ids))
	if err := dao.CmsCategory.Ctx(ctx).Fields(columns.Id, columns.Name).WhereIn(columns.Id, ids).Scan(&categories); err != nil {
		return nil, err
	}
	for _, category := range categories {
		result[category.Id] = category.Name
	}
	return result, nil
}

// normalizeImportedArticleContent trims article HTML imported from requests or sample data.
func normalizeImportedArticleContent(content string) string {
	return html.UnescapeString(content)
}

// ensureArticleSlugAvailable prevents duplicate article slugs.
func (s *serviceImpl) ensureArticleSlugAvailable(ctx context.Context, slug string, currentID int64) error {
	columns := dao.CmsArticle.Columns()
	model := dao.CmsArticle.Ctx(ctx).Where(columns.Slug, strings.TrimSpace(slug))
	if currentID > 0 {
		model = model.WhereNot(columns.Id, currentID)
	}
	count, err := model.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewCode(CodeArticleSlugExists)
	}
	return nil
}

// publishedAtForStatus chooses a publication time for published articles. An
// explicit Unix millisecond timestamp wins and may schedule a future
// publication; otherwise the previous time is kept and first publishes fall
// back to the current time. Draft saves return nil so the stored value stays
// untouched.
func publishedAtForStatus(status int, explicitMillis *int64, oldPublishedAt *gtime.Time) *gtime.Time {
	if status != ArticleStatusPublished {
		return nil
	}
	if explicitMillis != nil {
		return gtime.New(time.UnixMilli(*explicitMillis))
	}
	if oldPublishedAt != nil {
		return oldPublishedAt
	}
	return gtime.Now()
}
