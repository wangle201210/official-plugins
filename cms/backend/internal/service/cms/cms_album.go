// This file implements CMS album management, whole-album image replacement,
// and public album reads. Image counts are assembled with one grouped
// aggregate per page and detail images with one query per album.

package cms

import (
	"context"

	"github.com/gogf/gf/v2/database/gdb"

	"lina-core/pkg/bizerr"
	"lina-plugin-cms/backend/internal/dao"
	"lina-plugin-cms/backend/internal/model/do"
	entitymodel "lina-plugin-cms/backend/internal/model/entity"
)

// AlbumImageMax caps the number of images per album.
const AlbumImageMax = 100

// ListAlbums returns paged management albums with category names and image counts.
func (s *serviceImpl) ListAlbums(ctx context.Context, in AlbumListInput) (*AlbumListOutput, error) {
	columns := dao.CmsAlbum.Columns()
	model := dao.CmsAlbum.Ctx(ctx)
	if in.CategoryId > 0 {
		model = model.Where(columns.CategoryId, in.CategoryId)
	}
	if in.Status != nil {
		model = model.Where(columns.Status, *in.Status)
	}
	if in.Name != "" {
		model = model.WhereLike(columns.Name, "%"+in.Name+"%")
	}
	model = model.OrderAsc(columns.Sort).OrderDesc(columns.Id)
	return s.scanAlbumPage(ctx, model, in.PageNum, in.PageSize)
}

// GetAlbum returns one management album with its ordered images.
func (s *serviceImpl) GetAlbum(ctx context.Context, id int64) (*AlbumItem, error) {
	var album *entitymodel.CmsAlbum
	if err := dao.CmsAlbum.Ctx(ctx).Where(dao.CmsAlbum.Columns().Id, id).Scan(&album); err != nil {
		return nil, err
	}
	if album == nil {
		return nil, bizerr.NewCode(CodeAlbumNotFound)
	}
	return s.wrapAlbumDetail(ctx, album)
}

// CreateAlbum creates one CMS album together with its full image list.
func (s *serviceImpl) CreateAlbum(ctx context.Context, in AlbumSaveInput) (int64, error) {
	if err := s.ensureCategoryExists(ctx, in.CategoryId); err != nil {
		return 0, err
	}
	userID := s.currentUserID(ctx)
	var albumID int64
	err := dao.CmsAlbum.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		id, err := dao.CmsAlbum.Ctx(ctx).Data(do.CmsAlbum{CategoryId: in.CategoryId, Name: in.Name, Cover: in.Cover, Description: in.Description, Sort: in.Sort, Status: in.Status, CreatedBy: userID, UpdatedBy: userID}).InsertAndGetId()
		if err != nil {
			return err
		}
		albumID = id
		return replaceAlbumImages(ctx, id, in.Images)
	})
	return albumID, err
}

// UpdateAlbum updates one CMS album and replaces its full image list.
func (s *serviceImpl) UpdateAlbum(ctx context.Context, in AlbumSaveInput) error {
	columns := dao.CmsAlbum.Columns()
	if _, err := s.GetAlbum(ctx, in.Id); err != nil {
		return err
	}
	if err := s.ensureCategoryExists(ctx, in.CategoryId); err != nil {
		return err
	}
	return dao.CmsAlbum.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, err := dao.CmsAlbum.Ctx(ctx).Where(columns.Id, in.Id).Data(do.CmsAlbum{CategoryId: in.CategoryId, Name: in.Name, Cover: in.Cover, Description: in.Description, Sort: in.Sort, Status: in.Status, UpdatedBy: s.currentUserID(ctx)}).Update(); err != nil {
			return err
		}
		return replaceAlbumImages(ctx, in.Id, in.Images)
	})
}

// DeleteAlbum removes one CMS album together with all of its image rows.
func (s *serviceImpl) DeleteAlbum(ctx context.Context, id int64) error {
	if _, err := s.GetAlbum(ctx, id); err != nil {
		return err
	}
	return dao.CmsAlbum.Transaction(ctx, func(ctx context.Context, _ gdb.TX) error {
		if _, err := dao.CmsAlbumImage.Ctx(ctx).Where(dao.CmsAlbumImage.Columns().AlbumId, id).Delete(); err != nil {
			return err
		}
		_, err := dao.CmsAlbum.Ctx(ctx).Where(dao.CmsAlbum.Columns().Id, id).Delete()
		return err
	})
}

// ListPublicAlbums returns enabled albums under enabled categories with image counts.
func (s *serviceImpl) ListPublicAlbums(ctx context.Context, in PublicAlbumListInput) (*AlbumListOutput, error) {
	columns := dao.CmsAlbum.Columns()
	model := s.applyPublicAlbumVisibility(ctx, dao.CmsAlbum.Ctx(ctx))
	if in.CategoryId > 0 {
		model = model.Where(columns.CategoryId, in.CategoryId)
	}
	model = model.OrderAsc(columns.Sort).OrderDesc(columns.Id)
	return s.scanAlbumPage(ctx, model, in.PageNum, in.PageSize)
}

