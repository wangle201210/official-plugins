// This file implements the CMS carousel slide create controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// SlideCreate creates one CMS carousel slide.
func (c *ControllerV1) SlideCreate(ctx context.Context, req *v1.SlideCreateReq) (res *v1.SlideCreateRes, err error) {
	id, err := c.cmsSvc.CreateSlide(ctx, cmssvc.SlideSaveInput{
		GroupCode: req.GroupCode,
		Title:     req.Title,
		Subtitle:  req.Subtitle,
		Image:     req.Image,
		Link:      req.Link,
		Sort:      req.Sort,
		Status:    req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &v1.SlideCreateRes{Id: id}, nil
}
