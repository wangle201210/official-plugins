// This file implements CMS visitor message management and public message creation.

package cms

import (
	"context"
	"fmt"
	"lina-core/pkg/bizerr"
	"lina-plugin-cms/backend/internal/dao"
	"lina-plugin-cms/backend/internal/model/do"
)

// ListMessages returns paged management visitor messages.
func (s *serviceImpl) ListMessages(ctx context.Context, in MessageListInput) (*MessageListOutput, error) {
	columns := dao.CmsMessage.Columns()
	model := dao.CmsMessage.Ctx(ctx)
	if in.Status != nil {
		model = model.Where(columns.Status, *in.Status)
	}
	if in.Keyword != "" {
		keyword := "%" + in.Keyword + "%"
		model = model.Where(fmt.Sprintf("(%s LIKE ? OR %s LIKE ? OR %s LIKE ? OR %s LIKE ?)", columns.Name, columns.Email, columns.Mobile, columns.Content), keyword, keyword, keyword, keyword)
	}
	total, err := model.Count()
	if err != nil {
		return nil, err
	}
	list := make([]*MessageItem, 0)
	if err = model.Page(normalizePageNum(in.PageNum), normalizePageSize(in.PageSize)).OrderDesc(columns.Id).Scan(&list); err != nil {
		return nil, err
	}
	return &MessageListOutput{List: list, Total: total}, nil
}

// ListPublicMessages returns approved public messages when the site switch allows display.
func (s *serviceImpl) ListPublicMessages(ctx context.Context, in PublicMessageListInput) (*MessageListOutput, error) {
	site, err := s.GetSite(ctx, true)
	if err != nil {
		return nil, err
	}
	if site.ShowMessages != StatusEnabled {
		return &MessageListOutput{List: []*MessageItem{}, Total: 0}, nil
	}
	approvedStatus := MessageStatusApproved
	return s.ListMessages(ctx, MessageListInput{PageNum: in.PageNum, PageSize: in.PageSize, Status: &approvedStatus})
}

// UpdateMessage updates review status and reply text for one visitor message.
func (s *serviceImpl) UpdateMessage(ctx context.Context, in MessageUpdateInput) error {
	columns := dao.CmsMessage.Columns()
	if err := s.ensureMessageExists(ctx, in.Id); err != nil {
		return err
	}
	_, err := dao.CmsMessage.Ctx(ctx).Where(columns.Id, in.Id).Data(do.CmsMessage{Status: in.Status, Reply: in.Reply, UpdatedBy: s.currentUserID(ctx)}).Update()
	return err
}

// DeleteMessage removes one visitor message.
func (s *serviceImpl) DeleteMessage(ctx context.Context, id int64) error {
	columns := dao.CmsMessage.Columns()
	if err := s.ensureMessageExists(ctx, id); err != nil {
		return err
	}
	_, err := dao.CmsMessage.Ctx(ctx).Where(columns.Id, id).Delete()
	return err
}

// CreatePublicMessage stores one public visitor message submission for review.
func (s *serviceImpl) CreatePublicMessage(ctx context.Context, in PublicMessageCreateInput) (int64, error) {
	return dao.CmsMessage.Ctx(ctx).Data(do.CmsMessage{Name: in.Name, Mobile: in.Mobile, Email: in.Email, Content: in.Content, Status: MessageStatusPending, UserIp: in.UserIp, UserAgent: in.UserAgent}).InsertAndGetId()
}

// ensureMessageExists verifies that a visitor message exists before mutation.
func (s *serviceImpl) ensureMessageExists(ctx context.Context, id int64) error {
	columns := dao.CmsMessage.Columns()
	count, err := dao.CmsMessage.Ctx(ctx).Where(columns.Id, id).Count()
	if err != nil {
		return err
	}
	if count == 0 {
		return bizerr.NewCode(CodeMessageNotFound)
	}
	return nil
}
