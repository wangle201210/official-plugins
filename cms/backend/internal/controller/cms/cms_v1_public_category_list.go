// This file implements the public CMS category tree controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// PublicCategoryList returns enabled public CMS categories.
func (c *ControllerV1) PublicCategoryList(ctx context.Context, _ *v1.PublicCategoryListReq) (res *v1.PublicCategoryListRes, err error) {
	list, err := c.cmsSvc.ListCategories(ctx, cmssvc.CategoryListInput{PublicOnly: true})
	if err != nil {
		return nil, err
	}
	return &v1.PublicCategoryListRes{List: toAPICategories(list)}, nil
}
