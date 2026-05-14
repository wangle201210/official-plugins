// This file implements the CMS category deletion controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// CategoryDelete deletes a CMS category.
func (c *ControllerV1) CategoryDelete(ctx context.Context, req *v1.CategoryDeleteReq) (res *v1.CategoryDeleteRes, err error) {
	if err = c.cmsSvc.DeleteCategory(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.CategoryDeleteRes{}, nil
}
