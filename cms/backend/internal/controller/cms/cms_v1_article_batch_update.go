// This file implements the CMS article batch status update controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// ArticleBatchUpdate publishes or unpublishes multiple CMS articles.
func (c *ControllerV1) ArticleBatchUpdate(ctx context.Context, req *v1.ArticleBatchUpdateReq) (res *v1.ArticleBatchUpdateRes, err error) {
	if err = c.cmsSvc.BatchUpdateArticleStatus(ctx, cmssvc.ArticleBatchStatusInput{Ids: req.Ids, Status: req.Status}); err != nil {
		return nil, err
	}
	return &v1.ArticleBatchUpdateRes{}, nil
}
