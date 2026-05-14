// This file implements the CMS carousel slide list controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// SlideList returns paged CMS carousel slides for management.
func (c *ControllerV1) SlideList(ctx context.Context, req *v1.SlideListReq) (res *v1.SlideListRes, err error) {
	page, err := c.cmsSvc.ListSlides(ctx, cmssvc.SlideListInput{
		PageNum:   req.PageNum,
		PageSize:  req.PageSize,
		GroupCode: req.GroupCode,
		Status:    req.Status,
		Keyword:   req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SlideListRes{
		List:  toAPISlides(page.List),
		Total: page.Total,
	}, nil
}
