// This file implements the CMS carousel slide update controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// SlideUpdate updates one CMS carousel slide.
func (c *ControllerV1) SlideUpdate(ctx context.Context, req *v1.SlideUpdateReq) (res *v1.SlideUpdateRes, err error) {
	if err := c.cmsSvc.UpdateSlide(ctx, cmssvc.SlideSaveInput{
		Id:        req.Id,
		GroupCode: req.GroupCode,
		Title:     req.Title,
		Subtitle:  req.Subtitle,
		Image:     req.Image,
		Link:      req.Link,
		Sort:      req.Sort,
		Status:    req.Status,
	}); err != nil {
		return nil, err
	}
	return &v1.SlideUpdateRes{}, nil
}
