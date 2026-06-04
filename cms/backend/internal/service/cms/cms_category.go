// This file implements CMS category management and category tree helpers.

package cms

import (
	"context"
	"lina-core/pkg/bizerr"
	"lina-plugin-cms/backend/internal/dao"
	"lina-plugin-cms/backend/internal/model/do"
	entitymodel "lina-plugin-cms/backend/internal/model/entity"
	"strings"
)

// ListCategories returns the CMS category tree with optional status and public filters.
func (s *serviceImpl) ListCategories(ctx context.Context, in CategoryListInput) ([]*CategoryItem, error) {
	columns := dao.CmsCategory.Columns()
	model := dao.CmsCategory.Ctx(ctx).OrderAsc(columns.Sort).OrderAsc(columns.Id)
	if in.PublicOnly {
		model = model.Where(columns.Status, StatusEnabled)
	} else if in.Status != nil {
		model = model.Where(columns.Status, *in.Status)
	}
	list := make([]*entitymodel.CmsCategory, 0)
	if err := model.Scan(&list); err != nil {
		return nil, err
	}
	return buildCategoryTree(list), nil
}

// CreateCategory validates and creates one CMS category.
func (s *serviceImpl) CreateCategory(ctx context.Context, in CategorySaveInput) (int64, error) {
	if err := s.ensureCategoryCodeAvailable(ctx, in.Code, 0); err != nil {
		return 0, err
	}
	if err := s.ensureCategoryParentAvailable(ctx, 0, in.ParentId); err != nil {
		return 0, err
	}
	userID := s.currentUserID(ctx)
	return dao.CmsCategory.Ctx(ctx).Data(do.CmsCategory{ParentId: in.ParentId, Code: in.Code, Name: in.Name, Type: in.Type, Path: in.Path, ListTemplate: in.ListTemplate, ContentTemplate: in.ContentTemplate, Cover: in.Cover, Outlink: in.Outlink, Title: in.Title, Keywords: in.Keywords, Description: in.Description, Sort: in.Sort, Status: in.Status, CreatedBy: userID, UpdatedBy: userID}).InsertAndGetId()
}

// UpdateCategory validates and updates one CMS category.
func (s *serviceImpl) UpdateCategory(ctx context.Context, in CategorySaveInput) error {
	columns := dao.CmsCategory.Columns()
	if err := s.ensureCategoryExists(ctx, in.Id); err != nil {
		return err
	}
	if err := s.ensureCategoryCodeAvailable(ctx, in.Code, in.Id); err != nil {
		return err
	}
	if err := s.ensureCategoryParentAvailable(ctx, in.Id, in.ParentId); err != nil {
		return err
	}
	_, err := dao.CmsCategory.Ctx(ctx).Where(columns.Id, in.Id).Data(do.CmsCategory{ParentId: in.ParentId, Code: in.Code, Name: in.Name, Type: in.Type, Path: in.Path, ListTemplate: in.ListTemplate, ContentTemplate: in.ContentTemplate, Cover: in.Cover, Outlink: in.Outlink, Title: in.Title, Keywords: in.Keywords, Description: in.Description, Sort: in.Sort, Status: in.Status, UpdatedBy: s.currentUserID(ctx)}).Update()
	return err
}

// DeleteCategory removes one CMS category after child and article usage checks.
func (s *serviceImpl) DeleteCategory(ctx context.Context, id int64) error {
	columns := dao.CmsCategory.Columns()
	articleColumns := dao.CmsArticle.Columns()
	if err := s.ensureCategoryExists(ctx, id); err != nil {
		return err
	}
	children, err := dao.CmsCategory.Ctx(ctx).Where(columns.ParentId, id).Count()
	if err != nil {
		return err
	}
	if children > 0 {
		return bizerr.NewCode(CodeCategoryHasChildren)
	}
	articles, err := dao.CmsArticle.Ctx(ctx).Where(articleColumns.CategoryId, id).Count()
	if err != nil {
		return err
	}
	if articles > 0 {
		return bizerr.NewCode(CodeCategoryHasArticles)
	}
	_, err = dao.CmsCategory.Ctx(ctx).Where(columns.Id, id).Delete()
	return err
}

// ensureCategoryExists verifies that a category exists before mutation.
func (s *serviceImpl) ensureCategoryExists(ctx context.Context, id int64) error {
	columns := dao.CmsCategory.Columns()
	count, err := dao.CmsCategory.Ctx(ctx).Where(columns.Id, id).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return bizerr.NewCode(CodeCategoryNotFound)
	}
	return nil
}

// ensureCategoryCodeAvailable prevents duplicate category codes.
func (s *serviceImpl) ensureCategoryCodeAvailable(ctx context.Context, code string, currentID int64) error {
	columns := dao.CmsCategory.Columns()
	model := dao.CmsCategory.Ctx(ctx).Where(columns.Code, strings.TrimSpace(code))
	if currentID > 0 {
		model = model.WhereNot(columns.Id, currentID)
	}
	count, err := model.Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return bizerr.NewCode(CodeCategoryCodeExists)
	}
	return nil
}

// ensureCategoryParentAvailable validates parent selection and cycle boundaries.
func (s *serviceImpl) ensureCategoryParentAvailable(ctx context.Context, categoryID int64, parentID int64) error {
	if parentID <= 0 {
		return nil
	}
	if categoryID > 0 && parentID == categoryID {
		return bizerr.NewCode(CodeCategoryParentInvalid)
	}
	if err := s.ensureCategoryExists(ctx, parentID); err != nil {
		return err
	}
	if categoryID <= 0 {
		return nil
	}
	columns := dao.CmsCategory.Columns()
	categories := make([]*entitymodel.CmsCategory, 0)
	if err := dao.CmsCategory.Ctx(ctx).Fields(columns.Id, columns.ParentId).Scan(&categories); err != nil {
		return err
	}
	parentByID := make(map[int64]int64, len(categories))
	for _, category := range categories {
		if category == nil {
			continue
		}
		parentByID[category.Id] = category.ParentId
	}
	visited := make(map[int64]struct{}, len(categories))
	for currentID := parentID; currentID > 0; currentID = parentByID[currentID] {
		if currentID == categoryID {
			return bizerr.NewCode(CodeCategoryParentInvalid)
		}
		if _, ok := visited[currentID]; ok {
			return bizerr.NewCode(CodeCategoryParentInvalid)
		}
		visited[currentID] = struct{}{}
		if _, ok := parentByID[currentID]; !ok {
			return nil
		}
	}
	return nil
}

// buildCategoryTree converts flat category rows into a sorted recursive tree.
func buildCategoryTree(list []*entitymodel.CmsCategory) []*CategoryItem {
	nodes := make(map[int64]*CategoryItem, len(list))
	roots := make([]*CategoryItem, 0)
	for _, category := range list {
		if category == nil {
			continue
		}
		nodes[category.Id] = &CategoryItem{CmsCategory: category}
	}
	for _, category := range list {
		if category == nil {
			continue
		}
		node := nodes[category.Id]
		if category.ParentId <= 0 {
			roots = append(roots, node)
			continue
		}
		parent := nodes[category.ParentId]
		if parent == nil {
			roots = append(roots, node)
			continue
		}
		parent.Children = append(parent.Children, node)
	}
	return roots
}
