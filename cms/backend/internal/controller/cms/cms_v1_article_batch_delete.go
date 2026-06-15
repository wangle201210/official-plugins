// This file implements the CMS article batch delete controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// ArticleBatchDelete soft deletes multiple CMS articles.
func (c *ControllerV1) ArticleBatchDelete(ctx context.Context, req *v1.ArticleBatchDeleteReq) (res *v1.ArticleBatchDeleteRes, err error) {
	if err = c.cmsSvc.BatchDeleteArticles(ctx, req.Ids); err != nil {
		return nil, err
	}
	return &v1.ArticleBatchDeleteRes{}, nil
}
