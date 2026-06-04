// This file implements the CMS friendly link create controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// LinkCreate creates one CMS friendly link.
func (c *ControllerV1) LinkCreate(ctx context.Context, req *v1.LinkCreateReq) (res *v1.LinkCreateRes, err error) {
	id, err := c.cmsSvc.CreateLink(ctx, cmssvc.LinkSaveInput{
		GroupCode: req.GroupCode,
		Name:      req.Name,
		Url:       req.Url,
		Logo:      req.Logo,
		Sort:      req.Sort,
		Status:    req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &v1.LinkCreateRes{Id: id}, nil
}
