// This file implements the CMS carousel slide delete controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// SlideDelete deletes one CMS carousel slide.
func (c *ControllerV1) SlideDelete(ctx context.Context, req *v1.SlideDeleteReq) (res *v1.SlideDeleteRes, err error) {
	if err := c.cmsSvc.DeleteSlide(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.SlideDeleteRes{}, nil
}
