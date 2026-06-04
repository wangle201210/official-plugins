// This file implements the public CMS visitor message list controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// PublicMessageList returns approved CMS visitor messages for public display.
func (c *ControllerV1) PublicMessageList(ctx context.Context, req *v1.PublicMessageListReq) (res *v1.PublicMessageListRes, err error) {
	out, err := c.cmsSvc.ListPublicMessages(ctx, cmssvc.PublicMessageListInput{
		PageNum:  req.PageNum,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	return &v1.PublicMessageListRes{List: toAPIPublicMessages(out.List), Total: out.Total}, nil
}
