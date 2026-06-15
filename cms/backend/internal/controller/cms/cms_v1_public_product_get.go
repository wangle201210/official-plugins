// This file implements the CMS public product detail controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// PublicProductGet returns one publicly visible CMS product by slug.
func (c *ControllerV1) PublicProductGet(ctx context.Context, req *v1.PublicProductGetReq) (res *v1.PublicProductGetRes, err error) {
	item, err := c.cmsSvc.GetPublicProductBySlug(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	return &v1.PublicProductGetRes{ProductItem: toAPIProduct(item)}, nil
}
