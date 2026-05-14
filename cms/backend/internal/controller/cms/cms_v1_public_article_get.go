// This file implements the public CMS article detail controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
)

// PublicArticleGet returns one published CMS article by slug.
func (c *ControllerV1) PublicArticleGet(ctx context.Context, req *v1.PublicArticleGetReq) (res *v1.PublicArticleGetRes, err error) {
	item, err := c.cmsSvc.GetPublicArticleBySlug(ctx, req.Slug)
	if err != nil {
		return nil, err
	}
	return &v1.PublicArticleGetRes{ArticleItem: toAPIArticle(item)}, nil
}
