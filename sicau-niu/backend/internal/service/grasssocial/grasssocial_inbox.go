// grasssocial_inbox.go implements the player inbox (paged list and mark-read) and
// the small store helpers shared by the steal and gift actions: the current
// balance read, the recipient-existence check and the inbox-message insert. The
// inbox is isolated to the authenticated player; mark-read verifies ownership
// before updating so a player can never operate on another player's message.

package grasssocial

import (
	"context"

	"lina-core/pkg/apitime"
	"lina-core/pkg/bizerr"
	"lina-plugin-sicau-niu/backend/internal/dao"
	"lina-plugin-sicau-niu/backend/internal/model/do"
	entitymodel "lina-plugin-sicau-niu/backend/internal/model/entity"
)

// Inbox paging defaults bound the player inbox page size.
const (
	defaultInboxPageNum  = 1
	defaultInboxPageSize = 10
	maxInboxPageSize     = 100
)

// readFlag values mark a message read state in the inbox_msg is_read column.
const (
	inboxUnread = 0
	inboxRead   = 1
)

// MessagesInput defines the player inbox list query.
type MessagesInput struct {
	// PageNum is the requested page number; defaults to 1 when non-positive.
	PageNum int
	// PageSize is the requested page size; defaults to 10 and is capped at 100.
	PageSize int
}

// MessagesOutput defines the player inbox list result.
type MessagesOutput struct {
	// List holds the current page of messages, newest first.
	List []*MessageView
	// Total is the player's total message count.
	Total int
}

// MessageView is one inbox message projected for the player.
type MessageView struct {
	// Id is the message ID.
	Id int64
	// MsgType is the message type string.
	MsgType string
	// Content is the message content text.
	Content string
	// IsRead reports whether the message is read.
	IsRead bool
	// CreatedAt is the message time as a Unix timestamp in milliseconds.
	CreatedAt *int64
}

// Messages returns the player's own inbox page with the total count.
func (s *serviceImpl) Messages(ctx context.Context, playerID int64, in *MessagesInput) (*MessagesOutput, error) {
	if playerID <= 0 {
		return nil, bizerr.NewCode(CodeQueryFailed)
	}
	pageNum, pageSize := normalizeInboxPagination(in)

	model := dao.InboxMsg.Ctx(ctx).Where(dao.InboxMsg.Columns().UserId, playerID)
	total, err := model.Count()
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}

	rows := make([]*entitymodel.InboxMsg, 0, pageSize)
	err = model.
		OrderDesc(dao.InboxMsg.Columns().Id).
		Page(pageNum, pageSize).
		Scan(&rows)
	if err != nil {
		return nil, bizerr.WrapCode(err, CodeQueryFailed)
	}

	list := make([]*MessageView, 0, len(rows))
	for _, row := range rows {
		list = append(list, &MessageView{
			Id:        row.Id,
			MsgType:   row.MsgType,
			Content:   row.Content,
			IsRead:    row.IsRead == inboxRead,
			CreatedAt: apitime.Milli(row.CreatedAt),
		})
	}
	return &MessagesOutput{List: list, Total: total}, nil
}

// MarkRead marks one of the player's own messages read after verifying ownership.
func (s *serviceImpl) MarkRead(ctx context.Context, playerID int64, messageID int64) error {
	if playerID <= 0 || messageID <= 0 {
		return bizerr.NewCode(CodeMessageNotFound)
	}

	owned, err := dao.InboxMsg.Ctx(ctx).
		Where(dao.InboxMsg.Columns().Id, messageID).
		Where(dao.InboxMsg.Columns().UserId, playerID).
		Count()
	if err != nil {
		return bizerr.WrapCode(err, CodeQueryFailed)
	}
	if owned == 0 {
		return bizerr.NewCode(CodeMessageNotFound)
	}

	if _, err = dao.InboxMsg.Ctx(ctx).
		Where(dao.InboxMsg.Columns().Id, messageID).
		Where(dao.InboxMsg.Columns().UserId, playerID).
		Data(do.InboxMsg{IsRead: inboxRead}).
		Update(); err != nil {
		return bizerr.WrapCode(err, CodeWriteFailed)
	}
	return nil
}

// currentBalance reads userID's current grass balance inside the caller's
// transaction, reading only the balance column. A missing account reports zero.
func (s *serviceImpl) currentBalance(ctx context.Context, userID int64) (int64, error) {
	balance, err := dao.GrassAccount.Ctx(ctx).
		Fields(dao.GrassAccount.Columns().Balance).
		Where(dao.GrassAccount.Columns().UserId, userID).
		Value()
	if err != nil {
		return 0, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return balance.Int64(), nil
}

// userExists reports whether an active player with id exists, used to validate a
// gift recipient before moving grass.
func (s *serviceImpl) userExists(ctx context.Context, id int64) (bool, error) {
	count, err := dao.User.Ctx(ctx).Where(dao.User.Columns().Id, id).Count()
	if err != nil {
		return false, bizerr.WrapCode(err, CodeQueryFailed)
	}
	return count > 0, nil
}

// insertInboxMessage appends one notification message to userID's inbox inside
// the caller's transaction so the message and the grass movement commit together.
func (s *serviceImpl) insertInboxMessage(ctx context.Context, userID int64, msgType MsgType, content string) error {
	if _, err := dao.InboxMsg.Ctx(ctx).Data(do.InboxMsg{
		UserId:  userID,
		MsgType: msgType.String(),
		Content: content,
		IsRead:  inboxUnread,
	}).Insert(); err != nil {
		return bizerr.WrapCode(err, CodeWriteFailed)
	}
	return nil
}

// normalizeInboxPagination applies the inbox paging defaults and max page-size cap.
func normalizeInboxPagination(in *MessagesInput) (int, int) {
	if in == nil {
		return defaultInboxPageNum, defaultInboxPageSize
	}
	pageNum := in.PageNum
	if pageNum <= 0 {
		pageNum = defaultInboxPageNum
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = defaultInboxPageSize
	}
	if pageSize > maxInboxPageSize {
		pageSize = maxInboxPageSize
	}
	return pageNum, pageSize
}
