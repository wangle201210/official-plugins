// This file implements CMS product management, public product reads, and
// product query helpers. Gallery images travel as a JSON array text column on
// the product row, so list and detail assembly never issues per-row queries.

package cms

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/os/gtime"

	"lina-core/pkg/bizerr"
	"lina-plugin-cms/backend/internal/dao"
	"lina-plugin-cms/backend/internal/model/do"
	entitymodel "lina-plugin-cms/backend/internal/model/entity"
)

// ProductGalleryMax caps the number of gallery images per product.
const ProductGalleryMax = 9

// ListProducts returns paged management products with filters and category names.
func (s *serviceImpl) ListProducts(ctx context.Context, in ProductListInput) (*ProductListOutput, error) {
	columns := dao.CmsProduct.Columns()
	model := dao.CmsProduct.Ctx(ctx)
	if in.CategoryId > 0 {
		model = model.Where(columns.CategoryId, in.CategoryId)
	}
	if in.Status != nil {
		model = model.Where(columns.Status, *in.Status)
	}
	if in.Name != "" {
		model = model.WhereLike(columns.Name, "%"+in.Name+"%")
	}
	model = model.OrderDesc(columns.IsTop).OrderAsc(columns.Sort).OrderDesc(columns.Id)
	return s.scanProductPage(ctx, model, in.PageNum, in.PageSize)
}

// GetProduct returns one management product with its category name.
func (s *serviceImpl) GetProduct(ctx context.Context, id int64) (*ProductItem, error) {
	var product *entitymodel.CmsProduct
	if err := dao.CmsProduct.Ctx(ctx).Where(dao.CmsProduct.Columns().Id, id).Scan(&product); err != nil {
		return nil, err
	}
	if product == nil {
		return nil, bizerr.NewCode(CodeProductNotFound)
	}
	return s.wrapProductItem(ctx, product)
}

// CreateProduct validates and creates one CMS product.
func (s *serviceImpl) CreateProduct(ctx context.Context, in ProductSaveInput) (int64, error) {
	if err := s.ensureCategoryExists(ctx, in.CategoryId); err != nil {
		return 0, err
	}
	if err := s.ensureProductSlugAvailable(ctx, in.Slug, 0); err != nil {
		return 0, err
	}
	userID := s.currentUserID(ctx)
	data := do.CmsProduct{CategoryId: in.CategoryId, Name: in.Name, Slug: in.Slug, Summary: in.Summary, Cover: in.Cover, Gallery: encodeProductGallery(in.Gallery), Price: in.Price, Spec: in.Spec, Content: in.Content, Keywords: in.Keywords, Description: in.Description, Sort: in.Sort, Status: in.Status, IsTop: in.IsTop, IsRecommend: in.IsRecommend, PublishedAt: publishedAtForStatus(in.Status, in.PublishedAt, nil), CreatedBy: userID, UpdatedBy: userID}
	return dao.CmsProduct.Ctx(ctx).Data(data).InsertAndGetId()
}

// UpdateProduct validates and updates one CMS product.
func (s *serviceImpl) UpdateProduct(ctx context.Context, in ProductSaveInput) error {
	columns := dao.CmsProduct.Columns()
	var oldProduct *entitymodel.CmsProduct
	if err := dao.CmsProduct.Ctx(ctx).Where(columns.Id, in.Id).Scan(&oldProduct); err != nil {
		return err
	}
	if oldProduct == nil {
		return bizerr.NewCode(CodeProductNotFound)
	}
	if err := s.ensureCategoryExists(ctx, in.CategoryId); err != nil {
		return err
	}
	if err := s.ensureProductSlugAvailable(ctx, in.Slug, in.Id); err != nil {
		return err
	}
	_, err := dao.CmsProduct.Ctx(ctx).Where(columns.Id, in.Id).Data(do.CmsProduct{CategoryId: in.CategoryId, Name: in.Name, Slug: in.Slug, Summary: in.Summary, Cover: in.Cover, Gallery: encodeProductGallery(in.Gallery), Price: in.Price, Spec: in.Spec, Content: in.Content, Keywords: in.Keywords, Description: in.Description, Sort: in.Sort, Status: in.Status, IsTop: in.IsTop, IsRecommend: in.IsRecommend, PublishedAt: publishedAtForStatus(in.Status, in.PublishedAt, oldProduct.PublishedAt), UpdatedBy: s.currentUserID(ctx)}).Update()
	return err
}

// DeleteProduct removes one CMS product.
func (s *serviceImpl) DeleteProduct(ctx context.Context, id int64) error {
	if _, err := s.GetProduct(ctx, id); err != nil {
		return err
	}
	_, err := dao.CmsProduct.Ctx(ctx).Where(dao.CmsProduct.Columns().Id, id).Delete()
	return err
}

// ListPublicProducts returns publicly visible products ordered for storefront lists.
func (s *serviceImpl) ListPublicProducts(ctx context.Context, in PublicProductListInput) (*ProductListOutput, error) {
	columns := dao.CmsProduct.Columns()
	model := s.applyPublicProductVisibility(ctx, dao.CmsProduct.Ctx(ctx))
	if in.CategoryId > 0 {
		model = model.Where(columns.CategoryId, in.CategoryId)
	}
	model = model.OrderDesc(columns.IsTop).OrderAsc(columns.Sort).OrderDesc(columns.PublishedAt).OrderDesc(columns.Id)
	return s.scanProductPage(ctx, model, in.PageNum, in.PageSize)
}

