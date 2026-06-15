// This file implements the CMS public product list controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// PublicProductList returns publicly visible CMS products.
func (c *ControllerV1) PublicProductList(ctx context.Context, req *v1.PublicProductListReq) (res *v1.PublicProductListRes, err error) {
	out, err := c.cmsSvc.ListPublicProducts(ctx, cmssvc.PublicProductListInput{
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		CategoryId: req.CategoryId,
	})
	if err != nil {
		return nil, err
	}
	return &v1.PublicProductListRes{List: toAPIProducts(out.List), Total: out.Total}, nil
}
