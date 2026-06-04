// This file implements the CMS category tree controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// CategoryList returns CMS category tree.
func (c *ControllerV1) CategoryList(ctx context.Context, req *v1.CategoryListReq) (res *v1.CategoryListRes, err error) {
	list, err := c.cmsSvc.ListCategories(ctx, cmssvc.CategoryListInput{Status: req.Status})
	if err != nil {
		return nil, err
	}
	return &v1.CategoryListRes{List: toAPICategories(list)}, nil
}
