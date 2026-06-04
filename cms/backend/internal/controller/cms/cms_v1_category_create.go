// This file implements the CMS category creation controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// CategoryCreate creates a CMS category.
func (c *ControllerV1) CategoryCreate(ctx context.Context, req *v1.CategoryCreateReq) (res *v1.CategoryCreateRes, err error) {
	id, err := c.cmsSvc.CreateCategory(ctx, cmssvc.CategorySaveInput{
		ParentId:        req.ParentId,
		Code:            req.Code,
		Name:            req.Name,
		Type:            req.Type,
		Path:            req.Path,
		ListTemplate:    req.ListTemplate,
		ContentTemplate: req.ContentTemplate,
		Cover:           req.Cover,
		Outlink:         req.Outlink,
		Title:           req.Title,
		Keywords:        req.Keywords,
		Description:     req.Description,
		Sort:            req.Sort,
		Status:          req.Status,
	})
	if err != nil {
		return nil, err
	}
	return &v1.CategoryCreateRes{Id: id}, nil
}
