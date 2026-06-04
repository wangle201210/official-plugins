// This file implements the CMS visitor message list controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// MessageList returns paged CMS visitor messages.
func (c *ControllerV1) MessageList(ctx context.Context, req *v1.MessageListReq) (res *v1.MessageListRes, err error) {
	out, err := c.cmsSvc.ListMessages(ctx, cmssvc.MessageListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
		Status:   req.Status,
		Keyword:  req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	return &v1.MessageListRes{List: toAPIMessages(out.List), Total: out.Total}, nil
}
