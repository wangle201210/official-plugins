// This file implements the CMS friendly link update controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// LinkUpdate updates one CMS friendly link.
func (c *ControllerV1) LinkUpdate(ctx context.Context, req *v1.LinkUpdateReq) (res *v1.LinkUpdateRes, err error) {
	if err := c.cmsSvc.UpdateLink(ctx, cmssvc.LinkSaveInput{
		Id:        req.Id,
		GroupCode: req.GroupCode,
		Name:      req.Name,
		Url:       req.Url,
		Logo:      req.Logo,
		Sort:      req.Sort,
		Status:    req.Status,
	}); err != nil {
		return nil, err
	}
	return &v1.LinkUpdateRes{}, nil
}
