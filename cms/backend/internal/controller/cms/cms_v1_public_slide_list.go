// This file implements the public CMS slide list controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// PublicSlideList returns enabled public slides.
func (c *ControllerV1) PublicSlideList(ctx context.Context, _ *v1.PublicSlideListReq) (res *v1.PublicSlideListRes, err error) {
	list, err := c.cmsSvc.ListPublicSlides(ctx)
	if err != nil {
		return nil, err
	}
	return &v1.PublicSlideListRes{List: toAPISlides(list)}, nil
}
