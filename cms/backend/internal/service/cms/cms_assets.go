// This file implements CMS friendly link and carousel slide management.

package cms

import (
	"context"
	"fmt"
	"lina-core/pkg/bizerr"
	"lina-plugin-cms/backend/internal/dao"
	"lina-plugin-cms/backend/internal/model/do"
	"strings"
)

// ListLinks returns paged friendly links with management filters.
func (s *serviceImpl) ListLinks(ctx context.Context, in LinkListInput) (*LinkListOutput, error) {
	columns := dao.CmsLink.Columns()
	model := dao.CmsLink.Ctx(ctx)
	if in.GroupCode != "" {
		model = model.Where(columns.GroupCode, strings.TrimSpace(in.GroupCode))
	}
	if in.Status != nil {
		model = model.Where(columns.Status, *in.Status)
	}
	if keywordText := strings.TrimSpace(in.Keyword); keywordText != "" {
		keyword := "%" + keywordText + "%"
		model = model.Where(fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", columns.Name, columns.Url), keyword, keyword)
	}
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	list := make([]*LinkItem, 0)
	if err = model.Page(normalizePageNum(in.PageNum), normalizePageSize(in.PageSize)).OrderAsc(columns.Sort).OrderAsc(columns.Id).Scan(&list); err != nil {
		return nil, err
	}
	return &LinkListOutput{List: list, Total: total}, nil
}

// CreateLink creates one friendly link.
func (s *serviceImpl) CreateLink(ctx context.Context, in LinkSaveInput) (int64, error) {
	userID := s.currentUserID(ctx)
	return dao.CmsLink.Ctx(ctx).Data(do.CmsLink{GroupCode: strings.TrimSpace(in.GroupCode), Name: strings.TrimSpace(in.Name), Url: strings.TrimSpace(in.Url), Logo: strings.TrimSpace(in.Logo), Sort: in.Sort, Status: in.Status, CreatedBy: userID, UpdatedBy: userID}).InsertAndGetId()
}

// UpdateLink updates one friendly link after existence validation.
func (s *serviceImpl) UpdateLink(ctx context.Context, in LinkSaveInput) error {
	columns := dao.CmsLink.Columns()
	if err := s.ensureLinkExists(ctx, in.Id); err != nil {
		return err
	}
	_, err := dao.CmsLink.Ctx(ctx).Where(columns.Id, in.Id).Data(do.CmsLink{GroupCode: strings.TrimSpace(in.GroupCode), Name: strings.TrimSpace(in.Name), Url: strings.TrimSpace(in.Url), Logo: strings.TrimSpace(in.Logo), Sort: in.Sort, Status: in.Status, UpdatedBy: s.currentUserID(ctx)}).Update()
	return err
}

// DeleteLink removes one friendly link.
func (s *serviceImpl) DeleteLink(ctx context.Context, id int64) error {
	columns := dao.CmsLink.Columns()
	if err := s.ensureLinkExists(ctx, id); err != nil {
		return err
	}
	_, err := dao.CmsLink.Ctx(ctx).Where(columns.Id, id).Delete()
	return err
}

// ListSlides returns paged slides with management filters.
func (s *serviceImpl) ListSlides(ctx context.Context, in SlideListInput) (*SlideListOutput, error) {
	columns := dao.CmsSlide.Columns()
	model := dao.CmsSlide.Ctx(ctx)
	if in.GroupCode != "" {
		model = model.Where(columns.GroupCode, strings.TrimSpace(in.GroupCode))
	}
	if in.Status != nil {
		model = model.Where(columns.Status, *in.Status)
	}
	if keywordText := strings.TrimSpace(in.Keyword); keywordText != "" {
		keyword := "%" + keywordText + "%"
		model = model.Where(fmt.Sprintf("(%s LIKE ? OR %s LIKE ?)", columns.Title, columns.Subtitle), keyword, keyword)
	}
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	list := make([]*SlideItem, 0)
	if err = model.Page(normalizePageNum(in.PageNum), normalizePageSize(in.PageSize)).OrderAsc(columns.Sort).OrderAsc(columns.Id).Scan(&list); err != nil {
		return nil, err
	}
	return &SlideListOutput{List: list, Total: total}, nil
}

// CreateSlide creates one slide.
func (s *serviceImpl) CreateSlide(ctx context.Context, in SlideSaveInput) (int64, error) {
	userID := s.currentUserID(ctx)
	return dao.CmsSlide.Ctx(ctx).Data(do.CmsSlide{GroupCode: strings.TrimSpace(in.GroupCode), Title: strings.TrimSpace(in.Title), Subtitle: strings.TrimSpace(in.Subtitle), Image: strings.TrimSpace(in.Image), Link: strings.TrimSpace(in.Link), Sort: in.Sort, Status: in.Status, CreatedBy: userID, UpdatedBy: userID}).InsertAndGetId()
}

// UpdateSlide updates one slide after existence validation.
func (s *serviceImpl) UpdateSlide(ctx context.Context, in SlideSaveInput) error {
	columns := dao.CmsSlide.Columns()
	if err := s.ensureSlideExists(ctx, in.Id); err != nil {
		return err
	}
	_, err := dao.CmsSlide.Ctx(ctx).Where(columns.Id, in.Id).Data(do.CmsSlide{GroupCode: strings.TrimSpace(in.GroupCode), Title: strings.TrimSpace(in.Title), Subtitle: strings.TrimSpace(in.Subtitle), Image: strings.TrimSpace(in.Image), Link: strings.TrimSpace(in.Link), Sort: in.Sort, Status: in.Status, UpdatedBy: s.currentUserID(ctx)}).Update()
	return err
}

// DeleteSlide removes one slide.
func (s *serviceImpl) DeleteSlide(ctx context.Context, id int64) error {
	columns := dao.CmsSlide.Columns()
	if err := s.ensureSlideExists(ctx, id); err != nil {
		return err
	}
	_, err := dao.CmsSlide.Ctx(ctx).Where(columns.Id, id).Delete()
	return err
}

// ListPublicLinks returns enabled links for public site rendering.
func (s *serviceImpl) ListPublicLinks(ctx context.Context) ([]*LinkItem, error) {
	columns := dao.CmsLink.Columns()
	list := make([]*LinkItem, 0)
	err := dao.CmsLink.Ctx(ctx).Where(columns.Status, StatusEnabled).OrderAsc(columns.Sort).OrderAsc(columns.Id).Scan(&list)
	return list, err
}

// ListPublicSlides returns enabled slides for public site rendering.
func (s *serviceImpl) ListPublicSlides(ctx context.Context) ([]*SlideItem, error) {
	columns := dao.CmsSlide.Columns()
	list := make([]*SlideItem, 0)
	err := dao.CmsSlide.Ctx(ctx).Where(columns.Status, StatusEnabled).OrderAsc(columns.Sort).OrderAsc(columns.Id).Scan(&list)
	return list, err
}

// ensureLinkExists verifies that a friendly link exists before mutation.
func (s *serviceImpl) ensureLinkExists(ctx context.Context, id int64) error {
	columns := dao.CmsLink.Columns()
	count, err := dao.CmsLink.Ctx(ctx).Where(columns.Id, id).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return bizerr.NewCode(CodeLinkNotFound)
	}
	return nil
}

// ensureSlideExists verifies that a slide exists before mutation.
func (s *serviceImpl) ensureSlideExists(ctx context.Context, id int64) error {
	columns := dao.CmsSlide.Columns()
	count, err := dao.CmsSlide.Ctx(ctx).Where(columns.Id, id).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return bizerr.NewCode(CodeSlideNotFound)
	}
	return nil
}
