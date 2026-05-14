// This file implements the CMS article deletion controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// ArticleDelete deletes a CMS article.
func (c *ControllerV1) ArticleDelete(ctx context.Context, req *v1.ArticleDeleteReq) (res *v1.ArticleDeleteRes, err error) {
	if err = c.cmsSvc.DeleteArticle(ctx, req.Id); err != nil {
		return nil, err
	}
	return &v1.ArticleDeleteRes{}, nil
}
