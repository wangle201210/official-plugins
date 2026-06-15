// This file implements the CMS product detail controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// ProductGet returns one CMS product.
func (c *ControllerV1) ProductGet(ctx context.Context, req *v1.ProductGetReq) (res *v1.ProductGetRes, err error) {
	item, err := c.cmsSvc.GetProduct(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.ProductGetRes{ProductItem: toAPIProduct(item)}, nil
}
