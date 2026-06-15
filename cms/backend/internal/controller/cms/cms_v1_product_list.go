// This file implements the CMS product list controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// ProductList returns paged CMS products.
func (c *ControllerV1) ProductList(ctx context.Context, req *v1.ProductListReq) (res *v1.ProductListRes, err error) {
	out, err := c.cmsSvc.ListProducts(ctx, cmssvc.ProductListInput{
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		CategoryId: req.CategoryId,
		Status:     req.Status,
		Name:       req.Name,
	})
	if err != nil {
		return nil, err
	}
	return &v1.ProductListRes{List: toAPIProducts(out.List), Total: out.Total}, nil
}