// GetPublicProductBySlug returns one publicly visible product by slug and increments views.
func (s *serviceImpl) GetPublicProductBySlug(ctx context.Context, slug string) (*ProductItem, error) {
	columns := dao.CmsProduct.Columns()
	var product *entitymodel.CmsProduct
	err := s.applyPublicProductVisibility(ctx, dao.CmsProduct.Ctx(ctx)).Where(columns.Slug, strings.TrimSpace(slug)).Scan(&product)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, bizerr.NewCode(CodePublicContentNotFound)
	}
	if _, err = dao.CmsProduct.Ctx(ctx).Where(columns.Id, product.Id).Data(do.CmsProduct{Views: gdb.Raw(columns.Views + " + 1")}).Update(); err != nil {
		return nil, err
	}
	product.Views++
	return s.wrapProductItem(ctx, product)
}

// applyPublicProductVisibility filters published, released products under enabled categories.
func (s *serviceImpl) applyPublicProductVisibility(ctx context.Context, model *gdb.Model) *gdb.Model {
	productColumns := dao.CmsProduct.Columns()
	categoryColumns := dao.CmsCategory.Columns()
	enabledCategorySubQuery := dao.CmsCategory.Ctx(ctx).Fields(categoryColumns.Id).Where(categoryColumns.Status, StatusEnabled)
	return model.
		Where(productColumns.Status, ArticleStatusPublished).
		WhereLTE(productColumns.PublishedAt, gtime.Now()).
		Where(productColumns.CategoryId+" IN (?)", enabledCategorySubQuery)
}

// scanProductPage scans a paged product query and wraps category names.
func (s *serviceImpl) scanProductPage(ctx context.Context, model *gdb.Model, pageNum int, pageSize int) (*ProductListOutput, error) {
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	list := make([]*entitymodel.CmsProduct, 0)
	if err = model.Page(normalizePageNum(pageNum), normalizePageSize(pageSize)).Scan(&list); err != nil {
		return nil, err
	}
	categoryNames, err := s.categoryNameMapByIDs(ctx, productCategoryIDs(list))
	if err != nil {
		return nil, err
	}
	items := make([]*ProductItem, 0, len(list))
	for _, product := range list {
		items = append(items, &ProductItem{CmsProduct: product, CategoryName: categoryNames[product.CategoryId], Gallery: decodeProductGallery(product.Gallery)})
	}
	return &ProductListOutput{List: items, Total: total}, nil
}

// wrapProductItem attaches the category name and decoded gallery to one product row.
func (s *serviceImpl) wrapProductItem(ctx context.Context, product *entitymodel.CmsProduct) (*ProductItem, error) {
	categoryNames, err := s.categoryNameMapByIDs(ctx, []int64{product.CategoryId})
	if err != nil {
		return nil, err
	}
	return &ProductItem{CmsProduct: product, CategoryName: categoryNames[product.CategoryId], Gallery: decodeProductGallery(product.Gallery)}, nil
}

// productCategoryIDs collects distinct category IDs from product rows.
func productCategoryIDs(list []*entitymodel.CmsProduct) []int64 {
	ids := make([]int64, 0, len(list))
	seen := make(map[int64]bool, len(list))
	for _, product := range list {
		if product == nil || product.CategoryId <= 0 || seen[product.CategoryId] {
			continue
		}
		ids = append(ids, product.CategoryId)
		seen[product.CategoryId] = true
	}
	return ids
}

// ensureProductSlugAvailable prevents duplicate product slugs.
func (s *serviceImpl) ensureProductSlugAvailable(ctx context.Context, slug string, currentID int64) error {
	columns := dao.CmsProduct.Columns()
	model := dao.CmsProduct.Ctx(ctx).Where(columns.Slug, strings.TrimSpace(slug))
	if currentID > 0 {
		model = model.WhereNot(columns.Id, currentID)
	}
	count, err := model.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewCode(CodeProductSlugExists)
	}
	return nil
}

// encodeProductGallery serializes gallery URLs to the JSON array text column,
// dropping blanks and clamping to the gallery cap.
func encodeProductGallery(urls []string) string {
	cleaned := make([]string, 0, len(urls))
	for _, url := range urls {
		url = strings.TrimSpace(url)
		if url == "" {
			continue
		}
		cleaned = append(cleaned, url)
		if len(cleaned) >= ProductGalleryMax {
			break
		}
	}
	encoded, err := gjson.New(cleaned).ToJsonString()
	if err != nil {
		return "[]"
	}
	return encoded
}

// decodeProductGallery parses the JSON array text column into gallery URLs;
// malformed content degrades to an empty list.
func decodeProductGallery(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" || value == "[]" {
		return []string{}
	}
	parsed, err := gjson.LoadContent([]byte(value))
	if err != nil {
		return []string{}
	}
	urls := parsed.Var().Strings()
	if len(urls) > ProductGalleryMax {
		urls = urls[:ProductGalleryMax]
	}
	return urls
}
