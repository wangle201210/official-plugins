// This file implements the CMS product update controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// ProductUpdate updates a CMS product.
func (c *ControllerV1) ProductUpdate(ctx context.Context, req *v1.ProductUpdateReq) (res *v1.ProductUpdateRes, err error) {
	err = c.cmsSvc.UpdateProduct(ctx, cmssvc.ProductSaveInput{
		Id:          req.Id,
		CategoryId:  req.CategoryId,
		Name:        req.Name,
		Slug:        req.Slug,
		Summary:     req.Summary,
		Cover:       req.Cover,
		Gallery:     req.Gallery,
		Price:       req.Price,
		Spec:        req.Spec,
		Content:     req.Content,
		Keywords:    req.Keywords,
		Description: req.Description,
		Sort:        req.Sort,
		Status:      req.Status,
		IsTop:       req.IsTop,
		IsRecommend: req.IsRecommend,
		PublishedAt: req.PublishedAt,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ProductUpdateRes{}, nil
}
