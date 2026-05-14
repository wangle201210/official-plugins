// This file implements the CMS category update controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// CategoryUpdate updates a CMS category.
func (c *ControllerV1) CategoryUpdate(ctx context.Context, req *v1.CategoryUpdateReq) (res *v1.CategoryUpdateRes, err error) {
	err = c.cmsSvc.UpdateCategory(ctx, cmssvc.CategorySaveInput{
		Id:              req.Id,
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
	return &v1.CategoryUpdateRes{}, nil
}
