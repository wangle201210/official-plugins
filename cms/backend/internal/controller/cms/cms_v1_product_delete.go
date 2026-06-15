// This file implements the CMS product delete controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// ProductDelete deletes one CMS product.
func (c *ControllerV1) ProductDelete(ctx context.Context, req *v1.ProductDeleteReq) (res *v1.ProductDeleteRes, err error) {
	if err = c.cmsSvc.DeleteProduct(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.ProductDeleteRes{}, nil
}
