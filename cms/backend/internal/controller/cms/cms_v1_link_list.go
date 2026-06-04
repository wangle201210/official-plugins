// This file implements the CMS friendly link list controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// LinkList returns paged CMS friendly links for management.
func (c *ControllerV1) LinkList(ctx context.Context, req *v1.LinkListReq) (res *v1.LinkListRes, err error) {
	page, err := c.cmsSvc.ListLinks(ctx, cmssvc.LinkListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		GroupCode: req.GroupCode,
		Status:    req.Status,
		Keyword:   req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	return &v1.LinkListRes{
		List:  toAPILinks(page.List),
		Total: page.Total,
	}, nil
}
