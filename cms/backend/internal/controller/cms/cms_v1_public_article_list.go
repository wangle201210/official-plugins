// This file implements the public CMS article list controller.

package cms

import (
	"context"

	"lina-plugin-cms/backend/api/cms/v1"
	cmssvc "lina-plugin-cms/backend/internal/service/cms"
)

// PublicArticleList returns published CMS articles.
func (c *ControllerV1) PublicArticleList(ctx context.Context, req *v1.PublicArticleListReq) (res *v1.PublicArticleListRes, err error) {
	out, err := c.cmsSvc.ListPublicArticles(ctx, cmssvc.PublicArticleListInput{
		PageNum:    req.PageNum,
		PageSize:   req.PageSize,
		CategoryId: req.CategoryId,
		Keyword:    req.Keyword,
	})
	if err != nil {
		return nil, err
	}
	return &v1.PublicArticleListRes{List: toAPIArticles(out.List), Total: out.Total}, nil
}
