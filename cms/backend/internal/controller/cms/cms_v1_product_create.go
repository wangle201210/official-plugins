// This file implements the CMS product creation controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// ProductCreate creates a CMS product.
func (c *ControllerV1) ProductCreate(ctx context.Context, req *v1.ProductCreateReq) (res *v1.ProductCreateRes, err error) {
	id, err := c.cmsSvc.CreateProduct(ctx, cmssvc.ProductSaveInput{
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
	return &v1.ProductCreateRes{Id: id}, nil
}