// GetPublicAlbum returns one enabled album with its ordered images.
func (s *serviceImpl) GetPublicAlbum(ctx context.Context, id int64) (*AlbumItem, error) {
	var album *entitymodel.CmsAlbum
	err := s.applyPublicAlbumVisibility(ctx, dao.CmsAlbum.Ctx(ctx)).Where(dao.CmsAlbum.Columns().Id, id).Scan(&album)
	if err != nil {
		return nil, err
	}
	if album == nil {
		return nil, bizerr.NewCode(CodePublicContentNotFound)
	}
	return s.wrapAlbumDetail(ctx, album)
}

// applyPublicAlbumVisibility filters enabled albums under enabled categories.
func (s *serviceImpl) applyPublicAlbumVisibility(ctx context.Context, model *gdb.Model) *gdb.Model {
	albumColumns := dao.CmsAlbum.Columns()
	categoryColumns := dao.CmsCategory.Columns()
	enabledCategorySubQuery := dao.CmsCategory.Ctx(ctx).Fields(categoryColumns.Id).Where(categoryColumns.Status, StatusEnabled)
	return model.
		Where(albumColumns.Status, StatusEnabled).
		Where(albumColumns.CategoryId+" IN (?)", enabledCategorySubQuery)
}

// replaceAlbumImages replaces the full image set of one album inside the
// caller's transaction: one delete plus one batched insert.
func replaceAlbumImages(ctx context.Context, albumID int64, images []AlbumImageInput) error {
	if len(images) > AlbumImageMax {
		return bizerr.NewCode(CodeAlbumImageLimitExceeded)
	}
	if _, err := dao.CmsAlbumImage.Ctx(ctx).Where(dao.CmsAlbumImage.Columns().AlbumId, albumID).Delete(); err != nil {
		return err
	}
	if len(images) == 0 {
		return nil
	}
	rows := make([]do.CmsAlbumImage, 0, len(images))
	for _, image := range images {
		rows = append(rows, do.CmsAlbumImage{AlbumId: albumID, Url: image.Url, Title: image.Title, Sort: image.Sort})
	}
	_, err := dao.CmsAlbumImage.Ctx(ctx).Data(rows).Insert()
	return err
}

// scanAlbumPage scans a paged album query and assembles category names and
// image counts with batched lookups.
func (s *serviceImpl) scanAlbumPage(ctx context.Context, model *gdb.Model, pageNum int, pageSize int) (*AlbumListOutput, error) {
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	list := make([]*entitymodel.CmsAlbum, 0)
	if err = model.Page(normalizePageNum(pageNum), normalizePageSize(pageSize)).Scan(&list); err != nil {
		return nil, err
	}
	categoryIDs := make([]int64, 0, len(list))
	albumIDs := make([]int64, 0, len(list))
	seenCategories := make(map[int64]bool, len(list))
	for _, album := range list {
		albumIDs = append(albumIDs, album.Id)
		if album.CategoryId > 0 && !seenCategories[album.CategoryId] {
			categoryIDs = append(categoryIDs, album.CategoryId)
			seenCategories[album.CategoryId] = true
		}
	}
	categoryNames, err := s.categoryNameMapByIDs(ctx, categoryIDs)
	if err != nil {
		return nil, err
	}
	imageCounts, err := albumImageCountMap(ctx, albumIDs)
	if err != nil {
		return nil, err
	}
	items := make([]*AlbumItem, 0, len(list))
	for _, album := range list {
		items = append(items, &AlbumItem{CmsAlbum: album, CategoryName: categoryNames[album.CategoryId], ImageCount: imageCounts[album.Id]})
	}
	return &AlbumListOutput{List: items, Total: total}, nil
}

// wrapAlbumDetail loads the category name, image count, and ordered images of one album.
func (s *serviceImpl) wrapAlbumDetail(ctx context.Context, album *entitymodel.CmsAlbum) (*AlbumItem, error) {
	categoryNames, err := s.categoryNameMapByIDs(ctx, []int64{album.CategoryId})
	if err != nil {
		return nil, err
	}
	imageColumns := dao.CmsAlbumImage.Columns()
	images := make([]*entitymodel.CmsAlbumImage, 0)
	if err = dao.CmsAlbumImage.Ctx(ctx).Where(imageColumns.AlbumId, album.Id).OrderAsc(imageColumns.Sort).OrderAsc(imageColumns.Id).Scan(&images); err != nil {
		return nil, err
	}
	return &AlbumItem{CmsAlbum: album, CategoryName: categoryNames[album.CategoryId], ImageCount: len(images), Images: images}, nil
}

// albumImageCountMap loads image counts for the supplied album IDs with one
// grouped aggregate query.
func albumImageCountMap(ctx context.Context, albumIDs []int64) (map[int64]int, error) {
	result := make(map[int64]int, len(albumIDs))
	if len(albumIDs) == 0 {
		return result, nil
	}
	columns := dao.CmsAlbumImage.Columns()
	records, err := dao.CmsAlbumImage.Ctx(ctx).
		Fields(columns.AlbumId, "COUNT(*) AS image_count").
		WhereIn(columns.AlbumId, albumIDs).
		Group(columns.AlbumId).
		All()
	if err != nil {
		return nil, err
	}
	for _, record := range records {
		result[record[columns.AlbumId].Int64()] = record["image_count"].Int()
	}
	return result, nil
}
