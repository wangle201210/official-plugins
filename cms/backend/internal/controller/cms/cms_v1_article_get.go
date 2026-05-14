// This file implements the CMS article detail controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// ArticleGet returns CMS article details.
func (c *ControllerV1) ArticleGet(ctx context.Context, req *v1.ArticleGetReq) (res *v1.ArticleGetRes, err error) {
	item, err := c.cmsSvc.GetArticle(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return &v1.ArticleGetRes{ArticleItem: toAPIArticle(item)}, nil
}
